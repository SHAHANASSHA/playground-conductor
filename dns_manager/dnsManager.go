package dns_manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

type DNSRecord struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	ChangeType string `json:"changetype"`
}

type DNSRequest struct {
	RRSets []DNSRecord `json:"rrsets"`
}

type Response struct {
	Key string `json:"key"`
}

type OpenPortSubdomain struct {
	Subdomain string `json:"subdomain"`
}

type PortData struct {
	ID       string              `json:"id"`
	Domain   string              `json:"domain"`
	OpenPort []OpenPortSubdomain `json:"openport"`
}

type DNSRecord2 struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`
}

type RRSet struct {
	Name       string       `json:"name"`
	Type       string       `json:"type"`
	ChangeType string       `json:"changetype"`
	Records    []DNSRecord2 `json:"records"`
	TTL        int          `json:"ttl"`
}

type DNSPayload struct {
	RRSets []RRSet `json:"rrsets"`
}

type OpenPortPayload struct {
	Domain    string `json:"domain"`
	Subdomain string `json:"subdomain"`
	PgIP      string `json:"pg_ip"`
	VM        string `json:"vm"`
	Port      string `json:"port"`
	ShortID   string `json:"short_id"`
}

func OpenPort(w http.ResponseWriter, r *http.Request, rdb *redis.Client, ctx context.Context) {
	if r.Method != "POST" {
		sendResponse(w, nil, http.StatusNotFound)
		return
	}
	var data OpenPortPayload
	pdnsKey := "a7eadf75278dd026e54c24d3aeff992cd5a8fc19"
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	domain := data.Domain
	subdomain := data.Subdomain
	pgIP := data.PgIP
	vm := data.VM
	port := data.Port
	shortID := data.ShortID

	ipMap := map[string]string{
		"vm1": "172.16.0.2",
		"vm2": "172.16.0.3",
		"vm3": "172.16.0.4",
		"vm4": "172.16.0.5",
		"vm5": "172.16.0.6",
	}

	payload := DNSPayload{
		RRSets: []RRSet{
			{
				Name:       subdomain,
				Type:       "A",
				ChangeType: "REPLACE",
				Records: []DNSRecord2{
					{Content: pgIP, Disabled: false},
				},
				TTL: 3600,
			},
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error marshaling DNS payload: %v", err)
		http.Error(w, "Error marshaling DNS payload", http.StatusBadRequest)
		return
	}

	req, err := http.NewRequest("PATCH",
		fmt.Sprintf("http://127.0.0.1:8081/api/v1/servers/localhost/zones/%s.", domain),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Fatalf("Error creating DNS request: %v", err)
		http.Error(w, "Error creating DNS request", http.StatusBadRequest)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", pdnsKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error sending DNS request: %v", err)
		http.Error(w, "Error sending DNS request", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("DNS update failed. Status: %d, Body: %s", resp.StatusCode, string(body))
		http.Error(w, "DNS update failed. Status: "+strconv.Itoa(resp.StatusCode)+", Body: "+string(body), http.StatusBadRequest)
		return
	}

	containerPayload := map[string]string{
		"key":    shortID,
		"target": fmt.Sprintf("%s:%s", ipMap[vm], port),
	}

	containerData, err := json.Marshal(containerPayload)
	if err != nil {
		log.Fatalf("Error marshaling container payload: %v", err)
		http.Error(w, "Error marshaling container payload", http.StatusBadRequest)
		return
	}

	containerResp, err := http.Post(
		fmt.Sprintf("http://%s:8080/open-close-port", pgIP),
		"application/json",
		bytes.NewBuffer(containerData),
	)
	if err != nil {
		log.Fatalf("Error sending container request: %v", err)
		http.Error(w, "Error sending container request", http.StatusBadRequest)
		return
	}
	defer containerResp.Body.Close()

	if containerResp.StatusCode != 200 {
		body, _ := io.ReadAll(containerResp.Body)
		log.Fatalf("Failed to open port. Status: %d, Body: %s", containerResp.StatusCode, string(body))
		http.Error(w, "Failed to open port. Status: "+strconv.Itoa(containerResp.StatusCode)+", Body: "+string(body), http.StatusBadRequest)
		return
	}

	sendResponse(w, nil, http.StatusOK)
}

func RemoveContainer(w http.ResponseWriter, r *http.Request, rdb *redis.Client, ctx context.Context) {
	if r.Method != "PATCH" {
		sendResponse(w, nil, http.StatusNotFound)
		return
	}
	var data PortData
	pdnsKey := "a7eadf75278dd026e54c24d3aeff992cd5a8fc19"
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var count = 0

	for _, pp := range data.OpenPort {
		if pp.Subdomain == "" {
			continue
		}

		reqBody := DNSRequest{
			RRSets: []DNSRecord{
				{
					Name:       pp.Subdomain,
					Type:       "A",
					ChangeType: "DELETE",
				},
			},
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			log.Println("JSON marshal error:", err)
			continue
		}

		req, err := http.NewRequest("PATCH",
			fmt.Sprintf("http://127.0.0.1:8081/api/v1/servers/localhost/zones/%s.", data.Domain),
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			log.Println("HTTP request creation error:", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", pdnsKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Request error:", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 204 {
			log.Printf("DNS update failed for %s. Status: %d\n", pp.Subdomain, resp.StatusCode)
			continue
		}
		count++
	}
	deleted, err := rdb.Del(ctx, data.ID).Result()
	if err != nil {
		fmt.Printf("%s", strconv.Itoa(int(deleted)))
		sendResponse(w, nil, http.StatusInternalServerError)
	}

	if count == len(data.OpenPort) {
		sendResponse(w, nil, http.StatusOK)
		return
	}
	sendResponse(w, nil, http.StatusInternalServerError)
}

func ClosePort(w http.ResponseWriter, r *http.Request, rdb *redis.Client, ctx context.Context) {
	if r.Method != "PATCH" {
		sendResponse(w, nil, http.StatusNotFound)
		return
	}
	vars := mux.Vars(r)
	short_id := vars["short_id"]

	var data PortData
	pdnsKey := "a7eadf75278dd026e54c24d3aeff992cd5a8fc19"
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	openport := data.OpenPort[0]

	reqBody := DNSRequest{
		RRSets: []DNSRecord{
			{
				Name:       openport.Subdomain,
				Type:       "A",
				ChangeType: "DELETE",
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Println("JSON marshal error:", err)
		http.Error(w, "JSON marshal error", http.StatusBadRequest)
		return
	}

	// To the DNS
	req, err := http.NewRequest("PATCH",
		fmt.Sprintf("http://127.0.0.1:8081/api/v1/servers/localhost/zones/%s.", data.Domain),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Println("HTTP request creation error:", err)
		http.Error(w, "HTTP request creation error", http.StatusBadRequest)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", pdnsKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Request error:", err)
		http.Error(w, "Request error", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		log.Printf("DNS update failed for %s. Status: %d\n", openport.Subdomain, resp.StatusCode)
		http.Error(w, "DNS update failed for "+openport.Subdomain+". Status: "+strconv.Itoa(resp.StatusCode), http.StatusBadRequest)
		return
	}

	// To the container
	ip := rdb.Get(ctx, data.ID).Val()
	url := fmt.Sprintf("http://%s:8080/open-close-port/%s", ip, short_id)
	req2, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Println("HTTP request creation error:", err)
		http.Error(w, "HTTP request creation error", http.StatusBadRequest)
		return
	}

	resp, err = http.DefaultClient.Do(req2)
	if err != nil {
		log.Println("Request error:", err)
		http.Error(w, "Request error", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Closing port failed in container")
		http.Error(w, "Closing port failed in container", http.StatusBadRequest)
		return
	}

	sendResponse(w, nil, http.StatusOK)
}

func sendResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
