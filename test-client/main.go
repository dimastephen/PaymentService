package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

type Body struct {
	Currency       string `json:"Currency"`
	Amount         int64  `json:"Amount"`
	MerchantID     string `json:"MerchantID"`
	IdempotencyKey string `json:"IdempotencyKey"`
}

func main() {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
			MaxIdleConns:        100,
			MaxConnsPerHost:     100,
		},
		Timeout: 5 * time.Second,
	}
	wg := sync.WaitGroup{}
	for i := range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(i) * time.Second)

			body := &Body{
				Currency:       "USD",
				Amount:         rand.Int63n(1000000),
				MerchantID:     gofakeit.Word(),
				IdempotencyKey: gofakeit.UUID(),
			}
			bodyBytes, err := json.Marshal(body)
			if err != nil {
				log.Fatal(err)
			}
			req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/payments", bytes.NewBuffer(bodyBytes))
			if err != nil {
				log.Fatal(err)
			}
			req.Header.Add("Content-Type", "application/json")

			resp, err := client.Do(req)
			fmt.Printf("request %v sent, status %v\n", i, resp.Status)
			resp, err = client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("request %v sent, status %v\n", i, resp.Status)
			defer resp.Body.Close()
			time.Sleep(2 * time.Second)
			req, err = http.NewRequest(http.MethodGet, "http://localhost:8080/payments/"+body.IdempotencyKey, nil)

			resp, err = client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("request GET %v sent, status %v\n", i, resp.Status)
			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)
			fmt.Printf("request %v: GET status %v, body: %s\n", i, resp.Status, string(respBody))
		}()
	}
	wg.Wait()

}
