package main

import (
	"fmt"
	"strings"
	"syscall/js"
)

var subbed_stdout chan string = make(chan string, 10)

const LENGTH_OF_IMPORT = 6

func main() {
	var keep_program_alive chan byte = make(chan byte)
	var program_lifetime_token byte

	js.Global().Set("go_handle_pyodide_imports", js.FuncOf(pyodide_setup))
	js.Global().Set("pyodidelog_write", js.FuncOf(pyodide_write_stdout))
	js.Global().Set("pyodidelog_readline", js.FuncOf(pyodide_read_stdout))
	js.Global().Set("pyodidelog_readall", js.FuncOf(pyodide_readall_stdout))

	program_lifetime_token = <-keep_program_alive
	print(program_lifetime_token)
}

func pyodide_setup(this js.Value, val []js.Value) interface{} {
	var pythoncode string
	var lines_pythoncode []string
	var linecounter int
	var character_offeset int

	if len(val) == 0 {
		Info_Log(ERR_INVALID_ARGS)
		return nil
	}
	pythoncode = val[0].String()

	linecounter = 1
	for _, v := range pythoncode {
		if v == '\n' {
			linecounter++
		}
	}
	fmt.Printf("python code consists of %d lines\n", linecounter)

	lines_pythoncode = strings.Split(pythoncode, "\n")
	for _, line_of_code := range lines_pythoncode {
		println(line_of_code)
		for character_offeset = 0; character_offeset < len(line_of_code); character_offeset++ {
			if line_of_code[character_offeset] != ' ' {
				break
			}
		}

		//includes space afteward with 0-offset
		if len(line_of_code) < character_offeset+LENGTH_OF_IMPORT {
			continue
		}

		if line_of_code[character_offeset:character_offeset+LENGTH_OF_IMPORT] == "import" {
			println("found import line")
			println(line_of_code[character_offeset:])
			js.Global().Get("pyodide").Call("loadPackage", line_of_code[character_offeset+LENGTH_OF_IMPORT+1:])
		}
	}
	return nil
}

//general assumption in these pyodide out functions: go strings are more byte-precise than js strings, so the conversion should be lossless.
func pyodide_write_stdout(this js.Value, val []js.Value) interface{} {
	if len(val) != 0 {
		subbed_stdout <- val[0].String() + "\n"
		return nil
	}
	subbed_stdout <- "\n"
	return nil
}

func pyodide_read_stdout(this js.Value, val []js.Value) interface{} {
	if len(subbed_stdout) != 0 {
		return <-subbed_stdout
	} else {
		return js.ValueOf(0)
	}
	return nil
}

//returns slice of js strings.
func pyodide_readall_stdout(this js.Value, val []js.Value) interface{} {
	var stdout []interface{}
	stdout = make([]interface{}, 0, 10)

	if len(subbed_stdout) != 0 {
		<-subbed_stdout
		for len(subbed_stdout) != 0 {
			stdout = append(stdout, <-subbed_stdout)
		}
		return js.ValueOf(stdout)
	}
	js.Global().Call("alert", "read nil")
	return js.ValueOf(stdout)
}
