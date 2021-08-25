package util

func UintContains(arr []uint, find uint) bool {
	for _, a := range arr {
		if a == find {
			return true
		}
	}
	return false
}
