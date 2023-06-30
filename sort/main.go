package main

func main() {
	funcs := map[string]func(*[]int){
		"quick":          quickSort[int],
		"parallel quick": quickSortPar[int],
		"merge":          mergeSort[int],
		"parallel merge": mergeSortPar[int],
		"count":          countSort,
	}
	raceFuncs(funcs)
}
