package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Address represents an IP address being used by a server
type Address struct {
	ServerAddress string
}

// map domain names to Addresses
var servers map[string]Address

// GetIPAddress converts a domain name to an IP address
// if the IP address does not exist, it returns the senders IP address
func GetIPAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	plainAddress := vars["domainName"]
	if reply, ok := servers[plainAddress]; ok {
		w.WriteHeader(http.StatusOK)
		if enc, err := json.Marshal(reply); err == nil {
			w.Write([]byte(enc))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		reply := r.RemoteAddr
		if enc, err := json.Marshal(reply); err == nil {
			w.Write([]byte(enc))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/bluebook/{domainName}", GetIPAddress).Methods("GET")
	log.Fatal(http.ListenAndServe(":8888", router))
}

func main() {
	servers = make(map[string]Address)
	// Hard Coded server addresses
	servers["here.com"] = Address{"192.168.1.8"}
	servers["there.com"] = Address{"192.168.1.7"}
	handleRequests()
}
