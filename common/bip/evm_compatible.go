package bip

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/mr-tron/base58"
	"github.com/reguluswee/walletus/common/chain/dep"
	"golang.org/x/crypto/argon2"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
)

var chains = dep.GetSupportedChains()

type ChainDerivedPath struct {
	Chain       dep.ChainDef
	DerivedPath string
	XPub        string
}

func SupportChains() []dep.ChainDef {
	return chains
}

func CheckValidChainCode(chainCode string) (dep.ChainDef, error) {
	if chainCode == "arbitrum" {
		chainCode = "ARB"
	}
	for _, v := range chains {
		if chainCode == v.Name {
			return v, nil
		}
	}
	return dep.ChainDef{}, fmt.Errorf("unsupport chain %s", chainCode)
}

func GenerateEvmDerivationChain(tenantIndex uint32, enc EncMaster, chainDef dep.ChainDef) (ChainDerivedPath, error) {
	var cdp ChainDerivedPath
	plainMaster, err := decryptMasterXprv(enc, []byte(tenantSecretPassword))
	if err != nil {
		return cdp, err
	}
	master2, err := hdkeychain.NewKeyFromString(string(plainMaster))

	if err != nil {
		return cdp, err
	}
	// path: m/44'/coin'/account'
	node, err := derivePathMust(master2, []uint32{
		44 + Hardened,
		chainDef.CoinType + Hardened,
		tenantIndex + Hardened,
	})
	if err != nil {
		return cdp, err
	}
	// neuter -> xpub
	xpubNode, err := node.Neuter()
	if err != nil {
		return cdp, err
	}
	xpub := xpubNode.String()
	fmt.Printf("[%s] xpub: %s\n", chainDef.Name, xpub)

	cdp.Chain = chainDef
	cdp.DerivedPath = fmt.Sprintf("m/44'/%d'/%d'/0", chainDef.CoinType, tenantIndex)
	cdp.XPub = xpub

	return cdp, nil
}

func DeriveEvmAddressFromXpub(xpub string, tenantIndex, addressIndex uint32, chainDef dep.ChainDef) (addr string, path string, err error) {
	node, err := hdkeychain.NewKeyFromString(xpub)
	if err != nil {
		return "", "", err
	}
	if node.IsPrivate() {
		return "", "", fmt.Errorf("expected xpub, got xprv")
	}

	// 2) 非硬化派生：/0/index
	ext, err := node.Child(0) // change=0 外部地址
	if err != nil {
		return "", "", err
	}
	leaf, err := ext.Child(addressIndex) // 非硬化：index < 2^31
	if err != nil {
		return "", "", err
	}

	// 3) 取公钥并编码为链上地址
	pub, err := leaf.ECPubKey()
	if err != nil {
		return "", "", err
	}
	ecdsaPub := pub.ToECDSA()

	switch chainDef.Name {
	case "ETH", "BSC", "OP", "ARB", "POLYGON":
		// EVM 兼容链地址：keccak(pub) 后 20 字节
		addr = gethcrypto.PubkeyToAddress(*ecdsaPub).Hex()
	case "TRON":
		// TRON 地址：Base58Check( 0x41 || evm20bytes || checksum4 )
		evm := gethcrypto.PubkeyToAddress(*ecdsaPub).Bytes() // 20 bytes
		raw := append([]byte{0x41}, evm...)
		sum := sha256.Sum256(raw)
		sum = sha256.Sum256(sum[:])
		full := append(raw, sum[0:4]...)
		addr = base58.Encode(full)
	default:
		return "", "", fmt.Errorf("unsupported chain: %s", chainDef.Name)
	}

	path = fmt.Sprintf("m/44'/%d'/%d'/0/%d", chainDef.CoinType, tenantIndex, addressIndex)

	return addr, path, nil
}

