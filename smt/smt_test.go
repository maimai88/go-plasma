// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package smt

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/wolkdb/go-plasma/deep"
)

// Baseline Output:
// 100000 key Memory write: 987.837926ms
// 100000 key LevelDB write: 19.010639896s
// 100000 key LevelDB 1 BATCH: 6.542241888s
// 100000 key LevelDB 10 BATCHes: 7.647330818s
// 100000 key LevelDB 100 BATCHes: 13.585051477s
func TestBaseline(t *testing.T) {
	t.SkipNow()
	nkeys := uint64(100000)

	// write 100K keys to memory
	rand.Seed(time.Now().UnixNano())
	st := time.Now()
	kv := make(map[uint64][]byte)
	for i := uint64(0); i < nkeys; i++ {
		k := Bytes32ToUint64(deep.Keccak256(Uint64ToBytes32(i % 1000000)))
		v := make([]byte, 4096)
		rand.Read(v)
		kv[k] = v
	}
	fmt.Printf("%d key Memory write: %s\n", nkeys, time.Since(st))

	// write 100K keys to leveldb
	path := fmt.Sprintf("/tmp/baseline%d", int32(time.Now().Unix()))
	ldb, err := leveldb.OpenFile(path, nil)
	if err != nil {
		t.Fatalf("OpenFile: %v ", err)
	}
	/*
		st = time.Now()
		for k, v := range kv {
			key := []byte(fmt.Sprintf("%d", k))
			err = ldb.Put(key, v, nil)
			if err != nil {
				t.Fatalf("ldb.Put: %v ", err)
			}
		}
		fmt.Printf("%d key LevelDB write: %s\n", nkeys, time.Since(st))
	*/
	/*
		st = time.Now()
		batch := new(leveldb.Batch)
		for k, v := range kv {
			key := []byte(fmt.Sprintf("%d", k))
			batch.Put(key, v)
		}
		err = ldb.Write(batch, nil)
		if err != nil {
			t.Fatalf("Write: %v ", err)
		}
		fmt.Printf("%d key LevelDB 1 BATCH: %s\n", nkeys, time.Since(st))
	*/
	/*	st = time.Now()
		batch := new(leveldb.Batch)
		n := 0
		for k, v := range kv {
			key := []byte(fmt.Sprintf("%d", k))
			batch.Put(key, v)
			if n == 1000 {
				err = ldb.Write(batch, nil)
				if err != nil {
					t.Fatalf("OpenFile: %v ", err)
				}
				batch = new(leveldb.Batch)
				n = 0
			} else {
				n++
			}
		}
		fmt.Printf("%d key LevelDB 100 BATCHes: %s\n", nkeys, time.Since(st))
	*/
	st = time.Now()
	batch := new(leveldb.Batch)
	n := 0
	for k, v := range kv {
		key := []byte(fmt.Sprintf("%d", k))
		batch.Put(key, v)
		if n == 10000 {
			err = ldb.Write(batch, nil)
			if err != nil {
				t.Fatalf("OpenFile: %v ", err)
			}
			batch = new(leveldb.Batch)
			n = 0
		} else {
			n++
		}
	}
	fmt.Printf("%d key LevelDB 10 BATCHes: %s\n", nkeys, time.Since(st))

}

/*
func TestSequence(t *testing.T) {

	// setup plasma store
	pcs, err := plasmachain.NewPlasmaChunkstore(plasmachain.DefaultChunkstorePath)
	if err != nil {
		t.Fatalf("[smt_test:NewCloudstore]%v", err)
	}
	defer pcs.Close()

	smt0 := smt.NewSparseMerkleTree(pcs)
	nkeys := uint64(2)
	kv := make(map[uint64]common.Hash)
	for i := uint64(0); i < nkeys; i++ {

		k := smt.UIntToByte(i)
		v := smt.Keccak256([]byte(fmt.Sprintf("value%d", i)))
		kv[i] = common.BytesToHash(v)
		err = smt0.Insert(k, v, 0, 0)
		if err != nil {
			t.Fatalf("SetKey: %v\n", err)
		}
	}
	smt0.Flush()
	smt0.Dump()
	chunkHash := smt0.ChunkHash()
	merkleRoot := smt0.MerkleRoot()
	fmt.Printf("Generated:  Hash: %x Merkle Root: %x\n", chunkHash, merkleRoot)
	passes := 0
	smt0 = smt.NewSparseMerkleTree(pcs)
	smt0.Init(chunkHash)
	for i := uint64(0); i < nkeys; i++ {
		k := i
		v1, found, proof, storageBytes, prevBlock, err := smt0.Get(smt.UIntToByte(k))
		smt0.Flush()
		// smt0.Dump()
		if err != nil {
			fmt.Printf("err not found %x %v \n", k, err)
		} else if found {
			if bytes.Compare(kv[k].Bytes(), v1) == 0 {
				checkproof := proof.Check(v1, merkleRoot.Bytes(), smt0.DefaultHashes, false)
				if checkproof {
					passes++
				} else {
					fmt.Printf("k:%x v:%x storageBytes:%d prevBlock: %d ", k, v1, storageBytes, prevBlock)
					t.Fatalf("CHECK PROOF ==> FAILURE\n")
				}
			} else {
				t.Fatalf("k:%x v:%x sb:%d kv[k]:%x INCORRECT\n", k, v1, storageBytes, kv[k])
			}
		} else {
			fmt.Printf("k:%x not found \n", k)
		}
	}
	fmt.Printf("%d/%d keys PASSED\n", passes, nkeys)

}
*/
type Config struct {
	PlasmaAddr string
	PlasmaPort uint64

	CloudstoreAddr string
	CloudstorePort uint64

	DataDir        string
	RemoteDisabled bool
}

