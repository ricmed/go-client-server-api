package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response:", err)
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatal("Error unmarshalling response:", err)
	}

	cotacao := result["bid"]
	err = ioutil.WriteFile("cotacao.txt", []byte(fmt.Sprintf("Dólar: %s", cotacao)), 0644)
	if err != nil {
		log.Fatal("Error writing to file:", err)
	}

	fmt.Println("Cotação salva em cotacao.txt")
}
