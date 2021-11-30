package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial.v1/enumerator"
)

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
