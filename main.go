package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial.v1"
	"go.bug.st/serial.v1/enumerator"
)

func main() {
	// regex patern to validate raw output.
	re, err := regexp.Compile(`\w{3,4}=\d{1,4};`)
	check(err)

	findArduino()
	read_sn("serial_numbers.txt")

	// connect to arduino
	mode := &serial.Mode{
		Parity:   serial.EvenParity,
		BaudRate: 115200,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	arduino, err := serial.Open("/dev/ttyACM0", mode)
	check(err)
	defer arduino.Close()

	// main loop
	// for {
	isValid(rawOutput(arduino), re)
	time.Sleep(1 * time.Second)
	// }
}

// Sends GET_ALL command to arduino, returns raw output
func rawOutput(conn serial.Port) string {
	_, err := conn.Write([]byte("<GET_ALL;>"))
	check(err)
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	return scanner.Text()
}

// Validates arduino output against regex pattern and few other conditions
func isValid(s string, re *regexp.Regexp) bool {
	if s[:4] == "V00=" &&
		strings.HasSuffix(s, ";") &&
		len(s) > 168 &&
		len(re.FindAll([]byte(s), -1)) >= 28 {
		log.Output(1, s)
		return true
	}
	log.Output(1, "Incorrect data received!")
	return false
}

// Transforms validated output to map, for convenient writing to influxdb
func convertToMap(s string) map[string]int {
	res := make(map[string]int)
	splitted_s := strings.Split(s, ";")
	for _, i := range splitted_s[:len(splitted_s)-1] {
		splitted_i := strings.Split(i, "=")
		number, err := strconv.Atoi(splitted_i[1])
		check(err)
		res[splitted_i[0]] = number
	}
	return res
}

// Scan ports for arduino, return port
func findArduino() {
	ports, err := enumerator.GetDetailedPortsList()
	check(err)
	fmt.Printf("%v", ports)
	for _, port := range ports {
		fmt.Printf("Found port: %s\n", port.Name)
		if port.IsUSB {
			fmt.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
			fmt.Printf("   USB serial %s\n", port.SerialNumber)
		}
	}
}

//read serial numbers from file, remove whitespace return array
func read_sn(path string) []string {
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sn := scanner.Text()
		if strings.Contains(sn, " ") {
			sn = strings.ReplaceAll(sn, " ", "")
		}
		lines = append(lines, sn)
	}
	fmt.Printf("Readding serial numbers: %v\n", lines)
	return lines
}

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}
