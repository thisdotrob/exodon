package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func SyncStarling(ctx context.Context, m PubSubMessage) error {
	var (
		host     = os.Getenv("EXODON_PG_HOST")
		port     = os.Getenv("EXODON_PG_PORT")
		user     = os.Getenv("EXODON_PG_USER")
		password = os.Getenv("EXODON_PG_PASSWORD")
		dbname   = os.Getenv("EXODON_PG_DBNAME")
	)

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	var updated_at string
	var changes_since string
	row := db.QueryRow("SELECT updated_at FROM starling_transactions ORDER BY updated_at desc LIMIT 1")
	switch err := row.Scan(&updated_at); err {
	case sql.ErrNoRows:
		changes_since = "2017-01-01T00:00:00Z"
	case nil:
		t, err := time.Parse(time.RFC3339, updated_at)
		if err != nil {
			panic(err)
		}
		changes_since = t.Add(time.Hour * 24 * 7 * -1).UTC().Format(time.RFC3339)
	default:
		panic(err)
	}

	fmt.Println(changes_since)

	token := os.Getenv("STARLING_TOKEN")

	client := &http.Client{}

	authHeader := fmt.Sprintf("Bearer %s", token)

	accounts_req, err := http.NewRequest("GET", "https://api.starlingbank.com/api/v2/accounts", nil)
	accounts_req.Header.Add("Authorization", authHeader)

	accounts_resp, err := client.Do(accounts_req)

	if err != nil {
		log.Fatal("Error making accounts request to Starling.\n[ERRO] -", err)
		panic(err)
	}

	defer accounts_resp.Body.Close()

	if accounts_resp.StatusCode != 200 {
		log.Fatal("Non 200 making accounts request to Starling.\n[CODE] -", accounts_resp.StatusCode)
	}

	accounts_body, _ := ioutil.ReadAll(accounts_resp.Body)

	accounts_file_err := ioutil.WriteFile("/tmp/accounts.json", accounts_body, 0644)
	if accounts_file_err != nil {
		log.Fatal("Error writing accounts json.\n[ERRO] -", accounts_file_err)
		panic(accounts_file_err)
	}

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

	transactions_req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.starlingbank.com/api/v2/feed/account/%s/category/%s?changesSince=%s",
			account_uid, default_category, changes_since),
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

	transactions_file_err := ioutil.WriteFile("/tmp/transactions.json", transactions_body, 0644)

	if transactions_file_err != nil {
		log.Fatal("Error writing transactions json.\n[ERRO] -", transactions_file_err)
		panic(transactions_file_err)
	}

	type Amount struct {
		Currency    string `json:"currency"`
		Minor_units int    `json:"minorUnits"`
	}

	type Transaction struct {
		Feed_item_uid                           string  `json:"feedItemUid"`
		Category_uid                            string  `json:"categoryUid"`
		Amount                                  Amount  `json:"amount"`
		Source_amount                           Amount  `json:"sourceAmount"`
		Direction                               string  `json:"direction"`
		Updated_at                              string  `json:"updatedAt"`
		Transaction_time                        string  `json:"transactionTime"`
		Settlement_time                         *string `json:"settlementTime"`
		Source                                  string  `json:"source"`
		Source_sub_type                         *string `json:"sourceSubType"`
		Status                                  string  `json:"status"`
		Counter_party_type                      string  `json:"counterPartyType"`
		Counter_party_uid                       *string `json:"counterPartyUid"`
		Counter_party_name                      string  `json:"counterPartyName"`
		Counter_party_sub_entity_uid            *string `json:"counterPartySubEntityUid"`
		Counter_party_sub_entity_name           *string `json:"counterPartySubEntityName"`
		Counter_party_sub_entity_identifier     *string `json:"counterPartySubEntityIdentifier"`
		Counter_party_sub_entity_sub_identifier *string `json:"counterPartySubEntitySubIdentifier"`
		Reference                               *string `json:"reference"`
		Country                                 string  `json:"country"`
		Spending_category                       string  `json:"spendingCategory"`
		User_note                               *string `json:"userNote"`
	}

	type Transactions struct {
		Feed_items []Transaction `json:"feedItems"`
	}

	var transactions Transactions

	if err := json.Unmarshal(transactions_body, &transactions); err != nil {
		panic(err)
	}

	sqlStatement := `
	  INSERT INTO starling_transactions (
		feed_item_uid,
		category_uid,
		currency,
		minor_units,
		source_currency,
		source_minor_units,
		direction,
		updated_at,
		transaction_time,
		settlement_time,
		source,
		source_sub_type,
		status,
		counter_party_type,
		counter_party_uid,
		counter_party_name,
		counter_party_sub_entity_uid,
		counter_party_sub_entity_name,
		counter_party_sub_entity_identifier,
		counter_party_sub_entity_sub_identifier,
		reference,
		country,
		spending_category,
		user_note
	  ) VALUES (
	        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
	        $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
	  ) ON CONFLICT (feed_item_uid) DO UPDATE SET
		category_uid = $2,
		currency = $3,
		minor_units = $4,
		source_currency = $5,
		source_minor_units = $6,
		direction = $7,
		updated_at = $8,
		transaction_time = $9,
		settlement_time = $10,
		source = $11,
		source_sub_type = $12,
		status = $13,
		counter_party_type = $14,
		counter_party_uid = $15,
		counter_party_name = $16,
		counter_party_sub_entity_uid = $17,
		counter_party_sub_entity_name = $18,
		counter_party_sub_entity_identifier = $19,
		counter_party_sub_entity_sub_identifier = $20,
		reference = $21,
		country = $22,
		spending_category = $23,
		user_note = $24`

	for _, t := range transactions.Feed_items {
		_, err = db.Exec(sqlStatement,
			t.Feed_item_uid,
			t.Category_uid,
			t.Amount.Currency,
			t.Amount.Minor_units,
			t.Source_amount.Currency,
			t.Source_amount.Minor_units,
			t.Direction,
			t.Updated_at,
			t.Transaction_time,
			t.Settlement_time,
			t.Source,
			t.Source_sub_type,
			t.Status,
			t.Counter_party_type,
			t.Counter_party_uid,
			t.Counter_party_name,
			t.Counter_party_sub_entity_uid,
			t.Counter_party_sub_entity_name,
			t.Counter_party_sub_entity_identifier,
			t.Counter_party_sub_entity_sub_identifier,
			t.Reference,
			t.Country,
			t.Spending_category,
			t.User_note,
		)

		if err != nil {
			fmt.Printf("%+v\n", t)
			panic(err)
		}
	}

	return nil
}

func main() {
	m := PubSubMessage{
		Data: []byte(""),
	}
	SyncStarling(context.Background(), m)
}
