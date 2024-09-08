package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Generate a new ECDSA private key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	// Save the private key
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		log.Fatal(err)
	}
	privPem := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	})
	privKeyPath := filepath.Join(".", "keys", "ec_private_key.pem")
	if err := os.WriteFile(privKeyPath, privPem, 0600); err != nil {
		log.Fatal(err)
	}

	// Save the public key
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		log.Fatal(err)
	}
	pubPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	pubKeyPath := filepath.Join(".", "keys", "ec_public_key.pem")
	if err := os.WriteFile(pubKeyPath, pubPem, 0644); err != nil {
		log.Fatal(err)
	}
}
