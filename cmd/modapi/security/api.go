package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
)

const secretSalt = "oP3z@6N!cY%2Qm8#Ls&bT7xR^F1uW$h9K*Za4EpjDqG"

func GenerateNDigitNumber(n int) (string, error) {
	if n < 5 {
		return "", fmt.Errorf("digit length must >= 5")
	}

	result := make([]byte, n)
	for i := range n {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		result[i] = byte(num.Int64()) + '0'
	}

	return string(result), nil
}

func generatePrefix(uniqueID string, id, name string) string {

	base := fmt.Sprintf("%s|%s|%s", uniqueID, name, id)

	hash := sha256.Sum256([]byte(base))

	b32 := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash[:])

	return b32[:5]
}

func GenerateAppIDAndKey(uniqueID string, name string, ts int64) (string, string) {
	id, _ := GenerateNDigitNumber(6)
	prefix := generatePrefix(uniqueID, id, name)

	base := fmt.Sprintf("%s|%s|%s|%s|%d", prefix, uniqueID, id, name, ts)

	h1 := sha256.Sum256([]byte(base))
	fullHash := hex.EncodeToString(h1[:]) // 64 hex chars

	appid := prefix + "-" + fullHash[:16]

	appKeyBase := appid + "|" + strconv.FormatInt(ts, 10) + "|" + secretSalt
	h2 := sha256.Sum256([]byte(appKeyBase))
	appkey := hex.EncodeToString(h2[:])

	return appid, appkey
}