// DefaultConfig contains default settings for use on the Ethereum main net.
var DefaultConfig = &Config{
	PlasmaAddr:     "localhost",
	PlasmaPort:     32003,
	CloudstoreAddr: "localhost",
	CloudstorePort: 9900,
	DataDir:        "/tmp/remtest",
	RemoteDisabled: true,
}

func (c *Config) GetPlasmaAddr() string     { return c.PlasmaAddr }
func (c *Config) GetPlasmaPort() uint64     { return c.PlasmaPort }
func (c *Config) GetCloudstoreAddr() string { return c.CloudstoreAddr }
func (c *Config) GetCloudstorePort() uint64 { return c.CloudstorePort }
func (c *Config) GetDataDir() string        { return c.DataDir }
func (c *Config) IsLocalMode() bool         { return c.RemoteDisabled }

func newRemoteStorage(config *Config) (rs *deep.RemoteStorage, err error) {
	log.Root().SetHandler(log.CallerFileHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true)))))
	_blockchainId := uint64(42)
	_chainType := "sql"
	operatorPrivateKey := "6545ddd10c1e0d6693ba62dec711a2d2973124ae0374d822f845d322fb251645"
	operatorKey, _ := crypto.HexToECDSA(operatorPrivateKey)
	rs, err = deep.NewRemoteStorage(_blockchainId, config, _chainType, operatorKey)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func TestSMT(t *testing.T) {

	pcs, err := newRemoteStorage(DefaultConfig)
	if err != nil {
		t.Fatalf("[smt_test:NewCloudstore]%v", err)
	}
	defer pcs.Close()

	smt0 := NewSparseMerkleTree(pcs)
	smt0.Flush()
	str := fmt.Sprintf("%x", smt0.MerkleRoot())
	// an empty tree gets you default hashes of level 65
	if strings.Compare(str, "b992a50058a2812b0fc4fe1bbbfb3d8ffd476fb89391408212e00a7019e10eff") != 0 {
		t.Fatalf("Empty tree / Default Hash level 65 mismatch")
	}

	nkeys := uint64(2)
	nversions := uint64(1)
	chunkHash := make(map[uint64]common.Hash)
	merkleRoot := make(map[uint64]common.Hash)
	kv := make(map[uint64]map[uint64]common.Hash)
	for ver := uint64(0); ver < nversions; ver++ {
		kv[ver] = make(map[uint64]common.Hash)
		inserts := 0
		for i := uint64(0); i < nkeys; i++ {
			storageBytesNew := uint64(3)
			k := Bytes32ToUint64(deep.Keccak256(Uint64ToBytes32(i % 1000000)))
			v := deep.Keccak256([]byte(fmt.Sprintf("%d%d", i, ver)))
			kv[ver][k] = common.BytesToHash(v)
			prevBlock := ver
			err = smt0.Insert(UIntToByte(k), v, storageBytesNew, prevBlock)
			if err != nil {
				t.Fatalf("SetKey: %v\n", err)
			} else {
				inserts++
				if ver > 0 && i > ver*ver {
					i = nkeys
				}
			}
		}
		st := time.Now()
		smt0.Flush()
		fmt.Printf("TIMING of %d (%d inserts): %s\n", ver, inserts, time.Since(st))
		chunkHash[ver] = smt0.ChunkHash()
		merkleRoot[ver] = smt0.MerkleRoot()
		fmt.Printf("Generated: Version %d Hash: %x Merkle Root: %x\n", ver, chunkHash[ver], merkleRoot[ver])
	}

	fmt.Println("----------------------------------------------------------------------------------------")
	time.Sleep(1 * time.Second)

	for ver := uint64(0); ver < nversions; ver++ {
		smt0 = NewSparseMerkleTree(pcs)
		smt0.Init(chunkHash[ver])
		fmt.Printf("BEFORE PROOF GEN: chunkHash: %x merkleroot: %x\n", smt0.ChunkHash(), smt0.MerkleRoot())
		passes := 0
		st := time.Now()
		for i := uint64(0); i < nkeys; i++ {
			k := Bytes32ToUint64(deep.Keccak256(Uint64ToBytes32(i % 10000)))
			v1, found, proof, _, _, err := smt0.Get(UIntToByte(k))
			fmt.Printf("AFTER PROOF GEN: chunkHash: %x merkleroot: %x PROOF: %v\n", smt0.ChunkHash(), smt0.MerkleRoot(), proof)
			if err != nil {
				fmt.Printf("err not found %x %v \n", k, err)
			} else if found {
				if bytes.Compare(kv[ver][k].Bytes(), v1) == 0 {
					checkproof := proof.Check(v1, merkleRoot[ver].Bytes(), false)
					if checkproof {
						passes++
					} else {
						fmt.Printf("k:%x v:%x ver %d -- ", k, v1, ver)
						t.Fatalf("CHECK PROOF ==> FAILURE\n")
					}
				} else {
					t.Fatalf("k:%x v:%x kv[k]:%x INCORRECT\n", k, v1, kv[k])
				}
			} else {
				fmt.Printf("k:%x not found \n", k)
			}
		}
		fmt.Printf("Version %d  -- %d/%d keys PASSED [%s]\n", ver, passes, nkeys, time.Since(st))
	}
}
