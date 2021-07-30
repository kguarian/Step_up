//Some convenient Console interface functions:
//Color change, error logging

package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

//we want our log and error messages to stand out, so we color code them. These constants help with that.
const (
	GREEN  = "Green"
	RESET  = "Reset"
	RED    = "Red"
	YELLOW = "Yellow"
	BLUE   = "Blue"

	ANSIRESET  string = "\x1b[0m"
	ANSIRED    string = "\x1b[31m"
	ANSIGREEN  string = "\x1b[92m"
	ANSIYELLOW string = "\x1b[33m"
	ANSIBLUE   string = "\u001b[34m"
)

//In SetConsoleColor, we change the console color using this map as a lookup table.
var colormap map[string]string = map[string]string{RESET: ANSIRESET, RED: ANSIRED, GREEN: ANSIGREEN, YELLOW: ANSIYELLOW, BLUE: ANSIBLUE}

//We sometimes exit after errors. When we do, we call this function. Error messages are red here. These will be the last output from the program.
func Errhandle_Exit(err error, reason string) {
	var file string
	var line int
	fmt.Printf("%s:", reason)
	_, file, line, _ = runtime.Caller(1)
	if err != nil {
		SetConsoleColor(RED)
		fmt.Printf("\t%s %d\t failed: %v\n", file, line, err)
		SetConsoleColor(RESET)
		os.Exit(1)
	} else {
		SetConsoleColor(GREEN)
		fmt.Printf("\t %s %d \t successful.\n", file, line)
		SetConsoleColor(RESET)
	}
}

//We call this function for kind of trivial errors. It doesn't kill the program, error messages are yellow here.
func Errhandle_Log(err error, reason string) {
	var file string
	var line int
	fmt.Printf("%s:", reason)
	_, file, line, _ = runtime.Caller(1)
	if err != nil {
		SetConsoleColor(YELLOW)
		fmt.Printf("\t%s %d\t failed: %v\n", file, line, err)
		SetConsoleColor(RESET)
	} else {
		SetConsoleColor(GREEN)
		fmt.Printf("\t %s %d \t successful.\n", file, line)
		SetConsoleColor(RESET)
	}
}

func Info_Log(thingtoprint interface{}) {
	var file string
	var line int
	fmt.Printf("LOG MESSAGE:")
	SetConsoleColor(BLUE)
	_, file, line, _ = runtime.Caller(1)
	fmt.Printf("\t%s %d\t: %v\n", file, line, thingtoprint)
	SetConsoleColor(RESET)
}

//I don't see why I wouldn't just add the string literals into the source code... Will consider removing this function.
//TODO: Consider removal.
func Addcolorpair(key, ansicode string) {
	switch len(key) {
	case 0:
		_, fn, line, _ := runtime.Caller(1)
		log.Printf("%sline %d, function %s%s", ANSIRED, line, fn, ANSIRESET)
		log.Printf("length 0 color key (\"consoleio.go\" line 18\n")
	default:
		colormap[key] = ansicode
	}
}

//Sets consolo color according to the string:string map above.
func SetConsoleColor(color string) {
	for key, value := range colormap {
		if key == color {
			fmt.Printf("%s", value)
		}
	}
}
