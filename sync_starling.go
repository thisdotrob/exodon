package starling

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func SyncStarling(ctx context.Context, m PubSubMessage) error {
	token := os.Getenv("STARLING_TOKEN")

	client := &http.Client{}

	authHeader := fmt.Sprintf("Bearer %s", token)

	accounts_req, err := http.NewRequest("GET", "https://api.starlingbank.com/api/v2/accounts", nil)
	accounts_req.Header.Add("Authorization", authHeader)

	accounts_resp, err := client.Do(accounts_req)

	if err != nil {
		log.Fatal("Error making accounts request to Starling.\n[ERRO] -", err)
	}

	defer accounts_resp.Body.Close()

	if accounts_resp.StatusCode != 200 {
		log.Fatal("Non 200 making accounts request to Starling.\n[CODE] -", accounts_resp.StatusCode)
	}

	accounts_body, _ := ioutil.ReadAll(accounts_resp.Body)

	type Account struct {
		Account_uid      string `json:"accountUid"`
		Default_category string `json:"defaultCategory"`
	}

	type Accounts struct {
		Accounts []Account `json:"accounts"`
	}

	var accounts_dat Accounts

	if err := json.Unmarshal(accounts_body, &accounts_dat); err != nil {
		panic(err)
	}

	account_uid := accounts_dat.Accounts[0].Account_uid
	default_category := accounts_dat.Accounts[0].Default_category

	last_synced := "2017-09-06T00:00:00.000Z"

	transactions_req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.starlingbank.com/api/v2/feed/account/%s/category/%s?changesSince=%s", account_uid, default_category, last_synced),
		nil,
	)
	transactions_req.Header.Add("Authorization", authHeader)
	transactions_resp, err := client.Do(transactions_req)

	if err != nil {
		log.Fatal("Error making transactions request to Starling.\n[ERRO] -", err)
	}

	defer transactions_resp.Body.Close()

	if transactions_resp.StatusCode != 200 {
		log.Fatal("Non 200 making transactions request to Starling.\n[CODE] -", transactions_resp.StatusCode)
	}

	transactions_body, _ := ioutil.ReadAll(transactions_resp.Body)

	type Amount struct {
		Currency    string `json:"currency"`
		Minor_units int    `json:"minorUnits"`
	}

	type Transaction struct {
		Feed_item_uid                           string `json:"feedItemUid"`
		Category_uid                            string `json:"categoryUid"`
		Amount                                  Amount `json:"amount"`
		Source_amount                           Amount `json:"sourceAmount"`
		Direction                               string `json:"direction"`
		Updated_at                              string `json:"updatedAt"`
		Transaction_time                        string `json:"transactionTime"`
		Settlement_time                         string `json:"settlementTime"`
		Source                                  string `json:"source"`
		Source_sub_type                         string `json:"sourceSubType"`
		Status                                  string `json:"status"`
		Counter_party_Type                      string `json:"counterPartyType"`
		Counter_party_uid                       string `json:"counterPartyUid"`
		Counter_party_Name                      string `json:"counterPartyName"`
		Counter_party_sub_entity_uid            string `json:"counterPartySubEntityUid"`
		Counter_party_sub_entity_name           string `json:"counterPartySubEntityName"`
		Counter_party_sub_entity_identifier     string `json:"counterPartySubEntityIdentifier"`
		Counter_party_sub_entity_sub_identifier string `json:"counterPartySubEntitySubIdentifier"`
		Reference                               string `json:"reference"`
		Country                                 string `json:"country"`
		Spending_category                       string `json:"spendingCategory"`
		User_note                               string `json:"userNote"`
	}

	type Transactions struct {
		Feed_items []Transaction `json:"feedItems"`
	}

	var transactions Transactions

	if err := json.Unmarshal(transactions_body, &transactions); err != nil {
		panic(err)
	}

	fmt.Println(transactions.Feed_items[0])

	return nil
}
