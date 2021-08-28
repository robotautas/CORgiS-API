package main

import (
	"bufio"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial.v1"
)

func main() {
	re, err := regexp.Compile(`\w{3,4}=\d{1,4};`)
	check(err)
	test_str := "V00=0;V01=0;V02=0;V03=0;V04=0;V05=0;V06=0;V07=0;V08=0;T01=0;T02=1;T03=2;T04=1;T05=2;T06=0;T07=2;T08=0;P01=990;P02=990;P03=990;P04=990;P05=990;P06=990;P07=990;P08=990;S00=00;S01=00;PUMP=0;"
	convertToMap(test_str)

	for {
		isValid(rawOutput(), re)
		time.Sleep(1 * time.Second)
	}
}

// Sends GET_ALL command to arduino, returns raw output
func rawOutput() string {
	mode := &serial.Mode{
		DataBits: 8,
		StopBits: 1,
	}
	arduino, err := serial.Open(conn)
	check(err)
	_, err = arduino.Write([]byte("<GET_ALL;>"))
	check(err)
	scanner := bufio.NewScanner(arduino)
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

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}
