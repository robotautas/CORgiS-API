package main

import (
	"time"

	"github.com/fatih/color"
)

var red = color.New(color.FgHiRed).PrintfFunc()

func timestamp() string {
	timeString := time.Now().Format("15:04:05.000")
	return timeString + " "
}

func printError(format string, a ...interface{}) {
	if a != nil {
		red(timestamp()+" "+format+"\n", a)
	} else {
		red(timestamp() + " " + format + "\n")
	}
}

var yellow = color.New(color.FgHiYellow).PrintfFunc()

func printWarning(format string, a ...interface{}) {
	if a != nil {
		yellow(timestamp()+" "+format+"\n", a)
	} else {
		yellow(timestamp() + " " + format + "\n")
	}
}

var green = color.New(color.FgHiGreen).PrintfFunc()

func printDebug(format string, a ...interface{}) {
	if a != nil {
		green(timestamp()+" "+format+"\n", a)
	} else {
		green(timestamp() + " " + format + "\n")
	}
}

var blue = color.New(color.FgHiCyan).PrintfFunc()

func printInfo(format string, a ...interface{}) {
	if a != nil {
		blue(timestamp()+" "+format+"\n", a)
	} else {
		blue(timestamp() + " " + format + "\n")
	}
}
