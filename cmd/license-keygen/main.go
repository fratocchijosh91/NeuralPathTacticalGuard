package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
)

func main() {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("errore generazione keypair: %v", err)
	}

	fmt.Printf("NP_LICENSE_PRIVATE_KEY_B64=%s\n", base64.StdEncoding.EncodeToString(privateKey))
	fmt.Printf("NP_LICENSE_PUBLIC_KEY_B64=%s\n", base64.StdEncoding.EncodeToString(publicKey))
}
