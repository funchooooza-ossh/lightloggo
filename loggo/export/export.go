package main

/*
#include <stdint.h>
*/
import "C"

import (
	"encoding/json"
	"funchooooza-ossh/loggo/core"
	"sync"
)

var (
	loggerStore         = map[uintptr]*core.Logger{}
	routeStore          = map[uintptr]*core.RouteProcessor{}
	currentID   uintptr = 1
	storeMu     sync.Mutex
)

func makeID() uintptr {
	storeMu.Lock()
	defer storeMu.Unlock()
	id := currentID
	currentID++
	return id
}

//export NewLoggerWithSingleRoute
func NewLoggerWithSingleRoute(routeID C.uintptr_t) C.uintptr_t {
	storeMu.Lock()
	defer storeMu.Unlock()

	route := routeStore[uintptr(routeID)]
	logger := core.NewLogger(route)
	id := makeID()
	loggerStore[id] = logger
	return C.uintptr_t(id)
}

//export Logger_InfoWithFields
func Logger_InfoWithFields(loggerID C.uintptr_t, msg *C.char, fieldsJSON *C.char) {
	storeMu.Lock()
	logger := loggerStore[uintptr(loggerID)]
	storeMu.Unlock()

	goMsg := C.GoString(msg)
	jsonStr := C.GoString(fieldsJSON)

	var fields map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &fields)
	if err != nil {
		logger.Info(goMsg, nil) // fallback
		return
	}

	logger.Info(goMsg, fields)
}

//export FreeLogger
func FreeLogger(loggerID C.uintptr_t) {
	storeMu.Lock()
	defer storeMu.Unlock()
	delete(loggerStore, uintptr(loggerID))
}

func main() {}
