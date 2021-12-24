package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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

// excecutes instructions from JSON

// KAIP sustabdyti procesa:
// 1. reikia kaupti aktyvius subprocesus kintamajame arba redis
// 2. reikia identifikuoti, kelinta bita pakeicia komanda
// 3. procesui sustojus iteruoti per aktyvius (sub)procesus (isskyrus lokalu)
//    ir tikrinti, ar dar kur nors naudojamas lokalaus proceso pakeitimas
// 4. jeigu kazkur dar naudojamas pakeitimas, palikti kaip yra
//    jeigu ne, atstatyti i default(?)
// REIKES:
// 5. visos logikos binary to dec.
// 6. instrukcijos kaip tipo(struct)
//    ir visos tikrinimo logikos
// KAIP PADARYTI, KAD PROCESO SUSTABDYMAS NESIPYKTU SU PAVIENIAIS NUSTATYMAIS
func process(c chan int, r []Task) {
	// // unique active instruction id created
	// var id int
	// for {
	// 	random := randInt(1000, 9999)
	// 	if !idInRedisArray("activeTaskIds", random) {
	// 		id = random
	// 		break
	// 	}
	// }

	// c <- id
	// // iterating through JSON converted to native
	// for _, subprocess := range r {
	// 	commands := subprocess["commands"]
	// 	for param, value := range commands.(map[string]interface{}) {
	// 		//here the commands are sent to arduino
	// 		value = int(value.(float64))
	// 		command := fmt.Sprintf("<SET_%v=%v;>", param, value)
	// 		println(command)
	// 		// arduino.ResetInputBuffer()
	// 		_, err := arduino.Write([]byte(command))
	// 		check(err)
	// 	}
	// 	// handle sleep between instructions
	// 	sleep := int(subprocess["sleep"].(float64))
	// 	fmt.Printf("sleeping for %vs\n", sleep)
	// 	for i := 0; i < sleep; i++ {
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }
	log.Output(1, fmt.Sprintf("Process %d completed!", c))
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
