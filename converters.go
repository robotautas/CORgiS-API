package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"
)

// structure to hold one Task of JSON instruction
// used for marshalling and unmarshalling JSON instructions
type Task struct {
	Vxx        map[string][][2]int `json:"Vxx,omitempty"`
	Txx        map[string]int      `json:"Txx,omitempty"`
	Pump       string              `json:"PUMP,omitempty"`
	Sleep      int                 `json:"Sleep"`
	StartTime  time.Time           `json:"start,omitempty"`
	FinishTime time.Time           `json:"finish,omitempty"`
}

var exampleJSON string = `[
    {
        "Vxx": {
            "V00": [[0, 1], [3, 0], [7, 1]],
            "V01": [[0, 0]]
        },
        "Txx": {
            "T01": 255
        },
        "PUMP": "ON",
        "Sleep": 10
    },
    {
        "Vxx":{"V00": [[3, 1]]},
        "Txx": {"T03": 300},
        "PUMP": "OFF",
        "Sleep": 30
    } 
]`

//for debug
var exampleStruct []Task = []Task{
	{
		Vxx: map[string][][2]int{
			"V00": {[2]int{5, 1}, [2]int{3, 0}},
		},
		Sleep: 20,
	},
	{
		Vxx: map[string][][2]int{
			"V00": {[2]int{5, 1}, [2]int{3, 0}},
			"V01": {[2]int{1, 0}, [2]int{3, 1}},
		},
		Txx: map[string]int{
			"T01": 255,
			"T02": 50,
		},
		Pump:  "OFF",
		Sleep: 20,
	},
}

// returns array of ones & zeros of length 8, representing a binary value of given parameter
func decToBinArray(d int) []int {
	if d < 0 || d > 255 {
		log.Panic("Value is less than zero or greater than 255!")
	}
	d64 := int64(d)
	binString := strconv.FormatInt(d64, 2)
	binArray := []int{}
	for _, ch := range binString {
		i, err := strconv.Atoi(string(ch))
		if err != nil {
			log.Panic("blogai")
		}
		binArray = append(binArray, i)
	}
	if len(binArray) < 8 {
		prependSlice := make([]int, 8-len(binArray))
		binArray = append(prependSlice, binArray...)
	}
	return binArray
}

// returns decimal value of given array, containing ones and zeros
func binArrayToDec(a []int) int {
	if len(a) < 8 {
		log.Panic("Must be array of length 8!")
	}
	num := 0
	for idx, byte := range a {
		if byte == 1 {
			num += int(math.Pow(10, float64((idx-7)*-1)))
		}
	}
	numString := strconv.Itoa(num)
	res, err := strconv.ParseInt(numString, 2, 10)
	if err != nil {
		log.Panic(err)
	}
	return int(res)
}

// given the
// current state of one of Vxx's represented by decimal value and
// what conditions have to be fulfilled for (sub)process to run -
// represented by value of type changes
// returns a dec Vxx number to be sent to the arduino device
func VxxChangeToDec(presentVal int, c changes) int {
	presentValToBin := decToBinArray(presentVal)
	newBinArr := append([]int(nil), presentValToBin...)
	for _, change := range c {
		if presentValToBin[change[0]] != change[1] {
			newBinArr[change[0]] = change[1]
		}
	}
	// some debugging prints, to be erased in future
	fmt.Printf("%v\n", c)
	fmt.Printf("%v\n", presentValToBin)
	fmt.Printf("%v\n", newBinArr)
	fmt.Printf("%v\n", binArrayToDec(newBinArr))
	return binArrayToDec(newBinArr)
}

func JSONToInstruction(s string) []Task {
	var Structure []Task
	err := json.Unmarshal([]byte(s), &Structure)
	if err != nil {
		message := fmt.Sprintf("Invalid JSON: %v", err)
		log.Output(1, message)
	}
	return Structure
}

func InstructionToJSON(sp []Task) string {
	res, err := json.Marshal(sp)
	if err != nil {
		message := fmt.Sprintf("Invalid JSON: %v", err)
		log.Output(1, message)
	}
	return string(res)
}

// given t Task
// and active Task from active tasks list
// func taskConflicts(tasks []Task) bool {
// 	for _, task := range tasks {
// 		for _, id := range getActiveTaskIds(){
// 			comparableStartTime:=
// 		}
// 	}
// 	return true
// }

func addTimeIntervals(tasks []Task) []Task {
	var modified []Task
	startTime := time.Now().Local()
	for _, task := range tasks {
		finishTime := startTime.Add(time.Second * time.Duration(task.Sleep))
		task.addTimeInterval(startTime, finishTime)
		modified = append(modified, task)
		startTime = finishTime
	}
	return modified
}

func (t *Task) addTimeInterval(s time.Time, f time.Time) {
	t.StartTime = s
	t.FinishTime = f
}
