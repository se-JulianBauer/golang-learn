package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

func randomArr(len, rang int) []int {
	arr := make([]int, len)
	for i := 0; i < len; i++ {
		arr[i] = rand.Intn(rang)
	}
	return arr
}
func timeClosure(f func(), timeout time.Duration) (bool, time.Duration) {
	start := time.Now()
	ch := make(chan bool, 1) // buffered to prevent goroutine leak
	go func() {
		f()
		ch <- true
	}()
	finished := false
	select {
	case <-ch:
		finished = true
	case <-time.After(timeout):
	}
	end := time.Now()
	return finished, end.Sub(start)
}

func checkSort(arr *[]int) bool {
	for i := 1; i < len(*arr); i++ {
		if (*arr)[i] < (*arr)[i-1] {
			return false
		}
	}
	return true
}

func testSortOnArr(arr *[]int, f func(*[]int)) (int, bool) {
	arrCopy := make([]int, len(*arr))
	copy(arrCopy, *arr)
	finished, time := timeClosure(func() {
		f(&arrCopy)
	}, 5*time.Second)
	// get float64 of time in seconds
	timeSec := int(time.Milliseconds())
	success := checkSort(&arrCopy) && finished
	return timeSec, success
}

type sortResult struct {
	time    int
	success bool
	name    string
}
type sortLengthList []sortResult

func (s sortLengthList) Len() int {
	return len(s)
}
func (s sortLengthList) Less(i, j int) bool {
	// lack of success gives lowest priority
	if s[i].success && !s[j].success {
		return true
	}
	if !s[i].success && s[j].success {
		return false
	}
	// if both success or both not success, sort by time
	return s[i].time < s[j].time
}
func (s sortLengthList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func raceFuncsOnArr(arr *[]int, arrName string, fs map[string]func(*[]int)) {
	res := make([]sortResult, 0, len(fs))
	for name, f := range fs {
		time, success := testSortOnArr(arr, f)
		res = append(res, sortResult{time, success, name})
	}
	sort.Sort(sortLengthList(res))
	fmt.Println("\nResults for", arrName, "array:")
	for _, r := range res {
		fmt.Printf("[%s sort] time: %d ms, array: %s, success: %t\n", r.name, r.time, arrName, r.success)
		if !r.success && len(*arr) <= 20 {
			fmt.Println("arr:", *arr)
		}

	}
}

func raceFuncs(funcs map[string]func(*[]int)) {
	arrs := map[string][]int{
		"1m":            randomArr(1_000_000, 1_000_000),
		"1m low range":  randomArr(1_000_000, 100),
		"1m high range": randomArr(1_000_000, 10_000_000),
		"5m":            randomArr(5_000_000, 5_000_000),
		"5m low range":  randomArr(5_000_000, 100),
		"5m high range": randomArr(5_000_000, 10_000_000),
	}
	// sort keys
	keys := make([]string, 0, len(arrs))
	for k := range arrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, arrName := range keys {
		arr := arrs[arrName]
		raceFuncsOnArr(&arr, arrName, funcs)
	}

}
