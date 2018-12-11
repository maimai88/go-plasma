// Copyright 2018 Wolk Inc.
// This file is part of the Wolk go-plasma library.

package merkletree

import (
	"fmt"
	"testing"
)

func TestMerkleTree(t *testing.T) {
	nitems := int64(100)
	o := make([][]byte, nitems)
	for x := int64(0); x < nitems; x++ {
		o[x] = Computehash([]byte(fmt.Sprintf("Val%d", x)))
		fmt.Printf("[%d] v(%s) keccak %x\n", x, []byte(fmt.Sprintf("Val%d", x)), o[x])

	}

	index := uint64(31)
	fmt.Printf("value: %x\n", o[index])

	// build merkle tree
	mtree := Merkelize(o)
	root := mtree[1]
	fmt.Printf("root: %x\n", root)

	// generate merkle proof
	b, err := Mk_branch(mtree, index)
	if err != nil {
		t.Fatalf("mk_branch: %v\n", err)
	}
	for i, x := range b {
		fmt.Printf("%d %x\n", i, x)
	}

	roothash, mkproof, err := GenProof(mtree, uint64(index))
	if err != nil {
		t.Fatalf("err: %v\n", err)
	}

	p, _ := ToProof(mkproof.Proof, mkproof.Index)
	fmt.Printf("[GenProof] root %x proof %x, ind %d\n[Proof]%s\n", roothash, mkproof.Proof, mkproof.Index, p.String())

	isValid, merkleroot, err := p.Verify(roothash)
	if err != nil {
		t.Fatalf("err: %v\n", err)
	} else {
		fmt.Printf("[Verify] isValid: %v merkleroot: %x\n", isValid, merkleroot)
	}

	merkleroot, proofstr, err := p.PrintProof()
	if err != nil {
		t.Fatalf("err: %v\n", err)
	} else {
		fmt.Printf("[Verify] merkleroot: %x\nproofstr:%s\n", merkleroot, proofstr)
	}

	//verify merkle proof
	res, err := Verify_branch(root, uint64(index), b)
	if err != nil {
		t.Fatalf("Merkle tree failure %v", err)
	} else {
		fmt.Printf("Merkle tree works %x\n", res)
	}
}
