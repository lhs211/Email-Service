package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Email represents a message sent by the user
type Email struct {
	UUID     string
	Content  string
	Sender   string
	Receiver string
}

// Address represents an IP address being used by a server
type Address struct {
	ServerAddress string
}

// used for hard coding ports and addresses
var msaPort string
var mtaPort string
var bluebookPort string
var bluebookAddress string
var sourceURL string

// GetAllFromOutbox creates a list of every message in every uses outbox on the MSA
func GetAllFromOutbox() []Email {
	result := CommunicateGET(sourceURL + "/outbox")
	var emails []Email
	json.Unmarshal([]byte(result), &emails)
	return emails
}

// DeleteMessageFromOutbox deletes each given message from a users outbox
func DeleteMessageFromOutbox(messages []Email) {
	for _, elem := range messages {
		CommunicateDELETE(sourceURL + "/outbox" + "/" + elem.Sender + "/" + elem.UUID)
	}
}

// AddMessageToInbox is given a list of emails which it sends to the appropiate MTA
func AddMessageToInbox(messages []Email) {
	for _, elem := range messages {
		domain := strings.Split(elem.Receiver, "@")
		result := CommunicateGET("http://" + bluebookAddress + ":" + bluebookPort + "/bluebook/" + domain[len(domain)-1])
		var ip Address
		json.Unmarshal([]byte(result), &ip)
		destinationURL := ip.ServerAddress + ":" + mtaPort + "/mta/inbox/" + elem.Receiver
		CommunicatePOST(destinationURL, elem)
	}
}

// CommunicateGET gets a response off the url specified as a parameter
func CommunicateGET(url string) []byte {
	client := &http.Client{}

	if req, err := http.NewRequest("GET", url, nil); err == nil {
		if resp, err1 := client.Do(req); err1 == nil {
			if body, err2 := ioutil.ReadAll(resp.Body); err2 == nil {
				return body
			} else {
				fmt.Printf("GET failed with %s\n", err2)
			}
		} else {
			fmt.Printf("GET failed with %s\n", err1)
		}
	} else {
		fmt.Printf("GET failed with %s\n", err)
	}
	return []byte{}
}

// CommunicatePOST sends an email to the url specified as a parameter
func CommunicatePOST(url string, email Email) {
	client := &http.Client{}

	if enc, err := json.Marshal(email); err == nil {
		if req, err1 := http.NewRequest("POST", url, bytes.NewBuffer(enc)); err1 == nil {
			if resp, err2 := client.Do(req); err2 == nil {
				if body, err3 := ioutil.ReadAll(resp.Body); err3 == nil {
					fmt.Println(string(body))
				} else {
					fmt.Printf("POST failed with %s\n", err3)
				}
			} else {
				fmt.Printf("POST failed with %s\n", err2)
			}
		} else {
			fmt.Printf("POST failed with %s\n", err1)
		}
	} else {
		fmt.Printf("POST failed with %s\n", err)
	}
}

// CommunicateDelete deletes an email at the specified url given as a parameter
func CommunicateDELETE(url string) {
	client := &http.Client{}

	if req, err := http.NewRequest("POST", url, nil); err == nil {
		if _, err1 := client.Do(req); err1 == nil {
			// nothing
		} else {
			fmt.Printf("DELETE failed with %s\n", err1)
		}
	} else {
		fmt.Printf("DELETE failed with %s\n", err)
	}
}

// PassToMSA forwards the posted data to the corresponding MSA
func PassToMSA(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var email Email
	if err := decoder.Decode(&email); err == nil {
		w.WriteHeader(http.StatusCreated)
		destinationURL := sourceURL + "/inbox/" + email.Receiver
		CommunicatePOST(destinationURL, email)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

// SendMessage calls methods to send an email across the network
func SendMessage(t time.Time) {
	messages := GetAllFromOutbox()
	AddMessageToInbox(messages)
	DeleteMessageFromOutbox(messages)
}

// DoEvery calls the specific function after each unit of time
func DoEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/mta/inbox", PassToMSA).Methods("POST")
	log.Fatal(http.ListenAndServe(":8888", router))
}

func main() {
	// hard coded port numbers and addresses
	msaPort = "3000"
	mtaPort = "3001"
	bluebookPort = "3002"
	bluebookAddress = "192.168.1.6"
	// get the address this servuce is running on
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	sourceURL = conn.LocalAddr().String()
	go DoEvery(10000*time.Millisecond, SendMessage)
	handleRequests()
}
