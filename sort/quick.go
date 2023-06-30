package main

import (
	"sync"

	"golang.org/x/exp/constraints"
)

//	func quickSort(arr []int) []int {
//		if len(arr) <= 1 {
//			return arr
//		}
//		pivot := arr[0]
//		left, right := make([]int, 0), make([]int, 0)
//		for i := 1; i < len(arr); i++ {
//			if arr[i] < pivot {
//				left = append(left, arr[i])
//			} else {
//				right = append(right, arr[i])
//			}
//		}
//		left, right = quickSort(left), quickSort(right)
//		return append(append(left, pivot), right...)
//	}
func quickSort[T constraints.Ordered](arr *[]T) {
	quickSortRec(arr, 0, len(*arr)-1)
}
func quickSortRec[T constraints.Ordered](arr *[]T, start, end int) {
	if start >= end {
		return
	}
	pivot := (*arr)[start]
	i, j := start, end
	for i < j {
		// find first el from the right that is smaller than pivot
		for i < j && (*arr)[j] >= pivot {
			j--
		}
		// find first el from the left that is bigger than pivot
		for i < j && (*arr)[i] <= pivot {
			i++
		}
		if i < j {
			// swap them
			(*arr)[i], (*arr)[j] = (*arr)[j], (*arr)[i]
		}
	}
	// swap pivot with new_pivot
	new_pivot := i
	(*arr)[start], (*arr)[new_pivot] = (*arr)[new_pivot], (*arr)[start]
	quickSortRec(arr, start, i-1)
	quickSortRec(arr, i+1, end)
}

func quickSortPar[T constraints.Ordered](arr *[]T) {
	quickSortParRec(arr, 0, len(*arr)-1)
}

const quickSortParThreshold = 1000

func quickSortParRec[T constraints.Ordered](arr *[]T, start, end int) {
	if start >= end {
		return
	}
	pivot := (*arr)[start]
	i, j := start, end
	for i < j {
		// find first el from the right that is smaller than pivot
		for i < j && (*arr)[j] >= pivot {
			j--
		}
		// find first el from the left that is bigger than pivot
		for i < j && (*arr)[i] <= pivot {
			i++
		}
		if i < j {
			// swap them
			(*arr)[i], (*arr)[j] = (*arr)[j], (*arr)[i]
		}
	}
	// swap pivot with new_pivot
	new_pivot := i
	(*arr)[start], (*arr)[new_pivot] = (*arr)[new_pivot], (*arr)[start]

	// if the subarray is small enough, don't spawn goroutines
	if end-start < quickSortParThreshold {
		quickSortRec(arr, start, i-1)
		quickSortRec(arr, i+1, end)
		return
	}
	// waitgroup waits for both goroutines to finish
	var wg sync.WaitGroup
	var recSort = func(start, end int) {
		wg.Add(1) // add has to be here so it causes wg.Wait() to actually wait
		go func() {
			defer wg.Done()
			quickSortParRec(arr, start, end)
		}()
	}
	recSort(start, i-1)
	recSort(i+1, end)
	wg.Wait()
}
