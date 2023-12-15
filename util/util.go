package util

import (
	"hash/fnv"
)

func Hash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))

	return int(h.Sum32())
}

/*
Supported glob-style patterns:

h?llo matches hello, hallo and hxllo
h*llo matches hllo and heeeello
h[ae]llo matches hello and hallo, but not hillo
h[^e]llo matches hallo, hbllo, ... but not hello
h[a-b]llo matches hallo and hbllo
*/

// regular regression
func PattenMatch(pattern, src string) bool {
	patLen, srcLen := len(pattern), len(src)

	if patLen == 0 {
		return srcLen == 0
	}

	if srcLen == 0 {
		for i := range pattern {
			if pattern[i] != '*' {
				return false
			}
		}
		return true
	}

	patPos, srcPos := 0, 0
	for patPos < patLen {
		switch pattern[patPos] {
		case '?': // h?llo
			srcPos++
		case '*': // h*llo
			for patPos < patLen && pattern[patPos] == '*' {
				patPos++
			}
			if patPos == patLen {
				return true
			}
			for srcPos < srcLen {
				for srcPos < srcLen && src[srcPos] != pattern[patPos] {
					srcPos++
				}
				if srcPos == srcLen {
					return patPos == patLen
				}
				if PattenMatch(pattern[patPos+1:], src[srcPos+1:]) {
					return true
				} else {
					srcPos++
				}
			}
			return false
		case '[':
			var not, close, match bool
			patPos++
			if patPos == patLen {
				return false
			}
			if pattern[patPos] == '^' {
				not = true
				patPos++
			}
			for patPos < patLen {
				if pattern[patPos] == ']' {
					close = true
					break
				} else if pattern[patPos] == '-' {
					start, end := pattern[patPos-1], pattern[patPos+1]
					if src[srcPos] >= start && src[srcPos] <= end {
						match = true
					}
					patPos++
				} else {
					if pattern[patPos] == src[srcPos] { //h[ae]llo and h[^e]llo
						match = true
					}
				}
				patPos++
			}
			if !close {
				return false
			}
			if not {
				match = !match
			}
			if !match {
				return false
			}
			srcPos++
		default:
			if pattern[patPos] != src[srcPos] {
				return false
			}
			srcPos++
		}
		patPos++

		// src has been cosumed
		if srcPos == srcLen {
			for patPos < patLen && pattern[patPos] == '*' {
				patPos++
			}
			break
		}
	}

	return patPos == patLen && srcPos == srcLen
}
