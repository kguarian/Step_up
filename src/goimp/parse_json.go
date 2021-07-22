package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"syscall/js"
)

var goOut *UnmarshallReadWriter = &UnmarshallReadWriter{InternalChannel: make([]string, 0, 10)}
var goErr *UnmarshallReadWriter = &UnmarshallReadWriter{InternalChannel: make([]string, 0, 10)}

var js_stdout_func js.Value
var js_stderr_func js.Value

var externalJson map[string]interface{}

//for JSON content, but is pretty general
type UnmarshallReadWriter struct {
	offset          int
	used_sz         int
	InternalChannel []string
}

func GoOut_Size(this js.Value, val []js.Value) (i interface{}) {
	i = js.ValueOf(goOut.used_sz - goOut.offset)
	return
}
func GoErr_Size(this js.Value, val []js.Value) (i interface{}) {
	i = js.ValueOf(goErr.used_sz - goErr.offset)
	return
}

func BeginInitializingPyodide(this js.Value, val []js.Value) (i interface{}) {
	js_stdout_func = jglob.Get("console").Get("log")
	js_stderr_func = jglob.Get("console").Get("error")
	jglob.Get("console").Set("log", js.FuncOf(Python_GoOut_Write))
	jglob.Get("console").Set("error", js.FuncOf(Python_GoErr_Write))

	jglob.Set("GoOutWrite", js.FuncOf(Python_GoOut_Write))
	jglob.Set("GoErrWrite", js.FuncOf(Python_GoErr_Write))
	jglob.Set("GoOutRead", js.FuncOf(Python_GoOut_Read))
	jglob.Set("GoErrRead", js.FuncOf(Python_GoErr_Read))
	jglob.Set("GoOutLength", js.FuncOf(GoOut_Size))
	jglob.Set("GoErrLength", js.FuncOf(GoErr_Size))
	return nil
}

func FinishInitializingPyodide(this js.Value, val []js.Value) (i interface{}) {
	var s string
	jglob.Get("console").Set("log", js_stdout_func)
	jglob.Get("console").Set("error", js_stderr_func)

	for goOut.offset != goOut.used_sz {
		s = goOut.InternalChannel[goOut.offset]
		fmt.Println(s)
		goOut.offset++
	}
	return
}

func Python_GoOut_Write(this js.Value, val []js.Value) (retInt interface{}) {
	if len(val) == 0 {
		goOut.Write([]byte("\n"))
		return nil
	}
	for _, v := range val {
		goOut.Write([]byte(v.String() + "\n"))
	}
	return
}

func Python_GoErr_Write(this js.Value, val []js.Value) (retInt interface{}) {
	if len(val) == 0 {
		goErr.Write([]byte("\n"))
		return nil
	}

	for _, v := range val {
		goErr.Write([]byte(v.String() + "\n"))
	}
	return
}

func Python_GoOut_Read(this js.Value, val []js.Value) (retInt interface{}) {
	var s string
	var err error
	s, err = goOut.Read()
	if err != nil {
		return js.ValueOf("error. Call stack trace for more info.")
	}
	return js.ValueOf(s)
}

func Python_GoErr_Read(this js.Value, val []js.Value) (retInt interface{}) {
	var s string
	var err error
	s, err = goErr.Read()
	if err != nil {
		return js.ValueOf("error. Call stack trace for more info.")
	}
	return js.ValueOf(s)
}

func (u *UnmarshallReadWriter) Read() (s string, err error) {
	if u.offset == u.used_sz || u.used_sz == 0 {
		return "error", errors.New("read error. Out of Bounds.")
	}
	s = u.InternalChannel[u.offset]
	u.offset++
	return s, nil

}

/*
	No error handling... Channel
*/
func (u *UnmarshallReadWriter) Write(b []byte) (n int, err error) {
	for u.offset > 9 {
		u.InternalChannel = u.InternalChannel[10:]

		u.offset -= 10
		u.used_sz -= 10
	}
	var s string = string(b)
	u.InternalChannel = append(u.InternalChannel, s)
	u.used_sz++
	return len(s), nil
}

func WriteIntermediate(u io.ReadWriter, b []byte) {
	u.Write(b)
}

