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

	req, err := http.NewRequest("GET", "https://api.starlingbank.com/api/v2/accounts", nil)

	authHeader := fmt.Sprintf("Bearer %s", token)

	req.Header.Add("Authorization", authHeader)

	resp, err := client.Do(req)

	if err != nil {
		log.Fatal("Error making request to Starling.\n[ERRO] -", err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var dat map[string]interface{}

	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}

	accounts := dat["accounts"].([]interface{})
	account := accounts[0].(map[string]interface{})
	accountUid := account["accountUid"].(string)
	defaultCategory := account["defaultCategory"].(string)
	log.Printf("accountUid: %s", accountUid)
	log.Printf("defaultCategory: %s", defaultCategory)

	return nil
}
