package main

func HasTrue(flags []bool) bool {
	for _, flag := range flags {
		if flag {
			return true
		}
	}

	return false
}

func AllTrue(flags []bool) bool {
	for _, flag := range flags {
		if !flag {
			return false
		}
	}

	return true
}