func AddressAndPrivFromPath(enc EncMaster, path, chainCode string) (addr string, priv *ecdsa.PrivateKey, err error) {
	// 1) 派生到叶子 xprv
	leaf, err := DeriveChildXprv(enc, path)
	if err != nil {
		return "", nil, err
	}

	// 2) 取 ECDSA 私钥 & 公钥
	btcecPriv, err := leaf.ECPrivKey()
	if err != nil {
		return "", nil, err
	}
	priv = btcecPriv.ToECDSA()
	pub := &priv.PublicKey

	// 3) 计算不同链的地址
	switch chainCode {
	case "ETH", "BSC", "OP", "ARB", "POLYGON":
		addr = gethcrypto.PubkeyToAddress(*pub).Hex() // 0x...
	case "TRON":
		evm20 := gethcrypto.PubkeyToAddress(*pub).Bytes() // 20 bytes
		raw := append([]byte{0x41}, evm20...)             // Tron前缀
		sum := sha256.Sum256(raw)
		sum = sha256.Sum256(sum[:])                    // doubleSha256
		addr = base58.Encode(append(raw, sum[0:4]...)) // T...
	default:
		return "", nil, fmt.Errorf("unsupported chain: %s", chainCode)
	}
	return addr, priv, nil
}

func DeriveChildXprv(enc EncMaster, path string) (*hdkeychain.ExtendedKey, error) {
	plain, err := decryptMasterXprv(enc, []byte(tenantSecretPassword))
	if err != nil {
		return nil, err
	}
	defer zero(plain)

	master, err := hdkeychain.NewKeyFromString(string(plain))
	if err != nil {
		return nil, err
	}

	idxs, err := parseBIP44Path(path) // e.g. m/44'/60'/1'/0/2 -> []uint32 with Hardened
	if err != nil {
		return nil, err
	}

	node := master
	for i := 1; i < len(idxs); i++ { // skip "m"
		node, err = node.Child(idxs[i])
		if err != nil {
			return nil, err
		}
	}
	if !node.IsPrivate() {
		return nil, fmt.Errorf("leaf is not private")
	}
	return node, nil // node.String() = 子 xprv
}

func parseBIP44Path(path string) ([]uint32, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}
	if !strings.HasPrefix(path, "m") {
		return nil, fmt.Errorf("invalid path: must start with 'm'")
	}

	// 仅 m 的情况
	if path == "m" {
		return []uint32{0}, nil
	}

	parts := strings.Split(path, "/")
	result := make([]uint32, 0, len(parts))
	result = append(result, 0) // 第一位留空给 m

	for _, p := range parts[1:] {
		hardened := strings.HasSuffix(p, "'")
		if hardened {
			p = strings.TrimSuffix(p, "'")
		}

		// 解析数字部分
		val, err := strconv.ParseUint(p, 10, 31)
		if err != nil {
			return nil, fmt.Errorf("invalid path segment: %s", p)
		}

		n := uint32(val)
		if hardened {
			n += hdkeychain.HardenedKeyStart
		}

		result = append(result, n)
	}

	return result, nil
}

// EVM 用：拿到 *ecdsa.PrivateKey
func ToECDSAPriv(child *hdkeychain.ExtendedKey) (*ecdsa.PrivateKey, error) {
	priv, err := child.ECPrivKey()
	if err != nil {
		return nil, err
	}
	return priv.ToECDSA(), nil
}

// ========= 派生工具 =========
func derivePathMust(key *hdkeychain.ExtendedKey, path []uint32) (*hdkeychain.ExtendedKey, error) {
	var curr = key
	var err error
	for _, i := range path {
		curr, err = curr.Child(i)
		if err != nil {
			return nil, err
		}
	}
	return curr, nil
}

// Ethereum/Polygon 地址（keccak 公钥，后 20 字节）
func ethAddressFromPub(pub *ecdsa.PublicKey) string {
	addr := gethcrypto.PubkeyToAddress(*pub)
	return addr.Hex()
}

// Tron 地址：base58check( 0x41 || evm20bytes )
func tronAddressFromPub(pub *ecdsa.PublicKey) string {
	evm := gethcrypto.PubkeyToAddress(*pub) // 20 bytes
	evmBytes := evm.Bytes()                 // len=20
	// prefix 0x41
	prefix := []byte{0x41}
	raw := append(prefix, evmBytes...)
	// checksum = sha256d(raw) 前4字节
	sum := sha256.Sum256(raw)
	sum = sha256.Sum256(sum[:])
	full := append(raw, sum[0:4]...)
	return base58.Encode(full)
}

