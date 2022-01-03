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

// excecutes instruction represented as []Task
func excecuteInstruction(c chan int, tasks []Task) {
	// unique instruction id created and stored to package level list
	var instructionId int

	activeInstructionsLimit := 1000
	activeTasksLimit:= 3000

	if len(instructionIds) > activeInstructionsLimit {
		log.Output(1, fmt.Sprintf("%d active instructions limit exceeded, abandoning instruction.", activeInstructionsLimit))
		return
	}

	mutex.Lock()
	for {
		random := randInt(1000, 9999)
		if !intInSlice(instructionIds, random) {
			instructionId = random
			instructionIds = append(instructionIds, instructionId)
			break
		}
	}
	mutex.Unlock()

	// register and excecute tasks one by one
	for _, task := range tasks {
		var taskId int
		
		if len(getActiveTaskIds()) > activeTasksLimit {
			log.Output(1, fmt.Sprintf("%d active tasks limit exceeded, abandoning instruction.", activeInstructionsLimit))
			return
		}
		
		for {
			random := randInt(1000, 9999)
			if !idInRedisArray("activeTaskIds", random) {
				taskJSON := taskToJSON(task)
				storeActiveTask(random, taskJSON)
				break
			}
		}

		currentSettings := outputToMap(singleOutputRead())
		
		for k, v := range task.Vxx{
			currentSetting := int(currentSettings[k].(float64))
			changedSetting := vxxRequirementsToDec(currentSetting, v)
			command := fmt.Sprintf("<SET_%v=%v;>", k, changedSetting)
			_, err := arduino.Write([]byte(command))
			check(err)
		}

		for k, v:= range task.Txx{
			command := fmt.Sprintf("<SET_%v=%v;>", k, v)
			_, err := arduino.Write([]byte(command))
			check(err)
		}

		pumpCommand :=


	}
	






	


	c <- id
	for _, task := range tasks {
		
		
		
		
		// commands := subprocess["commands"]
		// for param, value := range commands.(map[string]interface{}) {
		// 	//here the commands are sent to arduino
		// 	value = int(value.(float64))
		// 	command := fmt.Sprintf("<SET_%v=%v;>", param, value)
		// 	println(command)
		// 	// arduino.ResetInputBuffer()
		// 	_, err := arduino.Write([]byte(command))
		// 	check(err)
		}
		// handle sleep between instructions
		sleep := int(subprocess["sleep"].(float64))
		fmt.Printf("sleeping for %vs\n", sleep)
		for i := 0; i < sleep; i++ {
			time.Sleep(1 * time.Second)
		}
	}
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
