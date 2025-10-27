package security

import (
	"log"
	"strings"
	"testing"
)

func TestAES(t *testing.T) {
	orig := "123cbkcku|asdfas|120848"

	res, err := Encrypt([]byte(orig))
	o, err := Decrypt(res)

	log.Println(res, err, o)

	result := strings.Split(o, "|")
	log.Println(result)
}
