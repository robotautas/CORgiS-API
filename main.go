package main

import (
	"bufio"
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

var mode = &serial.Mode{
	Parity:   serial.EvenParity,
	BaudRate: 115200,
	DataBits: 8,
	StopBits: serial.OneStopBit,
}

var arduino, _ = serial.Open(findArduinoPort(), mode)

var re, _ = regexp.Compile(`\w{3,4}=\w{1,4};`)

var VxxParams = []string{"V00", "V01", "V02", "V03", "V04", "V05", "V06", "V07", "V08"}
var TxxParams = []string{"T00", "T01", "T02", "T03", "T04", "T05", "T06", "T07", "T08", "T09", "T10", "T11", "T12", "T13", "T14", "T15", "T16", "T17", "T18", "T19", "T20", "T21"}
var pumpParams = []string{"PUMP_ON", "PUMP_OFF"}

var OS = runtime.GOOS

func main() {
	sendCommand("<SET_V00=0;>")
	flushRedis()
	initInstructionArrayRedis()
	go DB_routine()
	http.HandleFunc("/", RootHandler)
	http.HandleFunc("/set", SetHandler)
	http.HandleFunc("/setmulti", SetMultiHandler)
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
	printInfo("Connected to database! %v, %s", dur, ver)

	if !databaseDataExists(con) {
		createDatabaseData1h(con)
	}

	defer arduino.Close()

	for {
		_, err := arduino.Write([]byte("<GET_ALL;>"))
		// If err, reinitialize connection to device
		if err != nil {
			printError("CONNECTION ERROR! %v", err)
			arduino, err = serial.Open(findArduinoPort(), mode)
			check(err)
		}

		scanner := bufio.NewScanner(arduino)
		scanner.Scan()
		output := scanner.Text()

		if outputIsValid(output, re) {
			output := outputToMap(output)
			writeLineToDatabase(con, output)
			// jsonString, err := json.Marshal(output)
			// check(err)
			truncateOutput(&output, 3)
			printWhite("%v", output)

		} else {
			printWarning("Invalid output!")
		}
		time.Sleep(1000 * time.Millisecond)
	}
}
