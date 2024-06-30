package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
)

func getLog() *slog.Logger {
	if os.Getenv("DEBUG") == "true" {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	return slog.Default()
}

func main() {
	logger := getLog()
	// Check if the keys directory exists
	logger.Debug("Checking if the keys directory exists...")
	if _, err := os.Stat("../keys"); os.IsNotExist(err) {
		logger.Debug("Keys directory does not exist, creating it...")
		err := os.Mkdir("../keys", 0755)

		if err != nil {
			logger.Error("Failed to create keys directory", "error", err)
			fmt.Println(err)
			return
		}
		logger.Debug("Keys directory created")
	} else {
		logger.Debug("Keys directory already exists")
	}

	// Generate a new key pair
	logger.Debug("Generating a new ED25519 key pair...")
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		logger.Error("Failed to generate a new ED25519 key pair", "error", err)
		return
	}
	logger.Debug("ED25519 key pair generated")

	// Encode the public key to PEM format
	logger.Debug("Encoding the public key to PEM format...")
	pubKey, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		logger.Error("Failed to encode the public key to PEM format", "error", err)
		return
	}
	pubBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKey,
	}
	logger.Debug("Public key encoded to PEM format")

	// Encode the private key to PEM format
	logger.Debug("Encoding the private key to PEM format...")
	privKey, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		logger.Error("Failed to encode the private key to PEM format", "error", err)
		return
	}
	privBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKey,
	}
	logger.Debug("Private key encoded to PEM format")

	// Write public and private keys to files
	logger.Debug("Writing public key to ../keys/public.key ...")
	pubFile, err := os.Create("../keys/public.key")
	if err != nil {
		logger.Error("Failed to create public key file", "error", err)
		return
	}
	defer pubFile.Close()
	pem.Encode(pubFile, &pubBlock)
	logger.Debug("Public key written to ../keys/public.key")

	logger.Debug("Writing private key to ../keys/private.key ...")
	privFile, err := os.Create("../keys/private.key")
	if err != nil {
		logger.Error("Failed to create private key file", "error", err)
		return
	}
	defer privFile.Close()
	pem.Encode(privFile, &privBlock)
	logger.Debug("Private key written to ../keys/private.key")

	logger.Info("Key pair generated successfully")
	pubPEM := pem.EncodeToMemory(&pubBlock)
	publicKey, _ := json.Marshal(string(pubPEM))
	fmt.Println("\nPublic key value:\n", string(publicKey))

	privPEM := pem.EncodeToMemory(&privBlock)
	privateKey, _ := json.Marshal(string(privPEM))
	fmt.Println("\nPrivate key value:\n", string(privateKey))
}
