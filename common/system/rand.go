package system

import (
	"context"
	"crypto/hmac"
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateNonce(length int) string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// GenerateShortInviteCode 生成5-8位的数字和大小写字母邀请码
func GenerateShortInviteCode() string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := rand.Intn(4) + 5 // 5~8
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type userSeq struct {
	Yymmdd    string    `gorm:"column:yymmdd;type:char(6);primaryKey"`
	Flag      uint8     `gorm:"column:flag;primaryKey"`
	LastSeq   uint      `gorm:"column:last_seq;not null;default:0"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (userSeq) TableName() string { return "user_seq" }

type Generator struct {
	db       *gorm.DB
	secret   []byte
	Flag     int
	loc      *time.Location
	TimeZone string
}

func New(flag int, timezone string) (*Generator, error) {
	if flag < 0 || flag > 9 {
		return nil, fmt.Errorf("flag must be 0..9")
	}
	loc := time.UTC
	if timezone != "" {
		l, err := time.LoadLocation(timezone)
		if err != nil {
			return nil, err
		}
		loc = l
	}
	return &Generator{db: GetDb(), Flag: flag, secret: []byte("nnnnn"), loc: loc, TimeZone: timezone}, nil
}

// ------------ 事务拿 seq：返回 0..9999 -------------
func (g *Generator) nextSeq(ctx context.Context, yymmdd string) (int, error) {
	var seq int // 用来带出闭包里的结果
	err := g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rec userSeq
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("yymmdd = ? AND flag = ?", yymmdd, g.Flag).
			Take(&rec).Error

		if err == nil {
			if rec.LastSeq >= 9999 {
				return fmt.Errorf("daily capacity exceeded")
			}
			seq = int(rec.LastSeq) // 展示用 0000..9999
			return tx.Model(&userSeq{}).
				Where("yymmdd = ? AND flag = ?", yymmdd, g.Flag).
				Update("last_seq", gorm.Expr("last_seq + 1")).Error
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 当天第一次：插入 1，展示 0
			rec = userSeq{Yymmdd: yymmdd, Flag: uint8(g.Flag), LastSeq: 1}
			if err := tx.Create(&rec).Error; err != nil {
				return err
			}
			seq = 0
			return nil
		}
		return err
	}, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, err
	}
	return seq, nil
}

// ------------ 混淆 yyMMdd：数字置换 + 循环右移 -------------
func buildDigitPerm(secret []byte) [10]byte {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte("dateperm"))
	seed := m.Sum(nil)
	var perm [10]byte
	for i := 0; i < 10; i++ {
		perm[i] = byte('0' + i)
	}
	k := 0
	for i := 9; i > 0; i-- {
		j := int(seed[k]) % (i + 1)
		perm[i], perm[j] = perm[j], perm[i]
		k++
		if k >= len(seed) {
			k = 0
		}
	}
	return perm
}
func buildRot(secret []byte) int {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte("rot"))
	return int(m.Sum(nil)[0] % 6)
}
func obfuscateYYMMDD(t time.Time, secret []byte, loc *time.Location) string {
	tt := t.In(loc)
	s := fmt.Sprintf("%02d%02d%02d", tt.Year()%100, int(tt.Month()), tt.Day())
	perm := buildDigitPerm(secret)
	rot := buildRot(secret)
	b := make([]byte, 6)
	for i := 0; i < 6; i++ {
		b[i] = perm[s[i]-'0']
	}
	if rot > 0 {
		r := rot % 6
		tmp := append([]byte{}, b[6-r:]...)
		b = append(tmp, b[:6-r]...)
	}
	return string(b)
}

// ------------ 3 位随机、Luhn -------------
func rand3() (string, error) {
	var buf [2]byte
	if _, err := cryptoRand.Read(buf[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("%03d", binary.BigEndian.Uint16(buf[:])%1000), nil
}
func luhnDigit(num string) int {
	sum, dbl := 0, true
	for i := len(num) - 1; i >= 0; i-- {
		d := int(num[i] - '0')
		if dbl {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		dbl = !dbl
	}
	return (10 - (sum % 10)) % 10
}

// ------------ 对外：生成 15 位编号 -------------
// flag(1) + 混淆yyMMdd(6) + seq(4) + rnd(3) + Luhn(1)
func (g *Generator) Generate(ctx context.Context) (string, error) {
	now := time.Now().In(g.loc)
	yymmdd := now.Format("060102")
	obf := obfuscateYYMMDD(now, g.secret, g.loc)

	seq, err := g.nextSeq(ctx, yymmdd) // 0..9999
	if err != nil {
		return "", err
	}

	r3, err := rand3()
	if err != nil {
		return "", err
	}

	body := fmt.Sprintf("%1d%s%04d%s", g.Flag, obf, seq, r3)
	cd := luhnDigit(body)
	return body + fmt.Sprintf("%d", cd), nil
}
