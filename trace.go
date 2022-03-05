// trace.go
package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// trace2/trace.go
var goroutineSpace = []byte("goroutine ")
var mu sync.Mutex
var m = make(map[uint64]int)

func printTrace(id uint64, name, arrow string, indent int, file string, line int) {
	indents := ""
	for i := 0; i < indent; i++ {
		indents += "    "
	}
	files := strings.Split(file, "/")
	fileName := files[len(files)-1]
	fmt.Printf("g[%05d]:%s%s%s  [%s:%d] \n", id, indents, arrow, name, fileName, line)
}

func curGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	// Parse the 4707 out of "goroutine 4707 ["
	b = bytes.TrimPrefix(b, goroutineSpace)
	i := bytes.IndexByte(b, ' ')
	if i < 0 {
		panic(fmt.Sprintf("No space found in %q", b))
	}
	b = b[:i]
	n, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse goroutine ID out of %q: %v", b, err))
	}
	return n
}

func Trace() func() {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		panic("not found caller")
	}

	// 方法名
	funcName := runtime.FuncForPC(pc).Name()
	// funcName := runtime.FuncForPC(pc)

	gid := curGoroutineID()

	mu.Lock()
	indents := m[gid]
	m[gid] = indents + 1 // 将缩进层次+1 存入 map
	mu.Unlock()

	printTrace(gid, funcName, "->", indents+1, file, line)
	// fmt.Printf("gid:[%05d] enter:[%s] file:[%s] line:[%d] \n", gid, funcName, file, line)

	return func() {
		mu.Lock()
		indents := m[gid]
		m[gid] = indents - 1 // 将缩进层次-1 存入 map
		mu.Unlock()
		printTrace(gid, funcName, "<-", indents, file, line)
	}
}

// trace2/trace.go
func A1() {
	defer Trace()()
	B1()
}

func B1() {
	defer Trace()()
	C1()
}

func C1() {
	defer Trace()()
	D()
}

func D() {
	defer Trace()()
}

func A2() {
	defer Trace()()
	B2()
}
func B2() {
	defer Trace()()
	C2()
}
func C2() {
	defer Trace()()
	D()
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		A2()
		wg.Done()
	}()

	A1()
	wg.Wait()
}

