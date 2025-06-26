package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/devdevaraj/conductor/wait_for_port"
)

type ReqIP struct {
	IP string `json:"ip"`
}

func WaitForVMs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendResponse(w, nil, http.StatusNotFound)
		return
	}
	var data ReqIP
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	err2 := wait_for_port.WaitForPort(data.IP, "8080", 30*time.Second, 300*time.Millisecond)
	if err2 != nil {
		http.Error(w, "Playground timeout", http.StatusBadRequest)
		return
	}
	url := fmt.Sprintf("http://%s:8080/wait-for-vms", data.IP)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Playground timeout", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "unexpected status code "+strconv.Itoa(resp.StatusCode), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read response"+strconv.Itoa(resp.StatusCode), http.StatusBadRequest)
		return
	}
	sendResponse(w, body, http.StatusOK)
}

func sendResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
