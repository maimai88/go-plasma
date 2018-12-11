// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package smt

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestCheckProof(t *testing.T) {
	var proof Proof
	proof.key = common.Hex2Bytes("79e4453dcbc77b29")
	fmt.Printf("TOKENID: %x\n", proof.key)
	proof.proofBits = uint64(0xc800000000000000)
	proof.proof = make([][]byte, 3)

	proof.proof[0] = common.Hex2Bytes("a5d59db538d26bd26e86b7fab2d688f8c03ab9d0dbf1adf2ef9bfa82de04b82b")
	proof.proof[1] = common.Hex2Bytes("49b4e065d6289c39dd4bb46545fd87a65edc5b9f9c8cc2fc6dfe9dc23b43d5a4")
	proof.proof[2] = common.Hex2Bytes("e8512edfdb95ea0eba5bdf718b981b3e845526b5d3ce2c463bc927cd5ad79a67")
	v := common.Hex2Bytes("7f2867b83f19a1443f67910d3f999a0385bbe50bf61c0df3795fbf23c081dd44")
	root := common.Hex2Bytes("ab06ee97217a525d229fe2f0ba129834b8a83742ae176b4987c5fdb95dc58797")
	fmt.Printf("TokenId: %v\n", proof.key)
	fmt.Printf("Proof Bytes: %x\n", proof.ProofBytes())
	fmt.Printf("Proof: %s\n", proof.String())
	if proof.Verify(v, root, true) {
		fmt.Printf("CheckProof pass\n")
	} else {
		fmt.Printf("CheckProof fail\n")
	}

	var p1, p2 Proof

	p1Leaf := common.Hex2Bytes("4fdcebb3247a9a715e416e68439e563e8faf57c804642441b93724d5b4fe0878")
	p1Root := common.Hex2Bytes("f361d4563fec05b5262f16c96aa062924256f61bd7482213ae23bf8bb2ad2e69")
	p1, _ = ToProof(uint64(0x9af84bc1208918b), common.Hex2Bytes("e000000000000000fb0d81010243cb5171ab9e619ca2a996a2f6eb2505a80b2e0252349c0f8d09105a581781cccb429b0de65eb3866f36039f615688ac29da96d9e12da69edabf97d77bd62537b7ed25202ba195360d56e3c8021109df6646f0f1cab6a6e130801a"))
	r1 := p1.Verify(p1Leaf, p1Root, false)
	fmt.Printf("result p1: %v\n", r1)
	fmt.Printf("p1: %+v leaf: %x root: %x\n", p1.String(), p1Leaf, p1Root)

	p2Leaf := common.Hex2Bytes("b2c958036da87b5e289cc04cd8fb40341f78f43485445687d74be69395d11dc7")
	p2Root := common.Hex2Bytes("a35a6b6586809dfc915a23b3b289c67b05f23a82fee5a5f080c56ddb9103ef75")
	p2, _ = ToProof(uint64(0x69eb463bc4f6b2df), common.Hex2Bytes("0000000000000000"))
	r2 := p2.Verify(p2Leaf, p2Root, false)

	fmt.Printf("p2: %+v leaf: %x root: %x\n", p2.String(), p2Leaf, p2Root)
	fmt.Printf("result p2: %v\n", r2)
}
