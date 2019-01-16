package ledger

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TransactionsSuite struct {
	suite.Suite
	ledgerEndpoint  string
	ledgerAuthToken string
}

func (ts *TransactionsSuite) SetupSuite() {
	ts.ledgerEndpoint = os.Getenv("LEDGER_ENDPOINT")
	ts.ledgerAuthToken = os.Getenv("LEDGER_AUTH_TOKEN")
}

func (ts *TransactionsSuite) TestCreateTransactionValid() {
	t := ts.T()
	ledgerApp := NewLedger(ts.ledgerEndpoint, ts.ledgerAuthToken)

	// Prepare and create a valid transaction
	txnID := NewUUID()
	txn := &Transaction{
		ID: txnID,
		Lines: []*TransactionLine{
			&TransactionLine{AccountID: "CX", Delta: 100},
			&TransactionLine{AccountID: "CY", Delta: -100},
		},
		Data: map[string]interface{}{"key": "val"},
	}
	err := ledgerApp.CreateTransaction(txn)
	assert.NoError(t, err, "Error while creating transaction")

	// Cross validate the transaction in the ledger
	ledgerSearchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"must": map[string]interface{}{
				"fields": []map[string]interface{}{
					{"id": map[string]interface{}{"eq": txnID}},
				},
			},
		},
	}
	payload, _ := json.Marshal(ledgerSearchQuery)
	ledgerSearchURL := ts.ledgerEndpoint + TransactionsSearchAPI
	req, _ := http.NewRequest("POST", ledgerSearchURL, bytes.NewBuffer(payload))
	req.Header.Add("Authorization", ts.ledgerAuthToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("Error while request to ledger:", err)
	}

	var transactions []Transaction
	err = unmarshalResponse(resp, &transactions)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(transactions), "Transactions count doesn't match")
	assert.Equal(t, txnID, transactions[0].ID, "Transaction ID doesn't match")
}

func (ts *TransactionsSuite) TestCreateTransactionInvalid() {
	t := ts.T()
	ledgerApp := NewLedger(ts.ledgerEndpoint, ts.ledgerAuthToken)

	// Prepare and create a valid transaction
	txnID := NewUUID()
	txn := &Transaction{
		ID: txnID,
		Lines: []*TransactionLine{
			&TransactionLine{AccountID: "CX", Delta: 101},
			&TransactionLine{AccountID: "CY", Delta: -100},
		},
		Data: map[string]interface{}{"key": "val"},
	}
	err := ledgerApp.CreateTransaction(txn)
	assert.Equal(t, ErrTransactionInvalid, err, "Invalid error code")
}

func (ts *TransactionsSuite) TestCreateTransactionDuplicate() {
	t := ts.T()
	ledgerApp := NewLedger(ts.ledgerEndpoint, ts.ledgerAuthToken)

	// Prepare a valid transaction
	txnID := NewUUID()
	txn := &Transaction{
		ID: txnID,
		Lines: []*TransactionLine{
			&TransactionLine{AccountID: "CX", Delta: 100},
			&TransactionLine{AccountID: "CY", Delta: -100},
		},
		Data: map[string]interface{}{"key": "val"},
	}

	// Create transaction for the first time
	err := ledgerApp.CreateTransaction(txn)
	assert.NoError(t, err, "Error while creating transaction")

	// Repeat the same transaction again
	err = ledgerApp.CreateTransaction(txn)
	assert.Equal(t, ErrTransactionDuplicate, err, "Invalid error code")
}

func (ts *TransactionsSuite) TestCreateTransactionConflict() {
	t := ts.T()
	ledgerApp := NewLedger(ts.ledgerEndpoint, ts.ledgerAuthToken)

	// Prepare and create a valid transaction
	txnID := NewUUID()
	txn1 := &Transaction{
		ID: txnID,
		Lines: []*TransactionLine{
			&TransactionLine{AccountID: "CX", Delta: 100},
			&TransactionLine{AccountID: "CY", Delta: -100},
		},
		Data: map[string]interface{}{"key": "val"},
	}
	err := ledgerApp.CreateTransaction(txn1)
	assert.NoError(t, err, "Error while creating transaction")

	// Repeat the transaction with same ID but different lines
	txn2 := &Transaction{
		ID: txnID,
		Lines: []*TransactionLine{
			&TransactionLine{AccountID: "CX", Delta: 200},
			&TransactionLine{AccountID: "CY", Delta: -200},
		},
		Data: map[string]interface{}{"key": "val"},
	}
	err = ledgerApp.CreateTransaction(txn2)
	assert.Equal(t, ErrTransactionConflict, err, "Invalid error code")
}

func TestTransactionsSuite(t *testing.T) {
	suite.Run(t, new(TransactionsSuite))
}