// ========= 加密/解密 =========
func encryptMaster(plain, seed []byte, password []byte) EncMaster {
	// 生成随机 salt
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	// Argon2id 派生密钥（32字节）
	key := argon2.IDKey(password, salt, 3, 64*1024, 1, 32)

	block, err := aes.NewCipher(key)
	must(err)
	aesgcm, err := cipher.NewGCM(block)
	must(err)

	noncePrv := make([]byte, aesgcm.NonceSize())
	_, _ = rand.Read(noncePrv)
	ct := aesgcm.Seal(nil, noncePrv, plain, nil)
	buf := append(noncePrv, ct...) // 末尾已含 GCM tag

	nonceSeed := make([]byte, aesgcm.NonceSize())
	_, _ = rand.Read(nonceSeed)
	ctSeed := aesgcm.Seal(nil, nonceSeed, seed, nil)
	bufSeed := append(nonceSeed, ctSeed...)

	return EncMaster{
		EncMasterXprv: "gcm:" + base64.StdEncoding.EncodeToString(buf),
		EncMasterSeed: "gcm:" + base64.StdEncoding.EncodeToString(bufSeed),
		KDF: KDFParams{
			Alg:  "argon2id",
			Salt: hex.EncodeToString(salt),
			Time: 3, Mem: 64 * 1024, Par: 1,
		},
	}
}

func decryptMasterXprv(enc EncMaster, password []byte) ([]byte, error) {
	if !strings.HasPrefix(enc.EncMasterXprv, "gcm:") {
		return nil, fmt.Errorf("bad enc_master_xprv format")
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(enc.EncMasterXprv, "gcm:"))
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %v", err)
	}
	salt, err := hex.DecodeString(enc.KDF.Salt)
	if err != nil {
		return nil, fmt.Errorf("hex decode salt failed: %v", err)
	}
	key := argon2.IDKey(password, salt, enc.KDF.Time, enc.KDF.Mem, enc.KDF.Par, 32)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher failed: %v", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes gcm failed: %v", err)
	}

	nonceSize := aesgcm.NonceSize()
	if len(raw) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ct := raw[:nonceSize], raw[nonceSize:]
	plain, err := aesgcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("aes gcm open failed: %v", err)
	}
	return plain, nil // string(plain) 即 xprv
}

func decryptMasterSeed(enc EncMaster, password []byte) ([]byte, error) {
	if !strings.HasPrefix(enc.EncMasterSeed, "gcm:") {
		return nil, fmt.Errorf("bad enc_master_seed format")
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(enc.EncMasterSeed, "gcm:"))
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %v", err)
	}
	salt, err := hex.DecodeString(enc.KDF.Salt)
	if err != nil {
		return nil, fmt.Errorf("hex decode salt failed: %v", err)
	}
	key := argon2.IDKey(password, salt, enc.KDF.Time, enc.KDF.Mem, enc.KDF.Par, 32)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher failed: %v", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes gcm failed: %v", err)
	}

	nonceSize := aesgcm.NonceSize()
	if len(raw) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ct := raw[:nonceSize], raw[nonceSize:]
	plain, err := aesgcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("aes gcm open failed: %v", err)
	}
	return plain, nil // string(plain) 即 xprv
}

// ========= BIP32 网络参数（仅用于序列化前缀；这里复用比特系前缀，不影响推导数学） =========

// MainNetParamsLikeBIP32 返回一组满足 hdkeychain 需求的“前缀参数”。
// 这里只是为了能生成/解析 xprv/xpub 字符串；不影响推导逻辑。
func MainNetParamsLikeBIP32() *chaincfg.Params {
	// 使用 btcutil 内置的主网前缀（xprv/xpub）
	return &chaincfg.MainNetParams
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// （可选）把内存中的敏感切片清零
func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// （可选）比较常量时间字符串相等
func constEq(a, b string) bool {
	return subtleCompare([]byte(a), []byte(b))
}

func subtleCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	diff := byte(0)
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}
