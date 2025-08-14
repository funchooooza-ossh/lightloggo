package main

/*
#include <stdint.h>
#include <stddef.h>
*/
import "C"

import (
	"funchooooza-ossh/loggo/core"
	"funchooooza-ossh/loggo/core/formatter"
	"funchooooza-ossh/loggo/core/writer"
	"sync"
	"unsafe"
)

var (
	loggerStore      = map[uintptr]*core.Logger{}
	routeStore       = map[uintptr]*core.RouteProcessor{}
	formatStyleStore = map[uintptr]*core.FormatStyle{}
	formatterStore   = map[uintptr]core.FormatProcessor{}
	writerStore      = map[uintptr]core.WriteProcessor{}
	dependencyStore  = map[uintptr][]uintptr{}

	currentID uintptr = 1
	storeMu   sync.Mutex
)

func makeID() uintptr {
	storeMu.Lock()
	defer storeMu.Unlock()

	id := currentID
	currentID++
	return id
}

//export NewLoggerWithRoutes
func NewLoggerWithRoutes(routeIDs *C.uintptr_t, count C.int) C.uintptr_t {
	storeMu.Lock()
	defer storeMu.Unlock()

	routes := make([]*core.RouteProcessor, 0, int(count))

	// конвертация C-массива → Go-слайс
	slice := unsafe.Slice(routeIDs, count)

	for i := 0; i < int(count); i++ {
		r := routeStore[uintptr(slice[i])]
		if r != nil {
			routes = append(routes, r)
		}
	}

	logger := core.NewLogger(routes...)
	id := makeID()
	loggerStore[id] = logger
	return C.uintptr_t(id)
}

//export NewRouteProcessor
func NewRouteProcessor(formatterID, writerID C.uintptr_t, level C.uintptr_t) C.uintptr_t {
	storeMu.Lock()
	defer storeMu.Unlock()

	formatter := formatterStore[uintptr(formatterID)]
	writer := writerStore[uintptr(writerID)]

	route := core.NewRouteProcessor(formatter, writer, core.LogLevel(level))

	id := makeID()
	routeStore[id] = route

	dependencyStore[id] = []uintptr{uintptr(formatterID), uintptr(writerID)}

	return C.uintptr_t(id)
}

//export NewStdoutWriter
func NewStdoutWriter() C.uintptr_t {
	storeMu.Lock()
	defer storeMu.Unlock()

	writer := &writer.StdoutWriter{}
	id := makeID()
	writerStore[id] = writer
	return C.uintptr_t(id)
}

//export NewFileWriter
func NewFileWriter(path *C.char, maxSizeMB C.long, maxBackups C.int, interval *C.char, compress *C.char) C.uintptr_t {
	storeMu.Lock()
	defer storeMu.Unlock()

	goPath := C.GoString(path)
	goInterval := writer.RotateInterval(C.GoString(interval))

	var goCompress *writer.Compress
	if compress != nil {
		c := writer.Compress(C.GoString(compress))
		goCompress = &c
	}

	writer, err := writer.NewFileWriter(
		goPath,
		int64(maxSizeMB),
		int(maxBackups),
		goInterval,
		goCompress,
	)
	if err != nil {
		return 0
	}

	id := makeID()
	writerStore[id] = writer
	return C.uintptr_t(id)
}

//export NewTextFormatter
func NewTextFormatter(styleID C.uintptr_t, maxDepth C.int) C.uintptr_t {
	storeMu.Lock()
	defer storeMu.Unlock()

	var style *core.FormatStyle
	if s, ok := formatStyleStore[uintptr(styleID)]; ok {
		style = s
	}

	depth := int(maxDepth)

	formatter := formatter.NewTextFormatter(style, &depth)

	id := makeID()
	formatterStore[id] = formatter
	dependencyStore[id] = []uintptr{uintptr(styleID)}

	return C.uintptr_t(id)
}

//export NewJsonFormatter
func NewJsonFormatter(styleId C.uintptr_t, maxDepth C.int) C.uintptr_t {
	storeMu.Lock()
	defer storeMu.Unlock()

	var style *core.FormatStyle
	if s, ok := formatStyleStore[uintptr(styleId)]; ok {
		style = s
	}

	depth := int(maxDepth)

	formatter := formatter.NewJsonFormatter(style, &depth)

	id := makeID()
	formatterStore[id] = formatter
	dependencyStore[id] = []uintptr{uintptr(styleId)}

	return C.uintptr_t(id)
}

