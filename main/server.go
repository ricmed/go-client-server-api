package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func getCotacao(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var cotacao Cotacao
	if err := json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
		return "", err
	}

	return cotacao.USDBRL.Bid, nil
}

func saveCotacao(ctx context.Context, db *sql.DB, bid string) error {
	query := "INSERT INTO cotacoes (bid, timestamp) VALUES (?, ?)"
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, bid, time.Now())
	return err
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	bid, err := getCotacao(ctx)
	if err != nil {
		log.Println("Error fetching cotacao:", err)
		http.Error(w, "Error fetching cotacao", http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		log.Println("Error opening database:", err)
		http.Error(w, "Error opening database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	ctx, cancel = context.WithTimeout(r.Context(), 10*time.Millisecond)
	defer cancel()

	if err := saveCotacao(ctx, db, bid); err != nil {
		log.Println("Error saving cotacao:", err)
		http.Error(w, "Error saving cotacao", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"bid": bid})
}

func main() {
	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY, bid TEXT, timestamp DATETIME)")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", handler)
	fmt.Println("Server is listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
