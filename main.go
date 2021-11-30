package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	client "github.com/influxdata/influxdb1-client"
	"go.bug.st/serial.v1"
	"go.bug.st/serial.v1/enumerator"
)

var ids []int

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
	go DB_routine()
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/set", setHandler)
	http.HandleFunc("/getall", getHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/stop", stopHandler)
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
			log.Output(1, fmt.Sprintf("%v", string(jsonString)))
			writeLineToDatabase(con, output)
		} else {
			log.Output(1, "Invalid output!")
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

// Validates raw arduino output against regex pattern and few other conditions.
func outputIsValid(s string, re *regexp.Regexp) bool {
	if len(s) > 168 && len(s) < 215 {
		if s[:4] == "V00=" &&
			strings.HasSuffix(s, ";") &&
			len(re.FindAll([]byte(s), -1)) >= 28 {
			// log.Output(1, s)
			return true
		}
	}
	return false
}

// Transforms output like "V00=0;V01=0;V02=0; ... S01=00;PUMP=0;" to map, for convenient writing to influxdb.
func outputToMap(s string) map[string]interface{} {
	println("STRINGAS: ", s)
	res := make(map[string]interface{})
	splitted_s := strings.Split(s, ";")
	for _, i := range splitted_s[:len(splitted_s)-1] {
		splitted_i := strings.Split(i, "=")
		// hex -> dec
		if strings.HasPrefix(i, "V") {
			num, err := strconv.ParseInt(splitted_i[1], 16, 64)
			check(err)
			res[splitted_i[0]] = num
		} else {
			number, err := strconv.ParseInt(splitted_i[1], 10, 64)
			check(err)
			res[splitted_i[0]] = number
		}
	}
	return res
}

// Scan ports for arduino, return first port whose serial number meets one of the S/Ns in serial_numbers.txt file.
// If arduino not found, it makes a program wait for device.
func findArduinoPort() string {
	if OS != "windows" {
		for {
			ports, err := enumerator.GetDetailedPortsList()
			check(err)
			for _, port := range ports {
				if port.IsUSB {
					for _, sn := range getSerialNumbers("serial_numbers.txt") {
						if sn == port.SerialNumber {
							return port.Name
						}
					}
				}
			}
			fmt.Println("\nArduino device not found. Check if connected!")
			time.Sleep(time.Second * 1)
		}
	}
	return "COM10"
}

//Reads serial numbers from file, removes whitespace and returns array
func getSerialNumbers(path string) []string {
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sn := scanner.Text()
		// TODO use strings.TrimSpace!
		if strings.Contains(sn, " ") {
			sn = strings.ReplaceAll(sn, " ", "")
		}
		lines = append(lines, sn)
	}
	fmt.Printf("\nReading serial numbers: %v\n", lines)
	return lines
}

// Returns a database connection
func getDBConnection() *client.Client {
	host, err := url.Parse(fmt.Sprintf("http://%s:%d", "localhost", 8086))
	check(err)
	conf := client.Config{
		URL: *host,
	}
	con, err := client.NewClient(conf)
	check(err)
	return con
}

// Checks if database 'data' is present
func databaseDataExists(con *client.Client) bool {
	q := client.Query{
		Command: "show databases",
	}
	response, err := con.Query(q)
	check(err)
	for _, v := range response.Results[0].Series[0].Values {
		if v[0] == "data" {
			return true
		}
	}
	return false
}

// create database data with retention policy 1h
func createDatabaseData1h(con *client.Client) {
	q := client.Query{
		Command: "CREATE DATABASE \"data\" WITH DURATION 1h REPLICATION 1",
	}
	_, err := con.Query(q)
	check(err)
}

