package main

import "crypto/ecdsa"

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

type Wallets struct {
	Wallets map[string]*Wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {

}
