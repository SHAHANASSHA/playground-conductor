package examiner

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ExaminerReq struct {
	IP   string `json:"ip"`
	VM   string `json:"vm"`
	Test string `json:"test"`
	Args string `json:"args"`
}

type ExaminerRes struct {
	Success bool `json:"success"`
}

func Examiner(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendResponse(w, nil, http.StatusNotFound)
		return
	}

	var data ExaminerReq
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("http://%s:8080/examiner/test/%s/%s?args=%s", data.IP, data.VM, data.Test, data.Args)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		// fmt.Errorf("failed to send POST request: %w", err)
		http.Error(w, "failed to send POST request", http.StatusBadRequest)
		return
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		http.Error(w, "Some error occured", http.StatusBadRequest)
		return
	}

	var result ExaminerRes
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Decode error:", err)
		return
	}
	sendResponse(w, result, http.StatusOK)
}

func sendResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
