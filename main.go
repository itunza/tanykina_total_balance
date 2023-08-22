package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const baseURL = "https://erp.spaceai.io/api/resource/Purchase Order"

type Response struct {
	Data []struct {
		TotalQty float64 `json:"total_qty"`
	} `json:"data"`
}

type JsonResponse struct {
	Sno           string  `json:"sno"`
	TotalQuantity float64 `json:"total_quantity"`
}

var client = &http.Client{}

func handler(w http.ResponseWriter, r *http.Request) {
	sno := r.URL.Query().Get("sno")
	if sno == "" {
		http.Error(w, "sno is required", http.StatusBadRequest)
		return
	}

	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endDate := now.Format("2006-01-02")
	startDate := firstOfMonth.Format("2006-01-02")

	url := fmt.Sprintf("%s?fields=[\"total_qty\"]&filters=[[\"sno\",\"like\",\"%%%s%%\"],[\"transaction_date\",\"Between\",[\"%s\",\"%s\"]],[\"docstatus\",\"=\",1]]&limit=200", baseURL, sno, startDate, endDate)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Failed to create request:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	token := os.Getenv("ERP_TOKEN")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to fetch data:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data Response
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println("Failed to parse response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	totalQty := 0.0
	for _, item := range data.Data {
		totalQty += item.TotalQty
	}

	totalQty = float64(int(totalQty*100)) / 100 //

	response := JsonResponse{
		Sno:           sno,
		TotalQuantity: totalQty,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Println("Failed to marshal JSON:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func main() {
	// load env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/getTotalQty", handler)
	log.Println("Server started on :8890")
	log.Fatal(http.ListenAndServe(":8890", nil))
}