func writeChallengePage(m map[string]interface{}) (s string) {
	fmt.Printf("Not Implemented Yet.\n")
	s = ""
	return
}
func describeMap(id string, m map[string]interface{}, depth int) {
	if len(m) == 0 {
		return
	}
	fmt.Fprintf(goOut, "struct {\n")
	for i, v := range m {
		fmt.Fprintf(goOut, "\t%s %s\n", i, reflect.TypeOf(v))
	}
	fmt.Fprintf(goOut, "}")
	fmt.Fprintf(goOut, "{ ")
	for i, v := range m {
		switch v.(type) {
		case string:
			fmt.Fprintf(goOut, "\"%s\":`%s`\n", i, strings.Replace(v.(string), "\n", "\\n", -1))
		case map[string]interface{}:
			fmt.Fprintf(goOut, "%s: map[string]interface{}{", id)
			for i, v := range v.(map[string]interface{}) {
				fmt.Fprintf(goOut, "\"%s\": %s\n", i, strings.Replace(fmt.Sprintf("%v", v), "\n", "\\n", -1))
			}
			fmt.Fprintf(goOut, "}\n")
		}
	}
	fmt.Fprintf(goOut, "\b\b}\n")
}

func describeSlice(id string, s []interface{}, depth int) {
	if len(s) == 0 {
		return
	}
	for i, v := range s {
		if v == nil {
			continue
		}
		switch v.(type) {
		case map[string]interface{}:
			describeMap(strconv.Itoa(i), v.(map[string]interface{}), depth+1)
		case []interface{}:
			describeSlice(strconv.Itoa(i), v.([]interface{}), depth+1)
		case string:
			fmt.Fprintf(goOut, "%d |> \n`%s`\n <| \n", i, strings.Replace(strings.Replace(strings.Replace(v.(string), "\\", "\\\\", -1), "\"", "\\\"", -1), "\n", "\\n", -1))
		default:
			fmt.Printf("%s|%s: [%d]%v\n",
				func(upper int) (rs string) {
					for i := 0; i < upper; i++ {
						rs += "|---"
					}
					return
				}(depth), id, i, reflect.TypeOf(v))
			return
		}
	}
}

//goal: parse json challenge to json struct containing the challenge specifications with standardized and type-appropriate in-memory data representation.

//TODO: parse json, have that code present
//func: ```expose_fields``` and types (helper)
//TODO: write struct for challenge, skipping empty/irrelevant lines, using expose_fields
//func: ```GetJson(af string) ChallengeSpec``` parse json from (const) server given abbreviated filename (i.e. chal1 instead of chal1.json), returning Dynamic ChallengeSpec (might require import for websocket).

/*Files will be small enough to fit in memory.*/
func GetFileContents(FileName string) (rs []byte, err error) {
	f, err := os.OpenFile(FileName, os.O_RDONLY, os.ModeDevice)
	if err != nil {
		return
	}
	rs, err = io.ReadAll(f)
	return
}

func expose_fields(jsonText []byte) (rs string, err error) {
	var unmarshalTarget map[string]interface{}
	err = json.Unmarshal(jsonText, &unmarshalTarget)
	if err != nil {
		return
	}
	describeMap("CHALLENGE.json", unmarshalTarget, 0)
	return
}

func GetInstructions(m map[string]interface{}) (retstring string) {
	retstring = m["block"].(map[string]interface{})["text"].(string)

	return
}

/*
	arg0: url source for json file
	TODO: this function
*/
func parseJsonPackage(this js.Value, val []js.Value) (ret interface{}) {
	var chan_resp chan *http.Response
	var chan_err chan error

	var url string
	if len(val) == 0 {
		return js.ValueOf(0)
	}

	externalJson = make(map[string]interface{})

	url = val[0].String()

	chan_resp, chan_err = make(chan *http.Response, 1), make(chan error, 1)

	go func(chan_resp chan *http.Response, chan_err chan error) {
		res, err := http.Get(url)
		if res != nil {
			if res.Body != nil {
				defer res.Body.Close()
				chan_resp <- res
			}
		}
		chan_err <- err

		if err != nil {
			fmt.Printf("HTTP request error. Exiting Goroutine Gracefully\n")
			return
		}

		r := bufio.NewReader(res.Body)
		x, err := r.ReadString(0)
		if err != nil && err != io.EOF {
			Errhandle_Log(err, err.Error())
			log.Panic(err)
			return
		}
		println(x)

		err = json.Unmarshal([]byte(x), &externalJson)
		describeMap("", externalJson, 0)
		fmt.Printf("%v\n", externalJson)
		block := externalJson["block"].(map[string]interface{})
		brief := block["text"].(string)
		js.Global().Get("instructions").Set("innerHTML", js.ValueOf(brief))
	}(chan_resp, chan_err)
	return
}

func GoGetRefCode(this js.Value, val []js.Value) (i interface{}) {
	var refcode string
	if externalJson == nil {
		return js.ValueOf(nil)
	}
	block := externalJson["block"].(map[string]interface{})
	source := block["source"].(map[string]interface{})
	refcode = source["code"].(string)
	refcode = strings.Replace(refcode, "\\n", "\n", -1)
	i = js.ValueOf(refcode)
	return
}
