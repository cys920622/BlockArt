package test

import (
	"../blockchain"
	"crypto/elliptic"
	"crypto/ecdsa"
	"crypto/rand"

	"encoding/json"
	"crypto/md5"
	"encoding/hex"
	"testing"
	"fmt"
	"reflect"
)

const GENESIS_BLOCK_HASH = "83218ac34c1834c26781fe4bde918ee4"
const RANDOM_NONCE = 1 // just putting a random nonce in the block since we are not testing it
const SVG_OP_ONE = "<path d=\"M 0 0 L 20 20\" stroke=\"red\" fill=\"transparent\"/>"
const SVG_OP_TWO = "<path d=\"M 30 30 L 40 40\" stroke=\"red\" fill=\"transparent\"/>"
const SVG_OP_THREE = "<path d=\"M 50 50 L 60 60\" stroke=\"red\" fill=\"transparent\"/>"

var p256 = elliptic.P256()
var minerOnePrivateKey, _ = ecdsa.GenerateKey(p256, rand.Reader)
var minerOnePublicKey = minerOnePrivateKey.PublicKey
var minerTwoPrivateKey, _ = ecdsa.GenerateKey(p256, rand.Reader)
var minerTwoPublicKey = minerTwoPrivateKey.PublicKey

// A mock block chain used to test traverse functions
// mimics a chain generated by two miners
// The chain will have the following structure: [(m1) means mined by miner1]
// NO OP BLOCK (m1) <- NO OP BLOCK (m2) <- OP BLOCK CONTAINING ONE SHAPE BY M1 AND ONE SHAPE BY M2 (m1) <- OP BLOCK CONTAINING ONE SHAPE MADE BY M2 (m2)
var noOPBlockMinerOne = blockchain.Block {
	BlockNum: 1,
	PrevHash: GENESIS_BLOCK_HASH,
	OpRecords: make(map[string]*blockchain.OpRecord),
	MinerPubKey: &minerOnePublicKey,
	Nonce: RANDOM_NONCE,
}
var blockOneHash = computeBlockHash(noOPBlockMinerOne)

var noOPBlockMinerTwo = blockchain.Block {
	BlockNum: 2,
	PrevHash: blockOneHash,
	OpRecords: make(map[string]*blockchain.OpRecord),
	MinerPubKey: &minerTwoPublicKey,
	Nonce: RANDOM_NONCE,
}
var blockTwoHash = computeBlockHash(noOPBlockMinerTwo)

// Generate OP-SIG
var svgOPOne = []byte(SVG_OP_ONE)
var r1, s1, _ = ecdsa.Sign(rand.Reader, minerOnePrivateKey, svgOPOne)
var svgOpTwo = []byte(SVG_OP_TWO)
var r2, s2, _ = ecdsa.Sign(rand.Reader, minerTwoPrivateKey, svgOpTwo)
var svgOpThree = []byte(SVG_OP_THREE)
var r3, s3, _ = ecdsa.Sign(rand.Reader, minerTwoPrivateKey, svgOpThree)

var minerOneOpRecordOne = blockchain.OpRecord {
	Op: SVG_OP_ONE,
	InkUsed: 20,
	OpSigR: r1,
	OpSigS: s1,
	AuthorPubKey: minerOnePublicKey,
}
var opRecOneHash = computeOpRecordHash(minerOneOpRecordOne)

var minerOneOpRecordTwo = blockchain.OpRecord {
	Op: SVG_OP_TWO,
	InkUsed: 10,
	OpSigR: r2,
	OpSigS: s2,
	AuthorPubKey: minerTwoPublicKey,
}
var opRecTwoHash = computeOpRecordHash(minerOneOpRecordTwo)

var minerTwoOpRecord = blockchain.OpRecord {
	Op: SVG_OP_THREE,
	InkUsed: 10,
	OpSigR: r3,
	OpSigS: s3,
	AuthorPubKey: minerTwoPublicKey,
}
var opRecThreeHash = computeOpRecordHash(minerTwoOpRecord)

var opRecordsBlockThree = make(map[string]*blockchain.OpRecord)
var opBlockMinerOne = blockchain.Block {
	BlockNum: 3,
	PrevHash: blockTwoHash,
	OpRecords: opRecordsBlockThree,
	MinerPubKey: &minerOnePublicKey,
	Nonce: RANDOM_NONCE,
}
var blockThreeHash = computeBlockHash(opBlockMinerOne)

var opRecordsBlockFour = make(map[string]*blockchain.OpRecord)
var opBlockMinerTwo = blockchain.Block {
	BlockNum: 4,
	PrevHash: blockThreeHash,
	OpRecords: opRecordsBlockFour,
	MinerPubKey: &minerTwoPublicKey,
	Nonce: RANDOM_NONCE,
}
var blockFourHash = computeBlockHash(opBlockMinerTwo)

var blockChain blockchain.BlockChain

func setUpBlockChain() {
	opRecordsBlockThree[opRecOneHash] = &minerOneOpRecordOne
	opRecordsBlockThree[opRecTwoHash] = &minerOneOpRecordTwo
	opRecordsBlockFour[opRecThreeHash] = &minerTwoOpRecord

	blocks := make(map[string]*blockchain.Block)
	blocks[blockOneHash] = &noOPBlockMinerOne
	blocks[blockTwoHash] = &noOPBlockMinerTwo
	blocks[blockThreeHash] = &opBlockMinerOne
	blocks[blockFourHash] = &opBlockMinerTwo

	blockChain = blockchain.BlockChain {
		Blocks: blocks,
		NewestHash: blockFourHash,
	}

	// Traverses the chain and print out content of each block in the chain
	newestHash := blockChain.NewestHash
	for blockHash := newestHash; blockHash != GENESIS_BLOCK_HASH; blockHash = blockChain.Blocks[blockHash].PrevHash {
		block := blockChain.Blocks[blockHash]
		fmt.Printf("Block Num: %d \nPrevHash: %s \nMinerPubKey: %+v\n", block.BlockNum, block.PrevHash, block.MinerPubKey.X)
		if len(block.OpRecords) == 0 {
			fmt.Printf("Block %d is a no op block\n\n", block.BlockNum)
		} else {
			fmt.Printf("Block %d contain the the following operations: \n", block.BlockNum)
			for k, _ := range block.OpRecords {
				fmt.Println(block.OpRecords[k].Op)
				if reflect.DeepEqual(block.OpRecords[k].AuthorPubKey, minerOnePublicKey) {
					fmt.Println("The above Operation was done by miner 1")
				} else {
					fmt.Println("The above Operation was done by miner 2")
				}
			}
			fmt.Println("")
		}
	}

}

func TestGetInkTraversal(t *testing.T) {
	setUpBlockChain()
	t.Error("Fail for now")
	// TODO: Add test for traversing the tree to get ink
}

func TestGetShapesTraversal(t *testing.T) {
	// setUpBlockChain()
	// TODO: Add test for traversing the tree to get all the shapes
}

func TestGetShapeTraversal(t *testing.T) {
	// setUpBlockChain()
	// TODO: Add test for traversing the tree to get a svg specified by shapeHash
}


// Compute the MD5 hash of a Block
func computeBlockHash(block blockchain.Block) string {
	bytes, _ := json.Marshal(block)
	hash := md5.New()
	hash.Write(bytes)
	return hex.EncodeToString(hash.Sum(nil))
}

// Compute the MD5 hash of a OpRecord
func computeOpRecordHash(opRecord blockchain.OpRecord) string {
	bytes, _ := json.Marshal(opRecord)
	hash := md5.New()
	hash.Write(bytes)
	return hex.EncodeToString(hash.Sum(nil))
}