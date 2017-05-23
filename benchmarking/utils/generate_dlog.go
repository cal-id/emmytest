package utils

import (
	"crypto/dsa"
	"crypto/rand"
	"github.com/xlab-si/emmy/dlog"
	"io"
	"math/big"
)

/*
 * Generate dlog parameters p, q, g for use with the schnorr protocol
 *
 * This adapts the implementation by crypto/dsa.
 * see here for license https://golang.org/src/crypto/dsa/dsa.go?s=1669:1756#L47
 */

// GenerateDlog creates a random dlog according to some
// parameters
// Takes N = number of bits in prime q
// Takes L = number of bits in p
// Both L and N must be divisible by 8
// Returns pointer to dlog
func GenerateDlog(N int, L int) (*dlog.ZpDLog, error) {
	dlog := dlog.ZpDLog{}
	var generatedParameters dsa.Parameters
	err := GenerateParameters(&generatedParameters, rand.Reader, N, L)
	if err != nil {
		return nil, err
	}
	dlog.G = generatedParameters.G
	dlog.OrderOfSubgroup = generatedParameters.Q
	dlog.P = generatedParameters.P
	return &dlog, nil
}

// numMRTests is the number of Miller-Rabin primality tests that we perform. We
// pick the largest recommended number from table C.1 of FIPS 186-3.
const numMRTests = 64

// GenerateParameters puts a random, valid set of DSA parameters into params.
// This function can take many seconds, even on fast machines.
// Taken from crpto/dsa but removed the requirement to only produce some sizes
// of L / N
func GenerateParameters(params *dsa.Parameters, rand io.Reader, N int, L int) error {
	// This function doesn't follow FIPS 186-3 exactly in that it doesn't
	// use a verification seed to generate the primes. The verification
	// seed doesn't appear to be exported or used by other code and
	// omitting it makes the code cleaner.
	qBytes := make([]byte, N/8)
	pBytes := make([]byte, L/8)

	q := new(big.Int)
	p := new(big.Int)
	rem := new(big.Int)
	one := new(big.Int)
	one.SetInt64(1)

GeneratePrimes:
	for {
		if _, err := io.ReadFull(rand, qBytes); err != nil {
			return err
		}

		qBytes[len(qBytes)-1] |= 1
		qBytes[0] |= 0x80
		q.SetBytes(qBytes)

		if !q.ProbablyPrime(numMRTests) {
			continue
		}

		for i := 0; i < 4*L; i++ {
			if _, err := io.ReadFull(rand, pBytes); err != nil {
				return err
			}

			pBytes[len(pBytes)-1] |= 1
			pBytes[0] |= 0x80

			p.SetBytes(pBytes)
			rem.Mod(p, q)
			rem.Sub(rem, one)
			p.Sub(p, rem)
			if p.BitLen() < L {
				continue
			}

			if !p.ProbablyPrime(numMRTests) {
				continue
			}

			params.P = p
			params.Q = q
			break GeneratePrimes
		}
	}

	h := new(big.Int)
	h.SetInt64(2)
	g := new(big.Int)

	pm1 := new(big.Int).Sub(p, one)
	e := new(big.Int).Div(pm1, q)

	for {
		g.Exp(h, e, p)
		if g.Cmp(one) == 0 {
			h.Add(h, one)
			continue
		}

		params.G = g
		return nil
	}
}