// write transformed outputs from arduino to database
func writeLineToDatabase(con *client.Client, output map[string]interface{}) {
	pt := client.Point{
		Measurement: "outputs",
		Fields:      output,
		Time:        time.Now()}
	pts := []client.Point{pt}
	bp := client.BatchPoints{
		Points:          pts,
		Database:        "data",
		RetentionPolicy: "autogen", // pabandyti koreguoti.
	}
	_, err := con.Write(bp)
	if err != nil {
		log.Fatal(err)
	}
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	param := r.URL.Query().Get("param")
	value := r.URL.Query().Get("value")

	// make sure, that param & value combination is valid
	if !URLParamValid(param) {
		w.Write([]byte("error: incorrect param!"))
		log.Output(1, "Invalid request!")
		return
	}
	if !URLValueValid(param, value) {
		w.Write([]byte("error: incorrect value!"))
		log.Output(1, "Invalid request!")
		return
	}

	// format and send a command to the device
	command := ""
	if strings.HasPrefix(param, "PUMP") {
		command = "<" + param + ";>"
		_, err := arduino.Write([]byte(command))
		check(err)
		log.Output(1, fmt.Sprintf("Command sent: %v", command))
	} else {
		command = "<SET_" + param + "=" + value + ";>"
		_, err := arduino.Write([]byte(command))
		if err != nil {
			w.Write([]byte("error: could not send a command to device, check if connected!"))
		}
		log.Output(1, fmt.Sprintf("Command sent: %v", command))
	}

	time.Sleep(30 * time.Millisecond)

	// format and send a response depending on parameter
	if stringInSlice(param, VxxParams) {
		valueToInt, err := strconv.ParseInt(value, 10, 64)
		check(err)
		for {
			answer := outputToMap(singleOutputRead())
			if answer[param] == valueToInt {
				jsonString, err := json.Marshal(answer)
				check(err)
				w.Write([]byte(jsonString))
				log.Output(1, "Valid response received.")
				break
			} else {
				logout := fmt.Sprintf("Response FAILED, %v != %v! Reading again..", answer[param], value)
				log.Output(1, logout)
				time.Sleep(20 * time.Millisecond)
			}
		}
	} else if stringInSlice(param, pumpParams) {
		for {
			answer := outputToMap(singleOutputRead())
			if param == "PUMP_ON" && answer["PUMP"] == int64(1) {
				jsonString, err := json.Marshal(answer)
				check(err)
				w.Write([]byte(jsonString))
				log.Output(1, "Valid response received.")
				break
			} else if param == "PUMP_OFF" && answer["PUMP"] == int64(0) {
				jsonString, err := json.Marshal(answer)
				check(err)
				w.Write([]byte(jsonString))
				log.Output(1, "Valid response received.")
				break
			} else {
				logout := fmt.Sprintf("Response FAILED! Param = '%v', pump value = '%v'", param, answer["PUMP"])
				log.Output(1, logout)
				time.Sleep(80 * time.Millisecond)
			}
		}
		// temperature is inertical, so it doesn't really need imediate response
	} else if stringInSlice(param, TxxParams) {
		answer := outputToMap(singleOutputRead())
		jsonString, err := json.Marshal(answer)
		check(err)
		w.Write([]byte(jsonString))
		log.Output(1, "Valid response received.")
	} else {
		w.Write([]byte("error: something unexpected happened"))
	}
	finish := time.Since(start)
	log.Output(1, fmt.Sprintf("Response took %v", finish))
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	answer := outputToMap(singleOutputRead())
	jsonString, err := json.Marshal(answer)
	check(err)
	w.Write([]byte(jsonString))
	log.Output(1, "Valid response received.")
}

// Accepts JSON string from request, starts a process routine
func startHandler(w http.ResponseWriter, r *http.Request) {
	c := make(chan int)

	// stopper <- "0"
	decoder := json.NewDecoder(r.Body)
	body := make([]map[string]string, 1, 1)
	err := decoder.Decode(&body)
	if err != nil {
		panic(err)
	}

	go process(c, body)
	cToString := strconv.Itoa(<-c)
	w.Write([]byte(cToString))

}

func stopHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		panic(err)
	}
	ids = append(ids, idInt)
	fmt.Printf("%v\n", ids)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(APIRules()))
}

func process(c chan int, r []map[string]string) {
	id := randInt(1000, 9999)
	c <- id
	for _, z := range r {
		times, err := strconv.ParseInt(z["sleep"], 10, 64)
		if err != nil {
			panic(err)
		}
		for i := 0; i < int(times); i++ {
			if contains(ids, id) {
				println("Stopped!")
				return
			}
			fmt.Printf("%v. %v: %v\n", i+1, z["param"], z["value"])
			time.Sleep(1 * time.Second)
		}
	}
}

// reads serial output untill it matches validation check
func singleOutputRead() string {
	for {
		arduino.ResetInputBuffer()
		_, err := arduino.Write([]byte("<GET_ALL;>"))
		check(err)
		// time.Sleep(30 * time.Millisecond)
		scanner := bufio.NewScanner(arduino)
		scanner.Scan()
		output := scanner.Text()
		if outputIsValid(output, re) {
			return output
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// checks if string is in slice
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// validates provided url value
func URLValueValid(p string, v string) bool {
	if len(v) > 0 {
		valueToInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return false
		}
		if stringInSlice(p, VxxParams) && valueToInt >= 0 && valueToInt < 256 {
			return true
		} else if stringInSlice(p, TxxParams) && valueToInt >= 0 && valueToInt < 1000 {
			return true
		}
	} else if stringInSlice(p, pumpParams) {
		return true
	}
	return false
}

// validates provided URL param
func URLParamValid(s string) bool {
	if stringInSlice(s, VxxParams) ||
		stringInSlice(s, TxxParams) ||
		stringInSlice(s, pumpParams) {
		return true
	}
	return false
}

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func randInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func APIRules() string {
	text := `This is an API part of middleware between graphitizer microcontroller and user interface.
API sends commands to microcontroller through HTTP GET requests.

SET request examples:
http://127.0.0.1:9999/set?param=V00&value=255
http://127.0.0.1:9999/set?param=T01&value=80
http://127.0.0.1:9999/set?param=PUMP_OFF

GET_ALL:
http://127.0.0.1:9999/getall
`
	return text
}
