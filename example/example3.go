package example

func test2() (int, int) {
	mutTwo := 2
	one, mutTwo := 1, 3

	return one, mutTwo
}
