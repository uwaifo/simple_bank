package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {

	//Transaction Sore object
	store := NewStore(testDB)

	accountOne := createRandomAccount(t)
	accountTwo := createRandomAccount(t)

	//NOTICE
	fmt.Println(">> before:", accountOne.Balance, accountTwo.Balance)

	// numTrx represents the number of goroutines/transactions
	numTrx := 5
	// testAmount represents the amount to the transacted on
	testAmount := int64(10)

	// errorChan and resultChan are channels to handle errors and eventual results of eacch transaction respectively
	errorChan := make(chan error)
	resultChan := make(chan TransferTxResult)

	// Run numTrx number concurrent transfer transaction
	for i := 0; i < numTrx; i++ {

		//txName := fmt.Sprintf("tx %d", i+1)
		// Spin a goroutine to call TransferTx
		go func() {
			//for loging we append our won value for transaction tracking to the Background context
			//ctx := context.WithValue(context.Background(), txKey, txName)
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountId: accountOne.ID,
				ToAccountID:   accountTwo.ID,
				Amount:        testAmount,
			})
			// pipeline to channel
			errorChan <- err
			resultChan <- result

		}()

	}

	//check for results

	//

	existed := make(map[int]bool)
	for i := 0; i < numTrx; i++ {
		err := <-errorChan
		require.NoError(t, err)

		result := <-resultChan
		require.NotEmpty(t, result)

		//Check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, accountOne.ID, transfer.FromAccountID)
		require.Equal(t, accountTwo.ID, transfer.ToAccountID)
		require.Equal(t, testAmount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		//Check from entry
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, accountOne.ID, fromEntry.AccountID)
		require.Equal(t, -testAmount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		//Check to entry
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, accountTwo.ID, toEntry.AccountID)
		require.Equal(t, testAmount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		//TODO Check and update account ballance

		//Check the accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, accountOne.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, accountTwo.ID, toAccount.ID)

		//check both accounts balance
		//NOTICE
		fmt.Println(">> TX:", fromAccount.Balance, toAccount.Balance)

		diffOne := accountOne.Balance - fromAccount.Balance
		diffTwo := toAccount.Balance - accountTwo.Balance
		require.Equal(t, diffOne, diffTwo)
		require.True(t, diffOne > 0)
		require.True(t, diffOne%testAmount == 0) // testAmmount,

		k := int(diffOne / testAmount)
		require.True(t, k >= 1 && k <= numTrx)
		require.NotContains(t, existed, k)
		existed[k] = true

	}
	//check the final updated balance
	updatedAccountOne, err := testQueries.GetAccount(context.Background(), accountOne.ID)
	require.NoError(t, err)

	updatedAccountTwo, err := testQueries.GetAccount(context.Background(), accountTwo.ID)
	require.NoError(t, err)

	//NOTICE
	fmt.Println(">> after:", accountOne.Balance, accountTwo.Balance)

	require.Equal(t, accountOne.Balance-int64(numTrx)*testAmount, updatedAccountOne.Balance)
	require.Equal(t, accountTwo.Balance+int64(numTrx)*testAmount, updatedAccountTwo.Balance)

}

/*
func TestTransferTxDeadLock(t *testing.T) {

	store := NewStore(testDB)

	accountOne := createRandomAccount(t)
	accountTwo := createRandomAccount(t)

	//NOTICE
	fmt.Println(">> before:", accountOne.Balance, accountTwo.Balance)

	numTrx := 10
	testAmount := int64(10)

	errorChan := make(chan error)

	for i := 0; i < numTrx; i++ {

		fromAccountID := accountOne.ID
		toAccountID := accountTwo.ID

		if i%2 == 1 {
			fromAccountID = accountTwo.ID
			toAccountID = accountOne.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountId: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        testAmount,
			})
			errorChan <- err

		}()

	}

	//check for results

	for i := 0; i < numTrx; i++ {
		err := <-errorChan
		require.NoError(t, err)

	}
	//check the final updated balance
	updatedAccountOne, err := testQueries.GetAccount(context.Background(), accountOne.ID)
	require.NoError(t, err)

	updatedAccountTwo, err := testQueries.GetAccount(context.Background(), accountTwo.ID)
	require.NoError(t, err)

	//NOTICE
	fmt.Println(">> after:", accountOne.Balance, accountTwo.Balance)

	require.Equal(t, accountOne.Balance-int64(numTrx)*testAmount, updatedAccountOne.Balance)
	require.Equal(t, accountTwo.Balance+int64(numTrx)*testAmount, updatedAccountTwo.Balance)

}
*/
func TestTransferTxDeadlockTwo(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	n := 10
	amount := int64(10)
	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID

		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				//FromAccountID: fromAccountID,
				FromAccountId: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// check the final updated balance
	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)
}
