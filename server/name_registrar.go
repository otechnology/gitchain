package server

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/env"
	"github.com/gitchain/gitchain/repository"
	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/util"
)

const RESERVATION_CONFIRMATIONS_REQUIRED = 3

func isValidReservation(reservation *transaction.Envelope, alloc *transaction.Envelope) bool {
	allocT := alloc.Transaction.(*transaction.NameAllocation)
	switch reservation.Transaction.(type) {
	case *transaction.NameReservation:
		tx1 := reservation.Transaction.(*transaction.NameReservation)
		return bytes.Compare(util.SHA256(append(allocT.Rand, []byte(allocT.Name)...)), tx1.Hashed) == 0
	default:
		return false
	}
}

func NameRegistrar() {
	ch := make(chan block.Block)
	router.PermanentSubscribe("/block", ch)
loop:
	select {
	case blk := <-ch:
		for i := range blk.Transactions {
			tx0 := blk.Transactions[i]
			tx := tx0.Transaction
			switch tx.(type) {
			case *transaction.NameAllocation:
				tx1 := tx.(*transaction.NameAllocation)
				// 1. find the reservation
				// 1.1. check if it was done with this server and there's a reference
				//      in scrap records
				reservation, err := env.DB.GetScrap(util.SHA256(append(tx1.Rand, []byte(tx1.Name)...)))
				var reservationTx *transaction.NameReservation
				if reservation == nil || err != nil {
					// 1.2 no scrap found, so try searching throughout database
					curBlock, err := env.DB.GetLastBlock()
					if err != nil {
						log.Printf("can't find last block during name allocation attempt")
						break
					}
					for curBlock != nil {
						for i := range curBlock.Transactions {
							if isValidReservation(curBlock.Transactions[i], tx0) {
								reservationTx = curBlock.Transactions[i].Transaction.(*transaction.NameReservation)
							}
						}

						h := curBlock.PreviousBlockHash
						curBlock, err = env.DB.GetBlock(h)
						if err != nil {
							log.Printf("can't find block %s during name allocation attempt", hex.EncodeToString(h))
							break
						}
					}

				} else {
					blk, err := env.DB.GetTransactionBlock(reservation)
					if err != nil {
						log.Printf("can't find block for name reservation %s", hex.EncodeToString(reservationTx.Hash()))
						break
					}
					for i := range blk.Transactions {
						if isValidReservation(blk.Transactions[i], tx0) {
							reservationTx = blk.Transactions[i].Transaction.(*transaction.NameReservation)
						}
					}

				}

				if reservationTx == nil {
					log.Printf("can't find corresponding name reservation for allocation %s", hex.EncodeToString(tx.Hash()))
					break
				}

				// 2. verify its maturity
				confirmations, err := env.DB.GetTransactionConfirmations(reservationTx.Hash())
				if err != nil {
					log.Printf("can't compute number of confirmations for reservation %s", hex.EncodeToString(reservationTx.Hash()))
					break
				}

				if confirmations >= RESERVATION_CONFIRMATIONS_REQUIRED {
					// this reservation is confirmed
					env.DB.PutRepository(repository.NewRepository(tx1.Name, repository.PENDING, tx1.Hash()))
				} else {
					// this allocation is wasted as the distance is not long enough
				}

			default:
				// ignore all other transactions
			}
		}
	}
	goto loop
}
