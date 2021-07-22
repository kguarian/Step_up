package main

import (
	"sync"
	"syscall/js"
)

const LENGTH_OF_IMPORT = 6

var jglob js.Value = js.Global()

var wg sync.WaitGroup = sync.WaitGroup{}

func main() {
	var lifetime chan byte = make(chan byte)

	jglob.Set("Use_Before_Py", js.FuncOf(BeginInitializingPyodide))
	jglob.Set("Use_After_Py", js.FuncOf(FinishInitializingPyodide))
	jglob.Set("Go_ParseJson", js.FuncOf(parseJsonPackage))
	jglob.Set("GoGetRefCode", js.FuncOf(GoGetRefCode))
	//problem: all goroutines fall asleep.
	//solution 1: occupy a goroutine
	//solution 2: escape the limitation
	//solution 2 leads to bad code when solution 1 has not been tried.
	<-lifetime
}
