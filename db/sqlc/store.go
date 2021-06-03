package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides functionality to run multiple *Queries and thus giving us transactions
type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// execTx excutes a function withinn a database transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil) /*&sql.TxOptions{}*/
	if err != nil {
		return nil
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction err: %v, rollback err: %v", err, rollbackErr)
		}
	}
	return tx.Commit()

}

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountId int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is the result of the TransferTxParams transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

//
//var txKey = struct{}{}

// TransferTx performs money transfer
// Creates transfer records, entries and update thes ballance on the account etc
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {

	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		//
		//txName := ctx.Value(txKey)
		//fmt.Println(txName, "create transfer")

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountId,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		//
		//fmt.Println(txName, "create entry one")

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountId,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		//
		//fmt.Println(txName, "create entry two")

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		//TODO : GET ACCOINT -> UPDATE ITS BALANCE (Ensure to use LOCK)

		//acc1
		//fmt.Println(txName, "get account one for update")

		/*accountOne, err := q.GetAccountForUpdate(ctx, arg.FromAccountId)
		if err != nil {
			return err
		}*/

		result.FromAccount, err = q.AddAccountBallance(ctx, AddAccountBallanceParams{
			ID:     arg.FromAccountId,
			Amount: -arg.Amount,
		})
		if err != nil {
			return err
		}

		//acc2
		//fmt.Println(txName, "get account two for update")

		/*accountTwo, err := q.GetAccountForUpdate(ctx, arg.ToAccountID)
		if err != nil {
			return err
		}*/
		//fmt.Println(txName, "update account two")

		result.ToAccount, err = q.AddAccountBallance(ctx, AddAccountBallanceParams{
			ID:     arg.ToAccountID,
			Amount: arg.Amount,
		})
		if err != nil {
			return err
		}
		return nil
	})

	return result, err

}
