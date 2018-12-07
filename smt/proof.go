// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package smt

import (
	"bytes"
	"fmt"
)

type Proof struct {
	key       []byte
	proof     [][]byte
	proofBits uint64
}

func (self *Proof) Check(leaf []byte, root []byte, verbose bool) bool {

	cur := leaf
	p := 0

	for i := uint64(0); i < 64; i++ {
		if (uint64(1<<i) & self.proofBits) > 0 {
			if p >= len(self.proof) {
				if verbose {
					fmt.Printf("Invalid proofBytes Length\n")
				}
				return false
			}

			if byte(0x01<<(i%8))&byte(self.key[(TreeDepth-1-i)/8]) > 0 {
				//Non-default Hash
				// i-th bit is "1", so hash with H([]) on the left
				cur = Computehash(self.proof[p], cur)
			} else {
				// i-th bit is "0", so hash with H([]) on the right
				cur = Computehash(cur, self.proof[p])
			}
			p++
		} else {
			//DefaultHash
			if byte(0x01<<(i%8))&byte(self.key[(TreeDepth-1-i)/8]) > 0 {
				cur = Computehash(GlobalDefaultHashes[i], cur)
			} else {
				cur = Computehash(cur, GlobalDefaultHashes[i])
			}
		}
	}
	res := bytes.Compare(cur, root) == 0
	if verbose {
		if res {
			fmt.Printf(" CheckProof success (root matche: %x)\n", cur)
		} else {
			fmt.Printf(" CheckProof FAILURE (expected root [%x] does NOT match actual root: %x)\n", root, cur)
		}
	}
	return res
}

func (self *Proof) PrintSMTProof(leaf []byte) string {
	cur := leaf
	p := 0

	out := fmt.Sprintf("****\nSMTProof\n")

	for i := uint64(0); i < 64; i++ {
		if (uint64(1<<i) & self.proofBits) > 0 {
			if p >= len(self.proof) {
				return fmt.Sprintf("Missing Proof afer depth %d\n", p)
			}
			if byte(0x01<<(i%8))&byte(self.key[(TreeDepth-1-i)/8]) > 0 {
				out = out + fmt.Sprintf("H%v | [P,*] bit%v=1 | H(P[%d]:%x, H[%d]:%x) => ", i+1, i, p, self.proof[p], i, cur)
				cur = Computehash(self.proof[p], cur)
			} else {
				out = out + fmt.Sprintf("H%v | [*,P] bit%v=0 | H(H[%d]:%x, P[%d]:%x) => ", i+1, i, i, cur, p, self.proof[p])
				cur = Computehash(cur, self.proof[p])
			}
			p++
		} else {
			if byte(0x01<<(i%8))&byte(self.key[(TreeDepth-1-i)/8]) > 0 {
				out = out + fmt.Sprintf("H%v | [D,*] bit%v=1 | H(D[%d]:%x, H[%d]:%x) => ", i+1, i, i, GlobalDefaultHashes[i], i, cur)
				cur = Computehash(GlobalDefaultHashes[i], cur)
			} else {
				out = out + fmt.Sprintf("H%v | [*,D] bit%v=0 | H(H[%d]:%x, D[%d]:%x) => ", i+1, i, i, cur, i, GlobalDefaultHashes[i])
				cur = Computehash(cur, GlobalDefaultHashes[i])
			}
		}
		out = out + fmt.Sprintf(" %x\n\n", cur)
	}
	out = out + fmt.Sprintf("SMTProof Root: %x\n****\n", cur)
	return out
}

func (self *Proof) String() string {
	out := fmt.Sprintf("{\"token\":\"%x\",\"proofBits\":\"%x\",\"proof\":[", self.key, self.proofBits)
	for i, p := range self.proof {
		if i > 0 {
			out = out + ","
		}
		out = out + fmt.Sprintf("\"0x%x\"", p)
	}
	out = out + "]}"
	return out
}

func (p *Proof) Bytes() (out []byte) {
	out = append(out, UIntToByte(p.proofBits)...)
	for _, h := range p.proof {
		out = append(out, h...)
	}
	return out
}

func ToProof(index uint64, proofBytes []byte) (Proof, error) {
	var p Proof
	if len(proofBytes)%32 != 8 {
		return p, fmt.Errorf("Invalid proofBytes Length")
	}
	var pbits, psegs []byte
	pbits, psegs = proofBytes[:8], proofBytes[8:]
	p.key = UIntToByte(index)
	p.proofBits = Bytes32ToUint64(pbits)
	p.proof = proofSplit(psegs)
	return p, nil
}

func proofSplit(segments []byte) [][]byte {
	var proof []byte
	proofs := make([][]byte, 0, len(segments)/32+1)
	for len(segments) >= 32 {
		proof, segments = segments[:32], segments[32:]
		proofs = append(proofs, proof)
	}
	if len(segments) > 0 {
		proofs = append(proofs, segments[:len(segments)])
	}
	return proofs
}
