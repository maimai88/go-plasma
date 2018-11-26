// Copyright 2018 Wolk Inc.
// This file is part of the Wolk go-plasma library.
package smt

import (
	"encoding/binary"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

func Computehash(data ...[]byte) []byte {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func ComputeDefaultHashes() (defaultHashes [TreeDepth][]byte) {
	empty := make([]byte, 0)
	defaultHashes[0] = Computehash(empty)
	for level := 1; level < TreeDepth; level++ {
		defaultHashes[level] = Computehash(defaultHashes[level-1], defaultHashes[level-1])
	}
	return defaultHashes
}

// helper stuff here for a while
func IntToByte(i int64) (k []byte) {
	k = make([]byte, 8)
	binary.BigEndian.PutUint64(k, uint64(i))
	return k
}

func UIntToByte(i uint64) (k []byte) {
	k = make([]byte, 8)
	binary.BigEndian.PutUint64(k, uint64(i))
	return k
}

func Uint64ToBytes32(i uint64) (k []byte) {
	k = make([]byte, 32)
	binary.BigEndian.PutUint64(k[24:32], uint64(i))
	return k
}

func Bytes32ToUint64(k []byte) (out uint64) {
	h := k[0:8]
	return binary.BigEndian.Uint64(h)
}
