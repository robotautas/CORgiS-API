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

	// connect to arduino
	mode := &serial.Mode{
		Parity:   serial.EvenParity,
		BaudRate: 115200,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	arduino, err := serial.Open(findArduino(), mode)
	check(err)
	defer arduino.Close()

	// main loop
	for {
		//TODO: paeksperimentuoti su output flush pries siunciant komanda
		_, err := arduino.Write([]byte("<GET_ALL;>"))
		// If err, reinitialize connection to device
		if err != nil {
			fmt.Printf("CONNECTION ERROR! %v", err)
			arduino, err = serial.Open(findArduino(), mode)
			check(err)
		}

		scanner := bufio.NewScanner(arduino)
		scanner.Scan()
		output := scanner.Text()
		if isValid(output, re) {
			println("valid!") // vietoje šito rašyti į duombazę
		}
		time.Sleep(1 * time.Second)
	}
}

// Validates arduino output against regex pattern and few other conditions.
func isValid(s string, re *regexp.Regexp) bool {
	if len(s) > 168 {
		if s[:4] == "V00=" &&
			strings.HasSuffix(s, ";") &&
			len(re.FindAll([]byte(s), -1)) >= 28 {
			log.Output(1, s)
			return true
		}
	}
	log.Output(1, "Incorrect data received!")
	return false
}

// Transforms validated output to map, for convenient writing to influxdb.
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

// Scan ports for arduino, return first port which serial number meets one of S/N's in serial_numbers.txt file.
// If arduino not found, it makes a program wait for device.
func findArduino() string {
	for {
		ports, err := enumerator.GetDetailedPortsList()
		check(err)
		for _, port := range ports {
			if port.IsUSB {
				for _, sn := range read_sn("serial_numbers.txt") {
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
	fmt.Printf("Reading serial numbers: %v\n", lines)
	return lines
}

// Helper function for dealing with errors
func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}
