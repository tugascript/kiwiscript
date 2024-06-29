package tokens

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
)

type TokenSecretData struct {
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
	ttlSec     int64
}

func NewTokenSecretData(publicKey, privateKey string, ttlSec int64) TokenSecretData {
	publicKeyBlock, _ := pem.Decode([]byte(publicKey))
	if publicKeyBlock == nil || publicKeyBlock.Type != "PUBLIC KEY" {
		panic("Invalid public key")
	}

	publicKeyData, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		panic(err)
	}

	publicKeyValue, ok := publicKeyData.(ed25519.PublicKey)
	if !ok {
		panic("Invalid public key")
	}

	privateKeyBlock, _ := pem.Decode([]byte(privateKey))
	if privateKeyBlock == nil || privateKeyBlock.Type != "PRIVATE KEY" {
		panic("Invalid private key")
	}

	privateKeyData, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		panic(err)
	}

	privateKeyValue, ok := privateKeyData.(ed25519.PrivateKey)
	if !ok {
		panic("Invalid private key")
	}

	return TokenSecretData{
		publicKey:  publicKeyValue,
		privateKey: privateKeyValue,
		ttlSec:     ttlSec,
	}
}

type Tokens struct {
	iss         string
	accessData  TokenSecretData
	refreshData TokenSecretData
	emailData   TokenSecretData
}

func NewTokens(accessData, refreshData, emailData TokenSecretData, url string) *Tokens {
	return &Tokens{
		accessData:  accessData,
		refreshData: refreshData,
		emailData:   emailData,
		iss:         url,
	}
}

func (t *Tokens) GetAccessTtl() int64 {
	return t.accessData.ttlSec
}

func (t *Tokens) GetRefreshTtl() int64 {
	return t.refreshData.ttlSec
}

func (t *Tokens) GetEmailTtl() int64 {
	return t.emailData.ttlSec
}
