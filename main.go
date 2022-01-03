package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"sync"
	"time"

	"go.bug.st/serial.v1"
)

var mutex sync.Mutex
var instructionIds []int
var killInstructionIds []int

type changes [][2]int

var mode = &serial.Mode{
	Parity:   serial.EvenParity,
	BaudRate: 115200,
	DataBits: 8,
	StopBits: serial.OneStopBit,
}

var arduino, _ = serial.Open(findArduinoPort(), mode)

var re, _ = regexp.Compile(`\w{3,4}=\w{1,4};`)

var VxxParams = []string{"V00", "V01", "V02", "V03", "V04", "V05", "V06", "V07", "V08"}
var TxxParams = []string{"T01", "T02", "T03", "T04", "T05", "T06", "T07", "T08"}
var pumpParams = []string{"PUMP_ON", "PUMP_OFF"}

var OS = runtime.GOOS

func main() {
	flushRedis()
	go DB_routine()
	http.HandleFunc("/", RootHandler)
	http.HandleFunc("/set", SetHandler)
	http.HandleFunc("/getall", GetHandler)
	http.HandleFunc("/start", StartHandler)
	http.HandleFunc("/stop", StopHandler)
	http.ListenAndServe(":9999", nil)
}

// Aquires DB & Microcontroller connections, starts a loop constantly sending command to get all states of parameters in the board, and writes them to database
func DB_routine() {
	con := getDBConnection()
	dur, ver, err := con.Ping()
	check(err)
	log.Printf("Connected to database! %v, %s", dur, ver)

	if !databaseDataExists(con) {
		createDatabaseData1h(con)
	}

	defer arduino.Close()

	for {
		_, err := arduino.Write([]byte("<GET_ALL;>"))
		// If err, reinitialize connection to device
		if err != nil {
			fmt.Printf("CONNECTION ERROR! %v", err)
			arduino, err = serial.Open(findArduinoPort(), mode)
			check(err)
		}

		scanner := bufio.NewScanner(arduino)
		scanner.Scan()
		output := scanner.Text()

		if outputIsValid(output, re) {
			output := outputToMap(output)
			jsonString, err := json.Marshal(output)
			check(err)
			log.Output(1, fmt.Sprintf("%v", len(jsonString)))
			writeLineToDatabase(con, output)
		} else {
			log.Output(1, "Invalid output!")
		}
		time.Sleep(1000 * time.Millisecond)
	}
}
