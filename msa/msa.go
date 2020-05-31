package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lithammer/shortuuid"
)

// Email represents a message sent by the user
type Email struct {
	UUID     string
	Content  string
	Sender   string
	Receiver string
}

// Inbox represents the messages sent to each user
type Inbox struct {
	Messages map[string]Email
}

// Outbox represents the messages being sent by each user
type Outbox struct {
	Messages map[string]Email
}

// Maps each user to an inbox and outbox
var userInbox map[string]Inbox
var userOutbox map[string]Outbox

// AddToInbox posts a type Email into the users inbox
// If the user does not exist, a new user folder is created
func AddToInbox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]

	currInbox := userInbox[user]
	if currInbox.Messages == nil {
		currInbox.Messages = make(map[string]Email)
	}

	decoder := json.NewDecoder(r.Body)
	var email Email
	if err := decoder.Decode(&email); err == nil {
		w.WriteHeader(http.StatusCreated)
		currInbox.Messages[email.UUID] = email
		userInbox[user] = currInbox
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

// ReadAllInbox finds every message in the users inbox and displays them
func ReadAllInbox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]

	if currInbox, ok := userInbox[user]; ok {
		w.WriteHeader(http.StatusOK)

		var emails []Email

		for _, v := range currInbox.Messages {
			emails = append(emails, v)
		}
		if enc, err := json.Marshal(emails); err == nil {
			w.Write([]byte(enc))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// ReadInboxMessage lets the user look at the contents of one message provided they know its UUID
func ReadInboxMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	uuid := vars["uuid"]

	if currInbox, ok := userInbox[user]; ok {
		w.WriteHeader(http.StatusOK)
		v := currInbox.Messages[uuid]
		if enc, err := json.Marshal(v); err == nil {
			w.Write([]byte(enc))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// DeleteInboxMessage lets the user delete a message provided they know its UUID
func DeleteInboxMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	uuid := vars["uuid"]

	if currInbox, ok := userInbox[user]; ok {
		if _, ok := currInbox.Messages[uuid]; ok {
			w.WriteHeader(http.StatusNoContent)
			delete(currInbox.Messages, uuid)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// AddToOutbox posts a type Email into the users outbox
// If the user does not exist, a new user is made
func AddToOutbox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]

	currOutbox := userOutbox[user]
	if currOutbox.Messages == nil {
		currOutbox.Messages = make(map[string]Email)
	}

	decoder := json.NewDecoder(r.Body)
	var email Email
	if err := decoder.Decode(&email); err == nil {
		w.WriteHeader(http.StatusCreated)

		u := shortuuid.New()
		email.UUID = u
		currOutbox.Messages[u] = email
		userOutbox[user] = currOutbox
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

// ReadEveryOutbox creates a list of every message in every users outbox
func ReadEveryOutbox(w http.ResponseWriter, r *http.Request) {
	var emails []Email
	for _, currOutbox := range userOutbox {
		for _, currEmail := range currOutbox.Messages {
			emails = append(emails, currEmail)
		}
	}
	if enc, err := json.Marshal(emails); err == nil {
		w.Write([]byte(enc))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ReadAllOutbox finds every message in the users outbox and displays them
func ReadAllOutbox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]

	if currOutbox, ok := userInbox[user]; ok {
		w.WriteHeader(http.StatusOK)

		var emails []Email

		for _, v := range currOutbox.Messages {
			emails = append(emails, v)
		}
		if enc, err := json.Marshal(emails); err == nil {
			w.Write([]byte(enc))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// ReadOutboxMessage lets the user look at the contents of one message provided they know its UUID
func ReadOutboxMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	uuid := vars["uuid"]

	if currOutbox, ok := userOutbox[user]; ok {
		w.WriteHeader(http.StatusOK)
		v := currOutbox.Messages[uuid]
		if enc, err := json.Marshal(v); err == nil {
			w.Write([]byte(enc))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// DeleteOutboxMessage lets the user delete a message provided they know its UUID
func DeleteOutboxMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	uuid := vars["uuid"]

	if currOutbox, ok := userOutbox[user]; ok {
		if _, ok := currOutbox.Messages[uuid]; ok {
			w.WriteHeader(http.StatusNoContent)
			delete(currOutbox.Messages, uuid)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/inbox/{user}", AddToInbox).Methods("POST")
	router.HandleFunc("/inbox/{user}", ReadAllInbox).Methods("GET")
	router.HandleFunc("/inbox/{user}/{uuid}", ReadInboxMessage).Methods("GET")
	router.HandleFunc("/inbox/{user}/{uuid}", DeleteInboxMessage).Methods("DELETE")

	router.HandleFunc("/outbox/{user}", AddToOutbox).Methods("POST")
	router.HandleFunc("/outbox", ReadEveryOutbox).Methods("GET")
	router.HandleFunc("/outbox/{user}", ReadAllOutbox).Methods("GET")
	router.HandleFunc("/outbox/{user}/{uuid}", ReadOutboxMessage).Methods("GET")
	router.HandleFunc("/outbox/{user}/{uuid}", DeleteOutboxMessage).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8888", router))
}

func main() {
	userInbox = make(map[string]Inbox)
	userOutbox = make(map[string]Outbox)
	handleRequests()
}
