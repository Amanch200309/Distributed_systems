package chord

import (
	"crypto/sha1"
	"math/big"
)

/*
HashKey computes the Chord identifier for a given string.

	Args: 	s (string to hash),
			m (number of bits in hash space)
	Returns: Hashed identifier as big.Int
	Uses SHA-1 to hash the string, then applies modulo 2^m to fit in hash space
*/
func HashKey(s string, m int) *big.Int {
	h := sha1.Sum([]byte(s))                        // SHA-1 hash of input string
	n := new(big.Int).SetBytes(h[:])                // Convert hash bytes to big.Int
	mod := new(big.Int).Lsh(big.NewInt(1), uint(m)) // Calculate 2^m
	return new(big.Int).Mod(n, mod)                 // Return n mod 2^m
}
