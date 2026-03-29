package matching

import (
	"fmt"
	"math"

	"github.com/gbsto/daisy/db"
)

// Word association

func GetStarted() {
	fmt.Println("Matching package initialized")

	hostnameList, err := db.GetHostnames()
	if err != nil {
		fmt.Println("Error getting hostnames:", err)
		return
	}

	for i := range hostnameList {
		for j := i + 1; j < len(hostnameList); j++ {
			if i != j {
				checkCorrelation(hostnameList[i], hostnameList[j])
			}
		}
	}
}

func checkCorrelation(word1, word2 string) string {
	a := bigrams(word1)
	b := bigrams(word2)
	result := cosine(a, b)
	if result == 1.0 {
		fmt.Printf("%s and %s are identical\n", word1, word2)
		return "Identical"
	} else if result >= 0.99 {
		fmt.Printf("%s and %s are very similar (similarity: %.2f)\n", word1, word2, result)
		return "Nearly the Same"
	} else if result >= 0.7 {
		fmt.Printf("%s and %s are very similar (similarity: %.2f)\n", word1, word2, result)
		return "Very Similar"
	} else if result >= 0.3 {
		fmt.Printf("%s and %s are maybe related (similarity: %.2f)\n", word1, word2, result)
		return "Maybe Related"
	} else {
		fmt.Printf("%s and %s are unrelated\n", word1, word2)
	}
	return "Unrelated"
}

//Result range:
// 1.0 = identical
// 0.7+ = very similar
// 0.3–0.6 = maybe related
// <0.3 = unrelated

// Bigrams - Correlate two strings by their bigrams (pairs of characters) and calculate the cosine similarity between them.
// Hostnames are often similar, so this can be a useful way to find related hostnames.
func bigrams(s string) map[string]int {
	m := map[string]int{}
	for i := 0; i < len(s)-1; i++ {
		m[s[i:i+2]]++
	}
	return m
}

func cosine(a, b map[string]int) float64 {
	var dot, na, nb float64
	for k, va := range a {
		vb := b[k]
		dot += float64(va * vb)
		na += float64(va * va)
	}
	for _, vb := range b {
		nb += float64(vb * vb)
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// Best production signal:
// flag if:
//     Jaccard > 0.75
// AND Pearson > 0.80
// This combination is extremely reliable for:
// MAC randomization detection
// same-phone detection
// carried-together devices
