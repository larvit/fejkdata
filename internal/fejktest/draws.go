package fejktest

func EachDraw(n int, draw func() string, check func(got string)) {
	for i := 0; i < n; i++ {
		check(draw())
	}
}

func DrawsInTurn(n int, draw func() string) []string {
	drawn := make([]string, 0, n)
	EachDraw(n, draw, func(got string) { drawn = append(drawn, got) })
	return drawn
}

func TallyDraws(n int, draw func() string) map[string]int {
	seen := map[string]int{}
	EachDraw(n, draw, func(got string) { seen[got]++ })
	return seen
}
