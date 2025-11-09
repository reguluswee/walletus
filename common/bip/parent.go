package bip

import (
	"fmt"

	"github.com/btcsuite/btcutil/hdkeychain"
	bip39 "github.com/tyler-smith/go-bip39"
)

const Hardened = hdkeychain.HardenedKeyStart
const tenantSecretPassword = "tenant_secret_password"

var chains = []ChainDef{
	{"ETH", 60},
	{"BSC", 60},
	{"OP", 60},
	{"ARB", 60},
	{"POLYGON", 60},
	{"TRON", 195},
	{"SOLANA", 501},
}

type KDFParams struct {
	Alg  string `json:"alg"`
	Salt string `json:"salt"` // hex
	Time uint32 `json:"time"`
	Mem  uint32 `json:"mem"`
	Par  uint8  `json:"par"`
}

type ChainDef struct {
	Name     string
	CoinType uint32
}

type EncMaster struct {
	// gcm:<base64(nonce|ciphertext|tag)>
	EncMasterXprv string    `json:"enc_master_xprv"`
	EncMasterSeed string    `json:"enc_master_seed"`
	KDF           KDFParams `json:"kdf_params"`
}

func GenerateMasterSeed() ([]byte, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return nil, fmt.Errorf("Generate entropy error: %s", err.Error())
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, fmt.Errorf("Generate mnemonic error: %s", err.Error())
	}
	seed := bip39.NewSeed(mnemonic, "") // empty passphrase, could be enhanced in future
	return seed, nil
}

func GenerateMasterXprv() (EncMaster, error) {
	var enc EncMaster
	seed, err := GenerateMasterSeed()
	if err != nil {
		return enc, err
	}

	master, err := hdkeychain.NewMaster(seed, MainNetParamsLikeBIP32())
	if err != nil {
		return enc, err
	}

	masterXprv := master.String()

	// === 2) 用 Argon2id + AES-GCM 加密 master xprv，得到 enc_master_xprv / kdf_params ===
	enc = encryptMaster([]byte(masterXprv), seed, []byte(tenantSecretPassword))

	return enc, nil
}

func GenerateDerivationChain(tenantIndex uint32, enc EncMaster, chainCode string) (ChainDerivedPath, error) {
	chainDef, err := CheckValidChainCode(chainCode)
	if err != nil {
		return ChainDerivedPath{}, fmt.Errorf("unsupport chain: %s", chainCode)
	}

	if chainDef.Name != "SOLANA" {
		return GenerateEvmDerivationChain(tenantIndex, enc, chainDef)
	}

	return GenerateSolDerivationChain(tenantIndex, enc)
}

func DeriveAddressFromXpub(enc EncMaster, xpub string, tenantIndex, addressIndex uint32, chainCode string) (addr string, path string, err error) {
	chainDef, err := CheckValidChainCode(chainCode)
	if err != nil {
		return "", "", fmt.Errorf("unsupport chain: %s", chainCode)
	}

	if chainDef.Name != "SOLANA" {
		return DeriveEvmAddressFromXpub(xpub, tenantIndex, addressIndex, chainDef)
	}

	solAddr, err := DeriveSOL(enc, tenantIndex, addressIndex)
	if err != nil {
		return "", "", err
	}
	return solAddr.Address, solAddr.Path, nil
}
