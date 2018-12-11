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

func (p *Proof) Verify(leaf []byte, root []byte, verbose bool) bool {
	merkleroot, err := p.GetRoot(leaf)
	if err != nil {
		if verbose {
			fmt.Printf("Err: %v", err)
		}
		return false
	}
	res := bytes.Compare(merkleroot, root) == 0
	if verbose {
		if res {
			fmt.Printf(" CheckProof success (root match: %x)\n", merkleroot)
		} else {
			fmt.Printf(" CheckProof FAILURE (expected root [%x] does NOT match actual root: %x)\n", root, merkleroot)
		}
	}
	return res
}

func (p *Proof) GetRoot(leaf []byte) (merkleroot []byte, err error) {
	cur := leaf
	d := 0
	for i := uint64(0); i < 64; i++ {
		if (uint64(1<<i) & p.proofBits) > 0 {
			if d >= len(p.proof) {
				return merkleroot, fmt.Errorf("Invalid Non-default depth at %d", d)
			}
			if byte(0x01<<(i%8))&byte(p.key[(TreeDepth-1-i)/8]) > 0 {
				cur = Computehash(p.proof[d], cur)
			} else {
				cur = Computehash(cur, p.proof[d])
			}
			d++
		} else {
			if byte(0x01<<(i%8))&byte(p.key[(TreeDepth-1-i)/8]) > 0 {
				cur = Computehash(GlobalDefaultHashes[i], cur)
			} else {
				cur = Computehash(cur, GlobalDefaultHashes[i])
			}
		}
	}
	return cur, nil
}

func (p *Proof) Root(leaf []byte) (root []byte) {
	merkleroot, err := p.GetRoot(leaf)
	if err != nil {
		return root
	}
	return merkleroot
}

func (p *Proof) String() string {
	out := fmt.Sprintf("{\"token\":\"%x\",\"proofBits\":\"%x\",\"proof\":[", p.key, p.proofBits)
	for i, seg := range p.proof {
		if i > 0 {
			out = out + ","
		}
		out = out + fmt.Sprintf("\"0x%x\"", seg)
	}
	out = out + "]}"
	return out
}

func (p *Proof) ProofBytes() (out []byte) {
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
