package chord

import (
	"crypto/sha1"
	"math/big"
)

func hashKey(s string, m int) *big.Int {
	h := sha1.Sum([]byte(s))
	n := new(big.Int).SetBytes(h[:])
	mod := new(big.Int).Lsh(big.NewInt(1), uint(m))
	return new(big.Int).Mod(n, mod)
}

