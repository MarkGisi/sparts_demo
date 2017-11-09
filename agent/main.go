package main

// Licensing: Apache-2.0
/*
 *  Copyright (c) 2017 Wind River Systems, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software  distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
 * OR CONDITIONS OF ANY KIND, either express or implied.
 */

import (
		"bytes"
		"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gorilla/mux" // BSD-3-Clause
)

const (
	config_file = "./agent_config.json"
)

type Configuration struct {
	HttpPort     int    `json:"http_port"`
	RebootScript string `json:"demo_reboot_script"`
}

func GetConfigurationInfo(configuration *Configuration) {

	// When this func is called the first time we want to load the config file
	// regardless of whether the value of config_reload_allowed is true.
	// If this config variable is false on furture invocations we do not
	// want to allow the config file to be reloaded. We created a temp struct
	// to load and check config_reload_allowed first. If this varible is true
	// then we can proceed to load the current values.

	var temp_config Configuration

	file, open_err := os.Open(config_file)
	if open_err != nil {
  		log.Fatal(open_err)
  	}
	decoder := json.NewDecoder(file)

	temp_config = Configuration{}
	err := decoder.Decode(&temp_config)
	if err != nil {
		fmt.Println("error:", err)
	}

    *configuration = temp_config
	fmt.Println("Configuration:")
	fmt.Println("-----------------------------------------------")
	fmt.Println("http port	  = ", configuration.HttpPort)
	fmt.Println("reboot script  = ", configuration.RebootScript)

}

// Obtain IP Address of host machine.
func GetHostIPAddress() string {
	conn, err := net.Dial("udp", "example.com:80")
	if err != nil {
		log.Printf("[TOOLS] SYSADMIIIIIN : cannot use UDP")
		return "0.0.0.0"
	}
	defer conn.Close()
	torn := strings.Split(conn.LocalAddr().String(), ":")
	return torn[0]
}

// Print useful info
func displayURLRequest(request *http.Request) {
	fmt.Println()
	fmt.Println("-----------------------------------------------")
	fmt.Println("URL Request: ", request.URL.Path)
	log.Println()
	fmt.Println("Client IP:", GetHostIPAddress())
}

// Display debug info about a url request
func displayURLReply(url_reply string) {
	// Display http reply content for monitoring and testing purposes
	fmt.Println("-----------------------------------------------")
	fmt.Println("URL Reply:")
	fmt.Println("---------------:")
	fmt.Println(url_reply)
}

// Standard method for acknowleding success status for http requests
func httpSuccessReply(http_reply http.ResponseWriter) {
	type messageReply struct {
		Status string `json:"status"`
	}
	httpSuccessReply := messageReply{Status: "success"}
	httpSendReply(http_reply, httpSuccessReply)
}


// Pretty print (format) the json reply.
func httpSendReply(http_reply http.ResponseWriter, data interface{}) {

	// We want to pretty print the json reply. We need to wrap:
	//    json.NewEncoder(http_reply).Encode(reply)
	// with the following code:

	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "   ") // tells how much to indent "  " spaces.
	err := encoder.Encode(data)

	displayURLReply(buffer.String())

	if err != nil {
		io.WriteString(http_reply, "error - could not encode reply")
	} else {
		io.WriteString(http_reply, buffer.String())
	}
}

// Standard method for acknowleding success status for http requests
func httpSuccesshttpSuccessReply(http_reply http.ResponseWriter) {
	type messageReply struct {
		Status string `json:"status"`
	}
	httpSuccessReply := messageReply{Status: "success"}
	httpSendReply(http_reply, httpSuccessReply)
}



// Handle:  GET /api/sparts_demo/ping
func GET_Ping_EndPoint(http_reply http.ResponseWriter, request *http.Request) {

	displayURLRequest(request)

	// reply success to indicate running.
	httpSuccessReply(http_reply)
}

// Handle:  GET /api/sparts_demo/reboot
func GET_Reboot_EndPoint(http_reply http.ResponseWriter, request *http.Request) {

/****
	// call script
	//cmd, err := exec.Command("sh", MAIN_config.RebootScript).Output()
	cmd, err := exec.Command("sh", "-c", MAIN_config.RebootScript).Output()
	//out, err := exec.Command("date").Output()
	//_, err := cmd.Output()
	if err != nil {
		println(err.Error())
	}
	fmt.Printf("output is: %s\n", cmd)
	***/
	displayURLRequest(request)

	cmd := exec.Command("sh", MAIN_config.RebootScript)
	
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Waiting for command to finish...")
	err = cmd.Wait()
	log.Printf("Command finished with error: %v", err)
	fmt.Println("output is:", out.String())
	
	// reply success to indicate running.
	httpSuccessReply(http_reply)
}

// Gloabl request counter
var host_pid = os.Getpid() // process id
var MAIN_config Configuration
var http_ip_address = GetHostIPAddress()
var router = mux.NewRouter()

func main() {
	fmt.Println()
	fmt.Println()

	// Read configuration file to set a number of global settings
	GetConfigurationInfo(&MAIN_config)


	router.HandleFunc("/api/sparts_demo/ping", GET_Ping_EndPoint).Methods("GET")
	router.HandleFunc("/api/sparts_demo/reboot", GET_Reboot_EndPoint).Methods("GET")

	fmt.Println()
	fmt.Println("-----------------------------------------------")
	fmt.Println("Starting Conductor ...")
	fmt.Println("Host IP:	=", http_ip_address)
	fmt.Println("Host Port:	=", MAIN_config.HttpPort)
	fmt.Println("Host PID:	=", host_pid)

	// Create port string, e.g., for port 8080 we create ":8080" needed for ListenAndServe ()
	port_str := ":" + strconv.Itoa(MAIN_config.HttpPort)
	fmt.Println("Listening on port", port_str, "...")

	// Listen and responsed to requests.
	log.Fatal(http.ListenAndServe(port_str, router))
}
