package main

import (
	"strings"
	"time"

	"github.com/fatih/color"
)

var red = color.New(color.FgHiRed).PrintfFunc()

func timestamp() string {
	timeString := time.Now().Format("15:04:05.000")
	return timeString + "\u23F5"
}

func printError(format string, a ...interface{}) {
	if a != nil {
		red(timestamp()+" "+format+"\n", a...)
	} else {
		red(timestamp() + " " + format + "\n")
	}
}

var yellow = color.New(color.FgHiYellow).PrintfFunc()

func printWarning(format string, a ...interface{}) {
	if a != nil {
		yellow(timestamp()+" "+format+"\n", a...)
	} else {
		yellow(timestamp() + " " + format + "\n")
	}
}

var green = color.New(color.FgHiGreen).PrintfFunc()

func printDebug(format string, a ...interface{}) {
	if a != nil {
		green(timestamp()+" "+format+"\n", a...)
	} else {
		green(timestamp() + " " + format + "\n")
	}
}

var blue = color.New(color.FgHiCyan).PrintfFunc()

func printInfo(format string, a ...interface{}) {
	if a != nil {
		blue(timestamp()+" "+format+"\n", a...)
	} else {
		blue(timestamp() + " " + format + "\n")
	}
}

var white = color.New(color.FgWhite).PrintfFunc()

func printWhite(format string, a ...interface{}) {
	if a != nil {
		white(timestamp()+" "+format+"\n", a...)
	} else {
		white(timestamp() + " " + format + "\n")
	}
}

// gets truncated versions of board settings output
// 1 - Vxx
// 2 - Vxx, Txx
// 3 - Vxx, Txx, Pump
// 5 (or any other int) - All vals.
func truncateOutput(m *map[string]interface{}, level int) map[string]interface{} {
	switch level {
	case 1:
		for k := range *m {
			if strings.HasPrefix(k, "P") ||
				strings.HasPrefix(k, "T") ||
				strings.HasPrefix(k, "S") {
				delete(*m, k)
			}
		}
		return *m
	case 2:
		for k := range *m {
			if strings.HasPrefix(k, "P") ||
				strings.HasPrefix(k, "S") {
				delete(*m, k)
			}
		}
		return *m
	case 3:
		for k := range *m {
			if (strings.HasPrefix(k, "P") && !strings.HasPrefix(k, "PU")) ||
				strings.HasPrefix(k, "S") {
				delete(*m, k)
			}
		}
		return *m
	default:
		return *m
	}
}
