package api

import (
	"encoding/hex"
	"net/http"

	"github.com/gitchain/gitchain/env"
)

type BlockService struct{}

type GetLastBlockArgs struct {
}

type GetLastBlockReply struct {
	Hash string
}

func (srv *BlockService) GetLastBlock(r *http.Request, args *GetLastBlockArgs, reply *GetLastBlockReply) error {
	block, err := env.DB.GetLastBlock()
	if err != nil {
		return err
	}
	reply.Hash = hex.EncodeToString(block.Hash())
	return nil
}

type GetBlockArgs struct {
	Hash string
}

type GetBlockReply struct {
	NextBlockHash     string
	PreviousBlockHash string
	MerkleRootHash    string
	Timestamp         int64
	Bits              uint32
	Nonce             uint32
	NumTransactions   int
}

func (srv *BlockService) GetBlock(r *http.Request, args *GetBlockArgs, reply *GetBlockReply) error {
	hash, err := hex.DecodeString(args.Hash)
	if err != nil {
		return err
	}
	block, err := env.DB.GetBlock(hash)
	if err != nil {
		return err
	}
	nextblock, err := env.DB.GetNextBlock(hash)
	if err != nil {
		return err
	}
	if nextblock != nil {
		reply.NextBlockHash = hex.EncodeToString(nextblock.Hash())
	}
	reply.PreviousBlockHash = hex.EncodeToString(block.PreviousBlockHash)
	reply.MerkleRootHash = hex.EncodeToString(block.MerkleRootHash)
	reply.Timestamp = block.Timestamp
	reply.Bits = block.Bits
	reply.Nonce = block.Nonce
	reply.NumTransactions = len(block.Transactions)
	return nil
}

type BlockTransactionsArgs struct {
	Hash string
}

type BlockTransactionsReply struct {
	Transactions []string
}

func (srv *BlockService) BlockTransactions(r *http.Request, args *BlockTransactionsArgs, reply *BlockTransactionsReply) error {
	hash, err := hex.DecodeString(args.Hash)
	if err != nil {
		return err
	}
	block, err := env.DB.GetBlock(hash)
	if err != nil {
		return err
	}
	for i := range block.Transactions {
		reply.Transactions = append(reply.Transactions, hex.EncodeToString(block.Transactions[i].Hash()))
	}
	return nil
}
