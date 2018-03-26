package main

func IsAppKey(key string) bool {
	if len(key) == 0 ||
		len(key) > 32 {
		return false
	}
	for _, r := range key {
		if r < '0' ||
			r > '9' && r < 'A' ||
			r > 'Z' && r < 'a' ||
			r > 'z' {
			return false
		}
	}
	return true
}
