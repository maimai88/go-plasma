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
	"math/big"
	mrand "math/rand"
	"testing"
	"time"
)

func TestPlasmaAccumulate(t *testing.T) {
	publicKey, _, err := GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("\n***** Genesis A_0=%v\n", base)

	// **** Block B_1 has coins c_0=2 and c_3=7 with txs
	block1_coins := make([]int64, 2)
	block1_coins[0] = 0 // 2
	block1_coins[1] = 3 // 7
	acc, witnesses := publicKey.Accumulate(block1_coins)
	// acc should be 3^14
	expected_acc1 := new(big.Int).Exp(base, big.NewInt(2*7), publicKey.N)
	if acc.Cmp(expected_acc1) != 0 {
		t.Fatalf("Unexpected accumulator")
	} else {
		fmt.Printf("\n***** Block B_1: A_1=%v (CHECKED)\n", acc)
		base = acc
	}

	// use c_5=13 (it will not be included)
	badItem := int64(5) // 13

	for i, w := range witnesses {
		// check that c_i in "block1_coins" is in the accumulator:
		fmt.Printf(" Coin c_%d = %v Witness: [%v]", block1_coins[i], CoinIDToPrime(block1_coins[i]), w)
		if !publicKey.Verify(acc, w, block1_coins[i]) {
			t.Fatal("... NOT FOUND!")
		} else {
			fmt.Printf("... Verified!\n")
		}

		// check that the badItem (c_4 = 11) isn't verified somehow..
		if publicKey.Verify(acc, w, badItem) {
			t.Fatal("bad item was verified")
		}
		// check that the next item (c_(i+1)) isn't verified somehow..
		if publicKey.Verify(acc, w, block1_coins[(i+1)%len(block1_coins)]) {
			t.Fatal("bad item was verified")
		}
	}

	// **** Block B_2 has coins c_1=3 and c_2=5 and c_4=11 with txs
	block2_coins := make([]int64, 3)
	block2_coins[0] = 1 // 3
	block2_coins[1] = 2 // 5
	block2_coins[2] = 4 // 11
	acc, witnesses = publicKey.Accumulate(block2_coins)
	expected_acc2 := new(big.Int).Exp(expected_acc1, big.NewInt(3*5*11), publicKey.N)
	if acc.Cmp(expected_acc2) != 0 {
		t.Fatalf("Unexpected accumulator")
	} else {
		fmt.Printf("\n***** Block B_2: A_2=%v W_2=[%v]\n", acc, witnesses)
		base = acc
	}
	for i, w := range witnesses {
		// check that c_i in "block2_coins" is in the accumulator:
		fmt.Printf(" Coin c_%d = %v Witness: [%v]", block2_coins[i], CoinIDToPrime(block2_coins[i]), w)
		if !publicKey.Verify(acc, w, block2_coins[i]) {
			t.Fatal("... NOT FOUND!")
		} else {
			fmt.Printf("... Verified!\n")
		}

		// check that the badItem (c_4 = 11) isn't verified somehow..
		if publicKey.Verify(acc, w, badItem) {
			t.Fatal("bad item was verified")
		}
		// check that the next item (c_(i+1)) isn't verified somehow..
		if publicKey.Verify(acc, w, block2_coins[(i+1)%len(block2_coins)]) {
			t.Fatal("bad item was verified")
		}
	}

	// **** Block B_2 has coins c_1=3 and c_2=5 and c_4=11 with txs
	for block := 3; block <= 10; block++ {
		var p float64
		switch block {
		case 3:
			p = .001
		case 4:
			p = .002
		case 5:
			p = .003
		case 6:
			p = .004
		case 7:
			p = .005
		case 8:
			p = .01
		case 9:
			p = .025
		case 10:
			p = .05
		}
		block_coins := make([]int64, 0)
		for i := int64(6); i < maxprimes; i++ {
			if mrand.Float64() < p {
				block_coins = append(block_coins, i)
			}
		}
		st := time.Now()
		fmt.Printf("\n**** Block B_%d - %d coins included | ", block, len(block_coins))
		acc, witnesses = publicKey.Accumulate(block_coins)
		base = acc
		verifications := 0
		st = time.Now()
		for i, w := range witnesses {
			// check that c_i in "block2_coins" is in the accumulator:
			// fmt.Printf(" Coin c_%d = %v Witness: [%v]", block_coins[i], CoinIDToPrime(block_coins[i]), w)
			if !publicKey.Verify(acc, w, block_coins[i]) {
				t.Fatal("... NOT FOUND!")
			}

			// check that the badItem (c_4 = 11) isn't verified somehow..
			if publicKey.Verify(acc, w, badItem) {
				t.Fatal("bad item was verified")
			}
			// check that the next item (c_(i+1)) isn't verified somehow..
			if publicKey.Verify(acc, w, block_coins[(i+1)%len(block_coins)]) {
				t.Fatal("bad item was verified")
			}
			verifications += 3
		}
		fmt.Printf("%d Verify ops (%s) | ", verifications, time.Since(st))
	}
	// do a proof of EXCLUSION on c_5=13  for the last 4 blocks
}

func TestFermat(t *testing.T) {
	t.SkipNow()
	st := time.Now()
	s := 1000001
	e := s + 10000
	evals := 0
	res := make([]int, 5)
	for i := int64(s); i < int64(e); i += 2 {
		evals++
		n := big.NewInt(i)
		if FermatPrime(n) {
			q := DivisibleByFirstFewPrimes(n)
			if q > 0 {
				res[0]++
				fmt.Printf("%v passed Fermat primality test ... but then is divisible by %d\n", n, q)
			} else if n.ProbablyPrime(10) == false {
				res[1]++
				fmt.Printf("%v passed Fermat primality test ... but ProbablyPrime returned false\n", n)
			} else {
				res[2]++
			}
		} else if n.ProbablyPrime(10) == true {
			res[3]++
		} else {
			res[4]++
		}
		if evals%1000 == 0 {
			fmt.Printf("%s Evals %d cases: %v\n", time.Since(st), evals, res)
		}
	}
}
