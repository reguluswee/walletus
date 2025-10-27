package system

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestUserNo(t *testing.T) {
	g, err := New(1, "Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		no, err := g.Generate(ctx)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("NO:", no)
	}

	// 看看当天 last_seq 是否增长
	var last int
	yymmdd := time.Now().In(time.FixedZone("CST-8", 8*3600)).Format("060102")
	if err := GetDb().Raw(`SELECT last_seq FROM user_seq WHERE yymmdd=? AND flag=?`, yymmdd, 1).Scan(&last).Error; err != nil {
		t.Fatal(err)
	}
	fmt.Println("1DB last_seq =", last)
}
