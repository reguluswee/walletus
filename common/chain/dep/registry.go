package dep

import (
	"sync"
)

type clientEntry struct {
	reader Client
}

var (
	regMu    sync.RWMutex
	registry = map[ChainCode]clientEntry{}
)

func Register(chain ChainDef, c Client) {
	regMu.Lock()
	defer regMu.Unlock()
	registry[ChainCode(chain.Name)] = clientEntry{reader: c}
}

func GetClient(chain ChainDef) (Client, bool) {
	regMu.RLock()
	defer regMu.RUnlock()
	e, ok := registry[ChainCode(chain.Name)]
	return e.reader, ok
}
