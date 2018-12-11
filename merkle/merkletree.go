// Copyright 2018 Wolk Inc.
// This file is part of the Wolk go-plasma library.

package merkletree

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

type Proof struct {
	Index uint64
	Proof []byte
}

func is_a_power_of_2(x uint64) bool {
	if x == 1 {
		return true
	}
	if x%2 == 1 {
		return false
	}
	return is_a_power_of_2(x / 2)
}

func Computehash(data ...[]byte) []byte {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func verifyLength(rawProof []byte) bool {
	if len(rawProof)%32 != 0 || len(rawProof) < 32 {
		return false
	}
	return true
}

func Merkelize(L [][]byte) [][]byte {
	for is_a_power_of_2(uint64(len(L))) == false {
		L = append(L, []byte(""))
	}
	LH := make([][]byte, len(L))
	for i, v := range L {
		LH[i] = v
	}
	nodes := make([][]byte, len(L))
	nodes = append(nodes, LH...)
	for i := len(L) - 1; i >= 0; i-- {
		nodes[i] = Computehash(append(nodes[i*2], nodes[i*2+1]...))
	}
	return nodes
}

func MerkelRoot(tree [][]byte) (merkelroot []byte) {
	return tree[1]
}

func Mk_branch(tree [][]byte, index uint64) (o [][]byte, err error) {
	if index > uint64(len(tree)/2) {
		return o, fmt.Errorf("Invalid idx")
	}
	index += uint64(len(tree)) / 2
	o = make([][]byte, 1)
	o[0] = tree[index]
	for index > 1 {
		o = append(o, tree[index^1])
		index = index / 2
	}
	return o, nil
}

func Verify_branch_int(root []byte, index uint64, proof [][]byte) (res *big.Int, err error) {
	res_byte, err := Verify_branch(root, index, proof)
	if err != nil {
		return res, err
	}
	res = common.BytesToHash(res_byte).Big()
	return res, nil
}

func Verify_branch(root []byte, index uint64, proof [][]byte) (res []byte, err error) {
	q := 1 << uint64(len(proof))
	index += uint64(q)
	v := proof[0]
	for _, p := range proof[1:] {
		if index%2 > 0 {
			v = Computehash(append(p, v...))
		} else {
			v = Computehash(append(v, p...))
		}
		index = index / 2
	}
	if bytes.Compare(v, root) != 0 {
		return res, fmt.Errorf("Mismatch root, got:[%x] expected:[%x]", v, root)
	}
	return proof[0], nil
}

func GenProof(tree [][]byte, ind uint64) (merkelroot []byte, p Proof, err error) {
	treelen := uint64(len(tree) / 2)
	if ind > treelen {
		return merkelroot, p, fmt.Errorf("Invalid idx")
	}
	p.Index = ind
	ind += treelen
	p.Proof = append(p.Proof, tree[ind]...)
	for ind > 1 {
		p.Proof = append(p.Proof, tree[ind^1]...)
		ind = ind / 2
	}
	return tree[1], p, nil
}

func (p *Proof) GetRoot() (merkleroot []byte, err error) {
	ind := p.Index
	mkproof := p.Proof
	if !verifyLength(mkproof) {
		return merkleroot, fmt.Errorf("Invalid proofBytes Length: %d", len(mkproof))
	}

	merkleroot = append(merkleroot, mkproof[0:32]...)
	merklepath := merkleroot
	for depth := 1; depth < len(mkproof)/32; depth++ {
		rhash := make([]byte, 32)
		copy(rhash, mkproof[depth*32:(depth+1)*32])
		if ind%2 > 0 {
			merkleroot = Computehash(append(rhash, merkleroot...))
		} else {
			merkleroot = Computehash(append(merkleroot, rhash...))
		}
		ind = ind / 2
		merklepath = append(merklepath, merkleroot...)
	}
	return merkleroot, nil
}

func (p *Proof) Verify(expectedMerkleRoot []byte) (isValid bool, merkleroot []byte, err error) {
	if merkleroot, err = p.GetRoot(); err != nil {
		return false, merkleroot, err
	}
	if bytes.Compare(expectedMerkleRoot, merkleroot) != 0 {
		return false, merkleroot, nil
	} else {
		return true, merkleroot, nil
	}
}

func (p *Proof) PrintProof() (merkleroot []byte, proofstr string, err error) {
	mkproof := p.Proof
	ind := p.Index
	if !verifyLength(mkproof) {
		return merkleroot, proofstr, fmt.Errorf("Invalid proofBytes Length: %d", len(mkproof))
	}
	merkleroot = append(merkleroot, mkproof[0:32]...)
	merklepath := merkleroot
	out := fmt.Sprintf("****\nBlockProof \nH0       %x (Leaf) \n", merkleroot)
	for depth := 1; depth < len(mkproof)/32; depth++ {
		rhash := make([]byte, 32)
		copy(rhash, mkproof[depth*32:(depth+1)*32])
		if ind%2 > 0 {
			out = out + fmt.Sprintf("H%d [*,P] H(%x,%x)", depth, rhash, merkleroot)
			merkleroot = Computehash(append(rhash, merkleroot...))
		} else {
			out = out + fmt.Sprintf("H%d [P,*] H(%x,%x)", depth, merkleroot, rhash)
			merkleroot = Computehash(append(merkleroot, rhash...))
		}
		ind = ind / 2
		out = out + fmt.Sprintf(" => %x\n", merkleroot)
		merklepath = append(merklepath, merkleroot...)
	}
	proofstr = proofstr + fmt.Sprintf("BlockRoot: %x\n****\n", merkleroot)
	return merkleroot, out, nil
}

func ToProof(mkProof []byte, ind uint64) (Proof, error) {
	var p Proof
	if !verifyLength(mkProof) {
		return p, fmt.Errorf("Invalid proofBytes Length: %d", len(mkProof))
	}
	p = Proof{Index: ind, Proof: mkProof}
	return p, nil
}

func (p *Proof) Leaf() (leaf []byte) {
	internalProof := p.Proof
	if verifyLength(internalProof) {
		leaf = append(leaf, internalProof[0:32]...)
	}
	return leaf
}

func (p *Proof) Root() (root []byte) {
	merkleroot, err := p.GetRoot()
	if err != nil {
		return root
	}
	return merkleroot
}

func (p *Proof) String() string {
	merkleroot, err := p.GetRoot()
	if err != nil {
		return "{}"
	}
	out := fmt.Sprintf("{\"Index\":\"%d\",\"Leaf\":\"%x\",\"Root\":\"%x\",\"Proof\":[", p.Index, p.Leaf(), merkleroot)
	for prev := 0; prev < len(p.Proof); prev += 32 {
		if prev > 0 {
			out = out + ","
		}
		out = out + common.Bytes2Hex(p.Proof[prev:prev+32])
	}
	out = out + "]}"
	return out
}
