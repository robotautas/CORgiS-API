package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
)

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
	fmt.Printf("%v\n", c)
	fmt.Printf("%v\n", presentValToBin)
	fmt.Printf("%v\n", newBinArr)
	fmt.Printf("%v\n", binArrayToDec(newBinArr))
	return binArrayToDec(newBinArr)
}
