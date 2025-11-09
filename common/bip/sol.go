package bip

import (
	"crypto/ed25519"
	"fmt"

	slip10 "github.com/anyproto/go-slip10"
	"github.com/mr-tron/base58"
)

type DerivedSOL struct {
	Address     string
	Ed25519Priv ed25519.PrivateKey
	Ed25519Pub  ed25519.PublicKey
	Path        string
}

func DeriveSOL(enc EncMaster, tenantIdx, addrIdx uint32) (DerivedSOL, error) {
	seed, err := decryptMasterSeed(enc, []byte(tenantSecretPassword))
	if err != nil {
		return DerivedSOL{}, err
	}
	defer zero(seed)

	// 常见：m/44'/501'/tenant'/0'；多地址：再加一层 addrIdx'
	path := fmt.Sprintf("m/44'/501'/%d'/0'", tenantIdx)
	if addrIdx != 0 {
		path = fmt.Sprintf("m/44'/501'/%d'/0'/%d'", tenantIdx, addrIdx)
	}
	return deriveSOLFromSeed(seed, path)
}

func deriveSOLFromSeed(seed []byte, path string) (DerivedSOL, error) {
	// SLIP-0010(ed25519)，全硬化路径
	node, err := slip10.DeriveForPath(path, seed)
	if err != nil {
		return DerivedSOL{}, err
	}

	pubBytes, privBytes := node.Keypair() // ed25519 keypair

	priv := ed25519.PrivateKey(privBytes) // len=64
	pub := ed25519.PublicKey(pubBytes)    // len=32
	addr := base58.Encode(pub)            // Solana 地址=公钥 base58

	return DerivedSOL{
		Address: addr, Ed25519Priv: priv, Ed25519Pub: pub, Path: path,
	}, nil
}

func GenerateSolDerivationChain(tenantIndex uint32, enc EncMaster) (ChainDerivedPath, error) {
	chainDef, _ := CheckValidChainCode("SOLANA")
	return ChainDerivedPath{
		Chain:       chainDef,
		XPub:        "",
		DerivedPath: fmt.Sprintf("m/44'/501'/%d", tenantIndex),
	}, nil
}
