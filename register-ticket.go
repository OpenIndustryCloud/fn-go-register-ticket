package main

/*
This API would accept JSON string as POST body and
create a Ticket in Zen Desk/Fresh Desk

INPUT - Zen Desk Create Ticket compliant JSON

OUTPUT - Ticket Meta Data JSON from Response Object
{
	"id": 133382282992,
	"ticket_id": 39,
	"created_at": "2017-10-25T18:32:55Z",
	"author_id": 115428050612,
	"metadata": {
		"system": {
		"ip_address": "2.122.25.146",
		"location": "Solihull, M2, United Kingdom",
		"latitude": 52.41669999999999,
		"longitude": -1.783299999999997
	},
	"custom": {}
	}
}
*/
import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	fdk "github.com/fnproject/fdk-go"
)

//Default values
var (
	endPoint    = "https://landg.zendesk.com/api/v2/tickets.json"
	apiKey      = ""
	apiPassword = ""
	namesapce   = "default"
	secretName  = "zendesk-secret"
	logger      = log.New(os.Stderr, "", 0)
)

func main() {
	fdk.Handle(fdk.HandlerFunc(myHandler))
}

func myHandler(ctx context.Context, in io.Reader, out io.Writer) {

	logger.Println("Executing Register Ticket API end point...", endPoint)
	//get API keys
	getAPIKeys(ctx, out)

	req, err := http.NewRequest("POST", endPoint, in)
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(apiKey, apiPassword)

	client := &http.Client{}
	zendeskAPIResp, err := client.Do(req)
	if err != nil {
		createErrorResponse(out, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.Compare(zendeskAPIResp.Status, "201 Created") != 0 {
		logger.Println("request status for ticket creation :" + zendeskAPIResp.Status)
		createErrorResponse(out, "error creating tickets", http.StatusBadRequest)
		return
	}

	var ticketResponse TicketResponse
	err = json.NewDecoder(zendeskAPIResp.Body).Decode(&ticketResponse)
	if err != nil || ticketResponse == (TicketResponse{}) {
		createErrorResponse(out, err.Error(), http.StatusBadRequest)
		return
	}
	defer zendeskAPIResp.Body.Close()

	//marshal response to JSON
	ticketAuditData := ticketResponse.Audit
	// ticketResponseJSON, err := json.Marshal(&ticketAuditData)
	// if err != nil {
	// 	createErrorResponse(out, err.Error(), http.StatusBadRequest)
	// 	return
	// }

	// w.Header().Set("content-type", "application/json")
	// w.Write([]byte(ticketResponseJSON))
	fdk.SetHeader(out, "Content-Type", "application/json")
	json.NewEncoder(out).Encode(ticketAuditData)
}

func createErrorResponse(out io.Writer, message string, status int) {
	errorJSON := &Error{
		Status:  status,
		Message: message}
	//Send custom error message to caller
	// w.WriteHeader(status)
	// w.Header().Set("content-type", "application/json")
	// w.Write([]byte(errorJSON))
	fdk.SetHeader(out, "Content-Type", "application/json")
	json.NewEncoder(out).Encode(errorJSON)

}

type Error struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func getAPIKeys(ctx context.Context, out io.Writer) {
	logger.Println("[CONFIG] Reading Env variables")

	//read config
	fnctx := fdk.Context(ctx)

	apiKey = fnctx.Config["apiKey"]           // from func.yaml
	apiPassword = fnctx.Config["apiPassword"] // from command line

	logger.Printf("pass %s Value\n", apiPassword)
	logger.Printf("Key %s Value \n", apiKey)

	for k, v := range fnctx.Config {
		logger.Printf("Key %v Value %v\n", k, v)
	}

	if len(apiKey) == 0 {
		createErrorResponse(out, "Missing API Key", http.StatusBadRequest)
	}
	if len(apiPassword) == 0 {
		createErrorResponse(out, "Missing API Password", http.StatusBadRequest)
	}

}

type TicketResponse struct {
	Ticket struct {
		URL        string      `json:"url,omitempty"`
		ID         int         `json:"id,omitempty"`
		ExternalID interface{} `json:"external_id,omitempty"`

		CreatedAt    time.Time   `json:"created_at,omitempty"`
		UpdatedAt    time.Time   `json:"updated_at,omitempty"`
		DueAt        interface{} `json:"due_at,omitempty"`
		TicketFormID int64       `json:"ticket_form_id,omitempty"`
	} `json:"ticket"`
	Audit struct {
		ID        int64     `json:"id,omitempty"`
		TicketID  int       `json:"ticket_id,omitempty"`
		CreatedAt time.Time `json:"created_at,omitempty"`
		AuthorID  int64     `json:"author_id,omitempty"`
		Metadata  struct {
			System struct {
				IPAddress string  `json:"ip_address,omitempty"`
				Location  string  `json:"location,omitempty"`
				Latitude  float64 `json:"latitude,omitempty"`
				Longitude float64 `json:"longitude,omitempty"`
			} `json:"system"`
			Custom struct {
			} `json:"custom"`
		} `json:"metadata"`
	} `json:"audit"`
}
