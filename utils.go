package main

func CalculateMaxNameLength(connNames []string) int {
	maxLength := 0
	for _, name := range connNames {
		if len(name) > maxLength {
			maxLength = len(name)
		}
	}
	return maxLength
}
