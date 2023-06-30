package main

func countSort(arr *[]int) {
	count := make(map[int]int)
	for _, el := range *arr {
		// this works because the default value for int is 0
		count[el]++
	}
	// sort keys
	keys := make([]int, 0, len(count))
	for k := range count {
		keys = append(keys, k)
	}
	quickSort(&keys) // quickSort lol
	// fill arr with sorted keys
	i := 0
	for _, k := range keys {
		for j := 0; j < count[k]; j++ {
			(*arr)[i] = k
			i++
		}
	}
}