//export NewFormatStyle
func NewFormatStyle(colorKeys, colorValues, colorLevel C.uintptr_t, keyColor, valueColor, reset *C.char) C.uintptr_t {
	storeMu.Lock()
	defer storeMu.Unlock()

	style := &core.FormatStyle{
		ColorKeys:   colorKeys != 0,
		ColorValues: colorValues != 0,
		ColorLevel:  colorLevel != 0,
		KeyColor:    C.GoString(keyColor),
		ValueColor:  C.GoString(valueColor),
		Reset:       C.GoString(reset),
	}

	id := makeID()
	formatStyleStore[id] = style
	return C.uintptr_t(id)
}

//export NewLoggerWithSingleRoute
func NewLoggerWithSingleRoute(routeID C.uintptr_t) C.uintptr_t {
	storeMu.Lock()
	defer storeMu.Unlock()

	route := routeStore[uintptr(routeID)]
	logger := core.NewLogger(route)

	id := makeID()
	loggerStore[id] = logger
	dependencyStore[id] = []uintptr{uintptr(routeID)}

	return C.uintptr_t(id)
}

func LogN(loggerId C.uintptr_t, level core.LogLevel,
	msg *C.char, msgLen C.size_t,
	fieldsJSON *C.char, fieldsLen C.size_t,
) {
	storeMu.Lock()
	lg := loggerStore[uintptr(loggerId)]
	storeMu.Unlock()
	if lg == nil {
		return
	}

	if !lg.AnyRouteShouldLog(level) {
		return
	}
	rts := lg.RoutesSnapshot()

	var goMsg []byte
	if msg != nil && msgLen > 0 {
		goMsg = C.GoBytes(unsafe.Pointer(msg), C.int(msgLen))
	}
	var fieldsRaw []byte
	if fieldsJSON != nil && fieldsLen > 0 {
		fieldsRaw = C.GoBytes(unsafe.Pointer(fieldsJSON), C.int(fieldsLen))
	}

	record := core.LogRecordRaw{
		Level:   level,
		Message: goMsg,
		Fields:  fieldsRaw,
	}

	for _, r := range rts {
		if r != nil && r.ShouldLog(level) {
			r.Enqueue(record)
		}
	}

}

//export Logger_Trace
func Logger_Trace(loggerId C.uintptr_t, msg *C.char, msgLen C.size_t,
	fields *C.char, fieldsLen C.size_t) {
	LogN(loggerId, core.Trace, msg, msgLen, fields, fieldsLen)
}

//export Logger_Debug
func Logger_Debug(loggerId C.uintptr_t, msg *C.char, msgLen C.size_t,
	fields *C.char, fieldsLen C.size_t) {
	LogN(loggerId, core.Debug, msg, msgLen, fields, fieldsLen)
}

//export Logger_Info
func Logger_Info(loggerId C.uintptr_t, msg *C.char, msgLen C.size_t,
	fields *C.char, fieldsLen C.size_t) {
	LogN(loggerId, core.Info, msg, msgLen, fields, fieldsLen)
}

//export Logger_Warning
func Logger_Warning(loggerId C.uintptr_t, msg *C.char, msgLen C.size_t,
	fields *C.char, fieldsLen C.size_t) {
	LogN(loggerId, core.Warning, msg, msgLen, fields, fieldsLen)
}

//export Logger_Error
func Logger_Error(loggerId C.uintptr_t, msg *C.char, msgLen C.size_t,
	fields *C.char, fieldsLen C.size_t) {
	LogN(loggerId, core.Error, msg, msgLen, fields, fieldsLen)
}

//export Logger_Exception
func Logger_Exception(loggerId C.uintptr_t, msg *C.char, msgLen C.size_t,
	fields *C.char, fieldsLen C.size_t) {
	LogN(loggerId, core.Exception, msg, msgLen, fields, fieldsLen)

}

func freeComponentAndDeps(id uintptr) {
	if deps, ok := dependencyStore[id]; ok {
		for _, depId := range deps {
			freeComponentAndDeps(depId)
		}
	}

	callCloseIfAvailable(id)
	deleteFromAllStores(id)

	delete(dependencyStore, id)
}

func callCloseIfAvailable(id uintptr) {
	if router, ok := routeStore[id]; ok {
		router.Close()
	}
	if logger, ok := loggerStore[id]; ok {
		logger.Close()
	}
}

func deleteFromAllStores(id uintptr) {
	delete(loggerStore, id)
	delete(routeStore, id)
	delete(formatStyleStore, id)
	delete(formatterStore, id)
	delete(writerStore, id)
}

//export FreeLogger
func FreeLogger(loggerID C.uintptr_t) {
	storeMu.Lock()
	defer storeMu.Unlock()
	freeComponentAndDeps(uintptr(loggerID))
}

//export Logger_Close
func Logger_Close(loggerID C.uintptr_t) {
	storeMu.Lock()
	logger := loggerStore[uintptr(loggerID)]
	storeMu.Unlock()

	logger.Close()
}

func main() {}
