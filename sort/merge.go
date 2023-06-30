package main

import (
	"sync"

	"golang.org/x/exp/constraints"
)

//	func mergeSort(arr []T) []T {
//		if len(arr) <= 1 {
//			return arr
//		}
//		mid := len(arr) / 2
//		left := mergeSort(arr[:mid])
//		right := mergeSort(arr[mid:])
//		return merge(left, right)
//	}
//
//	func merge(left []T, right []T) []T {
//		result := make([]T, 0, len(left)+len(right))
//		l, r := 0, 0
//		for l < len(left) && r < len(right) {
//			if left[l] < right[r] {
//				result = append(result, left[l])
//				l++
//			} else {
//				result = append(result, right[r])
//				r++
//			}
//		}
//		result = append(result, left[l:]...)
//		result = append(result, right[r:]...)
//		return result
//	}
func mergeSort[T constraints.Ordered](arr *[]T) {
	buffer := make([]T, len(*arr))
	mergeSortRec(arr, &buffer, 0, len(*arr)-1)
}
func mergeSortRec[T constraints.Ordered](arr *[]T, buffer *[]T, start, end int) {
	if start >= end {
		return
	}
	mid := (start + end) / 2
	mergeSortRec(arr, buffer, start, mid)
	mergeSortRec(arr, buffer, mid+1, end)
	mergeRef(arr, buffer, start, mid, end)
}
func mergeRef[T constraints.Ordered](arr *[]T, buffer *[]T, start, mid, end int) {
	i, j, k := start, mid+1, start
	// fill buffer by merging the two sorted subarrays
	for i <= mid && j <= end {
		if (*arr)[i] < (*arr)[j] {
			(*buffer)[k] = (*arr)[i]
			i++
		} else {
			(*buffer)[k] = (*arr)[j]
			j++
		}
		k++
	}
	// fill the rest of the elements bc the loop exited
	// after one of the subarrays was exhausted
	for i <= mid {
		(*buffer)[k] = (*arr)[i]
		i++
		k++
	}
	for j <= end {
		(*buffer)[k] = (*arr)[j]
		j++
		k++
	}
	// copy sorted buffer back to arr
	for i := start; i <= end; i++ {
		(*arr)[i] = (*buffer)[i]
	}
}

func mergeSortPar[T constraints.Ordered](arr *[]T) {
	buffer := make([]T, len(*arr))
	mergeSortParRec(arr, &buffer, 0, len(*arr)-1)
}

const mergeSortParThreshold = 1000

func mergeSortParRec[T constraints.Ordered](arr *[]T, buffer *[]T, start, end int) {
	if start >= end {
		return
	}
	mid := (start + end) / 2
	if end-start < mergeSortParThreshold {
		mergeSortRec(arr, buffer, start, mid)
		mergeSortRec(arr, buffer, mid+1, end)
		mergeRef(arr, buffer, start, mid, end)
		return
	}
	var wg sync.WaitGroup
	var recSort = func(start, end int) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mergeSortRec(arr, buffer, start, end)
		}()
	}
	recSort(start, mid)
	recSort(mid+1, end)
	wg.Wait()
	mergeRef(arr, buffer, start, mid, end)
}
