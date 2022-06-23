package types

import (
	"math/rand"
	"testing"
)

func generateKey32() [32]byte {
	var rig [32]byte
	n, _ := rand.Read(rig[:])
	if n != 32 {
		panic("not 32")
	}
	return rig
}

func bytesEqual(b1, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}
	for i, x := range b1 {
		if x != b2[i] {
			return false
		}
	}
	return true
}

func Test_SymmetricEncryption(t *testing.T) {
	key := generateKey32()
	data := generateKey32() // whatever
	ciphertxt, err := EncryptSymmetric(data[:], key[:])
	if err != nil {
		panic(err)
	}
	plaintxt, err := DecryptSymmetric(ciphertxt, key[:])
	if err != nil {
		panic(err)
	}
	if !bytesEqual(plaintxt, data[:]) {
		t.Fatalf("Not equal")
	}
}
