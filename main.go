package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
)

// Pack sizes config file
type Config struct {
	PackSizes []int `json:"packSizes"`
}

// JSON request - order quantity
type OrderRequest struct {
	Quantity int `json:"quantity"`
}

// JSON Response - number of packs
type PackResponse struct {
	Packs map[int]int `json:"packs"`
}

// Pack Sizes from config.json file
var packSizes []int

// Load the pack sizes from the config.json file
func loadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Error opening config file:", err)
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error decoding config file:", err)
		return
	}

	packSizes = config.PackSizes
}

// Pack calculator function - function could be improved as we don't always want the smallest pack to be assigned if there's any remaining
// Sort pack sizes in descending order
func calculatePacks(order int, packSizes []int) map[int]int {
	sort.Sort(sort.Reverse(sort.IntSlice(packSizes)))
	// key (pack sizes) value (number of packs)
	result := make(map[int]int)
	remaining := order

	// Loop through each pack size
	for _, size := range packSizes {
		if remaining >= size {
			numPacks := remaining / size
			result[size] = numPacks
			remaining -= numPacks * size
		}
	}

	// Handle remaining order & assign one more pack (smallest). 
	if remaining > 0 {
		result[packSizes[len(packSizes)-1]]++
	}

	return result
}

// HTTP Post Request 
func packHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// JSON Request Body
	var order OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Calculate Packs
	packs := calculatePacks(order.Quantity, packSizes)
	response := PackResponse{Packs: packs}

	// JSON Response Body
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
    loadConfig()
    if len(packSizes) == 0 {
        fmt.Println("No pack sizes configured. Please check the config file.")
        return
    }

    http.HandleFunc("/calculate-packs", packHandler)
    http.Handle("/", http.FileServer(http.Dir("./static")))

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080" // Default port
    }

    fmt.Println("Server started at port:", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        fmt.Println("Failed to start server:", err)
    }
}
