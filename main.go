package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	client "github.com/influxdata/influxdb1-client"
	"go.bug.st/serial.v1"
	"go.bug.st/serial.v1/enumerator"
)

var mode = &serial.Mode{
	Parity:   serial.EvenParity,
	BaudRate: 115200,
	DataBits: 8,
	StopBits: serial.OneStopBit,
}

var arduino, _ = serial.Open(findArduinoPort(), mode)

func main() {
	go DB_routine()
	http.HandleFunc("/set", setHandler)
	http.ListenAndServe(":9999", nil)
}

var re, _ = regexp.Compile(`\w{3,4}=\w{1,4};`)

// Aquires DB & Microcontroller connections, starts a loop constantly sending command to get all states of parameters in the board, and writes them to database
func DB_routine() {
	// regex pattern to validate raw output from arduino. Searches for strings like V00=254;
	r, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		println(err)
	}
	defer r.Close()

	con := getDBConnection()
	dur, ver, err := con.Ping()
	check(err)
	log.Printf("Connected to database! %v, %s", dur, ver)

	if !databaseDataExists(con) {
		createDatabaseData1h(con)
	}

	defer arduino.Close()
	var counter int64 = 0
	for {
		//TODO: paeksperimentuoti su output flush pries siunciant komanda
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
			r.Do("set", "values", string(jsonString))
			log.Output(1, string(jsonString))
			if counter%20 == 0 {
				writeLineToDatabase(con, output)
			}
			// perdaryti kad su counteriu istorinius duomenis rašytų tik kas 10 kartą. o i redis - kiekvieną
		}
		println(counter)
		counter++
		time.Sleep(50 * time.Millisecond)
	}
}

// Validates raw arduino output against regex pattern and few other conditions.
func outputIsValid(s string, re *regexp.Regexp) bool {
	log.Output(1, s)
	if len(s) > 168 {
		if s[:4] == "V00=" &&
			strings.HasSuffix(s, ";") &&
			len(re.FindAll([]byte(s), -1)) >= 28 {
			// log.Output(1, s)
			return true
		}
	}

	// log.Output(1, s)
	return false
}

// Transforms output like "V00=0;V01=0;V02=0; ... S01=00;PUMP=0;" to map, for convenient writing to influxdb.
func outputToMap(s string) map[string]interface{} {
	res := make(map[string]interface{})
	splitted_s := strings.Split(s, ";")
	for _, i := range splitted_s[:len(splitted_s)-1] {
		splitted_i := strings.Split(i, "=")
		// hex > dec
		if strings.HasPrefix(i, "V") {
			num, err := strconv.ParseInt(splitted_i[1], 16, 64)
			check(err)
			res[splitted_i[0]] = num
		} else {
			number, err := strconv.Atoi(splitted_i[1])
			check(err)
			res[splitted_i[0]] = number
		}
	}
	return res
}

// Scan ports for arduino, return first port whose serial number meets one of the S/Ns in serial_numbers.txt file.
// If arduino not found, it makes a program wait for device.
func findArduinoPort() string {
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
		// TODO: sutvarkyti loginimą, kad nuosekliai logintų pvz. tik su log paketu
		fmt.Println("\nArduino device not found. Check if connected!")
		time.Sleep(time.Second * 1)
	}
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
		RetentionPolicy: "autogen",
	}
	_, err := con.Write(bp)
	if err != nil {
		log.Fatal(err)
	}
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query().Get("param")
	value := r.URL.Query().Get("value")
	command := ""
	if strings.HasPrefix(param, "PUMP") {
		command = "<" + param + ";>"
	} else {
		command = "<SET_" + param + "=" + value + ";>"
	}
	_, err := arduino.Write([]byte(command))
	check(err)
	w.Write([]byte(command))
}

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}
