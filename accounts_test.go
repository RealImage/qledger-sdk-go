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

type AccountsSuite struct {
	suite.Suite
	ledgerEndpoint  string
	ledgerAuthToken string
}

func (as *AccountsSuite) SetupSuite() {
	as.ledgerEndpoint = os.Getenv("LEDGER_ENDPOINT")
	as.ledgerAuthToken = os.Getenv("LEDGER_AUTH_TOKEN")
}

func (as *AccountsSuite) TestGetAccountBeforeTransaction() {
	t := as.T()
	accountID := NewUUID()
	ledgerApp := NewLedger(as.ledgerEndpoint, as.ledgerAuthToken)

	account, err := ledgerApp.GetAccount(accountID)
	assert.Equal(t, ErrAccountNotfound, err, "Invalid error code")
	var nilAccount *Account
	assert.Equal(t, nilAccount, account, "Account should not exist")
}

func (as *AccountsSuite) TestGetAccountAfterTransaction() {
	t := as.T()
	accountID := NewUUID()
	amount := 1000

	// Add balance to CREDIT account
	txn := &Transaction{
		ID: NewUUID(),
		Lines: []*TransactionLine{
			&TransactionLine{AccountID: accountID, Delta: amount},
			&TransactionLine{AccountID: "TESTACCOUNT", Delta: -amount},
		},
	}
	payload, _ := json.Marshal(txn)
	ledgerTransactionsURL := as.ledgerEndpoint + TransactionsAPI
	req, _ := http.NewRequest("POST", ledgerTransactionsURL, bytes.NewBuffer(payload))
	req.Header.Add("Authorization", as.ledgerAuthToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("Error while request to ledger:", err)
	}
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Invalid response code")

	ledgerApp := NewLedger(as.ledgerEndpoint, as.ledgerAuthToken)
	account, err := ledgerApp.GetAccount(accountID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, accountID, account.ID, "Account ID doesn't match")
	assert.Equal(t, amount, account.Balance, "Account balance doesn't match")
}

func TestAccountsSuite(t *testing.T) {
	suite.Run(t, new(AccountsSuite))
}
