// Learn about Plasma Prime RSA Accumulators
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package accumulator implements a cryptographic accumulator.
// An accumulator is like a merkle tree but the proofs are constant size.
// This package is just a toy.
package accumulator

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"time"

	"vuvuzela.io/concurrency"
)

const (
	maxprimes = 10000
)

var basic_primes []*big.Int

func init() {
	st := time.Now()
	basic_primes = GeneratePrimes(maxprimes)
	fmt.Printf("Computed %d basic primes from %v to %v (%s)\n", len(basic_primes), basic_primes[0], basic_primes[len(basic_primes)-1], time.Since(st))
}

func GeneratePrimes(nprimes int64) []*big.Int {
	primes := make([]*big.Int, nprimes)
	primes[0] = big.NewInt(2)
	n := int64(1)
	t := new(big.Int)
	for i := int64(3); n < nprimes; i += 2 {
		t.SetInt64(i)
		if t.ProbablyPrime(10) {
			primes[n] = big.NewInt(i)
			n++
		}
	}
	return primes
}

// PrivateKey is the private key for an RSA accumulator.
// It is not needed for typical uses of an accumulator.
type PrivateKey struct {
	P, Q    *big.Int
	N       *big.Int // N = P*Q
	Totient *big.Int // Totient = (P-1)*(Q-1)
}

type PublicKey struct {
	N *big.Int
}

var base = big.NewInt(3)
var bigOne = big.NewInt(1)
var bigTwo = big.NewInt(2)

// GenerateKey generates an RSA accumulator keypair. The private key
// is mostly used for debugging and should usually be destroyed
// as part of a trusted setup phase.
func GenerateKey(random io.Reader) (*PublicKey, *PrivateKey, error) {
	for {
		p, err := rand.Prime(random, 1024)
		if err != nil {
			return nil, nil, err
		}
		q, err := rand.Prime(random, 1024)
		if err != nil {
			return nil, nil, err
		}

		pminus1 := new(big.Int).Sub(p, bigOne)
		qminus1 := new(big.Int).Sub(q, bigOne)
		totient := new(big.Int).Mul(pminus1, qminus1)

		g := new(big.Int).GCD(nil, nil, base, totient)
		if g.Cmp(bigOne) == 0 {
			privateKey := &PrivateKey{
				P:       p,
				Q:       q,
				N:       new(big.Int).Mul(p, q),
				Totient: totient,
			}
			publicKey := &PublicKey{
				N: new(big.Int).Set(privateKey.N),
			}
			return publicKey, privateKey, nil
		}
	}
}

func FermatPrime(n *big.Int) bool {
	one := big.NewInt(1)
	t0 := new(big.Int)
	t0.Mod(t0.Exp(big.NewInt(2), t0.Sub(n, one), nil), n)
	return t0.Cmp(one) == 0
}

func DivisibleByFirstFewPrimes(n *big.Int) int64 {
	z := big.NewInt(0)
	arr := [9]int64{3, 5, 7, 11, 13, 17, 19, 23, 29}
	for _, i := range arr {
		d := big.NewInt(i)
		m := new(big.Int).Mod(n, d)
		if m.Cmp(z) == 0 {
			return i
		}
	}
	return 0
}

func CoinIDToPrime(coinid int64) *big.Int {
	// TODO: try the 25000 * coinid idea with FermatPrime, DivisibleBy
	if coinid < maxprimes {
		return basic_primes[coinid]
	}
	panic("invalid coinid seen")
	return nil
}

func (key *PublicKey) Accumulate(coinid []int64) (acc *big.Int, witnesses []*big.Int) {
	primes := make([]*big.Int, len(coinid))
	for i, ci := range coinid {
		primes[i] = CoinIDToPrime(ci)
	}

	st := time.Now()
	acc = new(big.Int).Set(base)
	for i := range primes {
		acc.Exp(acc, primes[i], key.N)
	}
	fmt.Printf("(Acc compute: %s) |", time.Since(st))

	st = time.Now()
	witnesses = make([]*big.Int, len(coinid))
	concurrency.ParallelFor(len(coinid), func(p *concurrency.P) {
		for i, ok := p.Next(); ok; i, ok = p.Next() {
			// TODO reuse computations
			wit := new(big.Int).Set(base)
			for j := range primes {
				if j != i {
					wit.Exp(wit, primes[j], key.N)
				}
			}
			witnesses[i] = wit
		}
	})
	fmt.Printf("(Witness compute: %s) |", time.Since(st))
	return
}

func (key *PublicKey) Verify(acc *big.Int, witness *big.Int, coinid int64) bool {
	c := CoinIDToPrime(coinid)
	v := new(big.Int).Exp(witness, c, key.N)
	return acc.Cmp(v) == 0
}
