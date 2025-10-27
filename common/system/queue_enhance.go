package system

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/reguluswee/walletus/common/log"
)

type processHandler[T comparable] func(slot T, lock *sync.WaitGroup)

var signInLocks sync.Map // mutex lock for hash or block number

type RichQueue[T comparable] struct {
	sync.Mutex
	slots  []T // here might be hash or block number
	notify chan struct{}
}

func NewRichQueue[T comparable]() *RichQueue[T] {
	return &RichQueue[T]{
		slots:  make([]T, 0),
		notify: make(chan struct{}, 1),
	}
}

func (q *RichQueue[T]) Enqueue(item T) {
	_, loaded := signInLocks.LoadOrStore(item, struct{}{})
	if loaded {
		return
	}

	q.Lock()
	defer q.Unlock()

	q.slots = append(q.slots, item)
	select {
	case q.notify <- struct{}{}:
	default:
	}
}

func (q *RichQueue[T]) EnqueueWithDelay(item T, delay time.Duration) {
	signInLocks.Delete(item)

	// 延迟后重新入队
	go func() {
		time.Sleep(delay)
		q.Enqueue(item)
	}()
}

func (q *RichQueue[T]) BatchEnqueue(items []T) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		_, loaded := signInLocks.LoadOrStore(item, struct{}{})
		if loaded {
			continue
		}
		q.slots = append(q.slots, item)
	}
	select {
	case q.notify <- struct{}{}:
	default:
	}
}

func (q *RichQueue[T]) Dequeue() (T, error) {
	q.Lock()
	defer q.Unlock()

	if len(q.slots) == 0 {
		var zero T
		return zero, fmt.Errorf("queue is empty")
	}
	item := q.slots[0]
	q.slots = q.slots[1:]

	return item, nil
}

// ✅ BatchDequeue 方法，防止 nil 切片返回，避免 panic
func (q *RichQueue[T]) BatchDequeue(size int) ([]T, error) {
	q.Lock()
	defer q.Unlock()

	if size <= 0 {
		return []T{}, fmt.Errorf("wrong size") // 返回空切片
	}
	if len(q.slots) == 0 {
		return []T{}, fmt.Errorf("empty length") // 返回空切片
	}

	var ret []T
	if len(q.slots) > size {
		ret = make([]T, size)
		copy(ret, q.slots[:size])
		q.slots = q.slots[size:]
	} else {
		ret = q.slots
		q.slots = []T{}
	}
	return ret, nil
}

func (q *RichQueue[T]) Size() int {
	q.Lock()
	defer q.Unlock()
	return len(q.slots)
}

func (q *RichQueue[T]) First() T {
	q.Lock()
	defer q.Unlock()
	var zero T
	if len(q.slots) == 0 {
		return zero
	}
	return q.slots[0]
}

func (q *RichQueue[T]) Last() T {
	q.Lock()
	defer q.Unlock()
	var zero T
	if len(q.slots) == 0 {
		return zero
	}
	return q.slots[len(q.slots)-1]
}

// ✅ Consumer 方法，避免死锁 & CPU 100% 问题
func (q *RichQueue[T]) ConsumerWithContext(ctx context.Context, size int, handler processHandler[T]) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-q.notify:
			for {
				items, err := q.BatchDequeue(size)
				if err != nil {
					break
				}
				var wg sync.WaitGroup
				for _, item := range items {
					wg.Add(1)
					go func(item T) {
						defer wg.Done()
						defer signInLocks.Delete(item)
						defer func() {
							if r := recover(); r != nil {
								log.Errorf("Recover queue handler from panic: %v\n", r)
							}
						}()
						handler(item, &wg)
					}(item)
				}
				wg.Wait()
			}
		}
	}
}
