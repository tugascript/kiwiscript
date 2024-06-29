package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	// Check if the keys directory exists
	if _, err := os.Stat("../keys"); os.IsNotExist(err) {
		// Create the keys directory
		err := os.Mkdir("../keys", 0755)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Generate a new key pair
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Encode the public key to PEM format
	pubKey, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		fmt.Println(err)
		return
	}
	pubBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKey,
	}

	// Encode the private key to PEM format
	privKey, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		fmt.Println(err)
		return
	}
	privBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKey,
	}

	// Write the keys to files
	pubFile, err := os.Create("../keys/public.key")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer pubFile.Close()
	pem.Encode(pubFile, &pubBlock)

	privFile, err := os.Create("../keys/private.key")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer privFile.Close()
	pem.Encode(privFile, &privBlock)
}
