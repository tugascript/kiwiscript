// Copyright (C) 2024 Afonso Barracha
//
// This file is part of KiwiScript.
//
// KiwiScript is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// KiwiScript is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with KiwiScript.  If not, see <https://www.gnu.org/licenses/>.

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
	oauthData   TokenSecretData
}

func NewTokens(accessData, refreshData, emailData, oauthData TokenSecretData, url string) *Tokens {
	return &Tokens{
		accessData:  accessData,
		refreshData: refreshData,
		emailData:   emailData,
		oauthData:   oauthData,
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

func (t *Tokens) GetOAuthTtl() int64 {
	return t.oauthData.ttlSec
}
