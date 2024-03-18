package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Svirex/microurl/internal/pkg/models"
)

func main() {
	body := []models.BatchRequestRecord{
		{
			CorrID: "1",
			URL:    "https://cya.ru",
		},
		{
			CorrID: "2",
			URL:    "https://dya.ru",
		},
	}
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten/batch", bytes.NewReader(bodyBytes))

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, _ := client.Do(req)

	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	fmt.Println(string(respBody))
}
