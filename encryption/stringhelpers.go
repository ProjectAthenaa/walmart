package encryption

func stringFromCharCode(x int) string{
	return string(rune(x))
}

func substr(s string, start, end int) string {
	counter, startIdx := 0, 0
	for i := range s {
		if counter == start {
			startIdx = i
		}
		if counter == end {
			return s[startIdx:i]
		}
		counter++
	}
	return s[startIdx:]
}