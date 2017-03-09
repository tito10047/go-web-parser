package misc

func SimilarTextPercent(first, second string) float64 {
	var firstLength, secondLength = len(first), len(second)
	sum := SimilarText(first, second)
	return float64(sum*200) / float64(firstLength+secondLength)
}

func SimilarText(first, second string) int {

	var pos1, pos2 = 0, 0
	var firstLength, secondLength = len(first), len(second)
	var l, max = 0, 0

	for p := 0; p < firstLength; p++ {
		for q := 0; q < secondLength; q++ {
			for l = 0; (p+l < firstLength) && (q+l < secondLength) && (first[p+l] == second[q+l]); l++ {
			}
			if l > max {
				max = l
				pos1 = p
				pos2 = q
			}
		}
	}

	sum := max

	if sum>0 {
		if pos1 > 0 && pos2 > 0 {
			sum += SimilarText(first[0:pos1], second[0:pos2])
		}

		if (pos1+max < firstLength) && (pos2+max < secondLength) {
			i := first[pos1+max:firstLength]
			i2 := second[pos2+max:secondLength]
			sum += SimilarText(
				i,
				i2)
		}
	}

	return sum

}
