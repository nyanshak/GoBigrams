/*
	Program: bigramTrain
	Usage: bigramTrain.exe -lm <lm out file> -text <input text file>
	
	reads input file, outputs unigram / bigram LM to output file
*/
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"bufio"
	"log"
	"math"
	"time"
	"sort"
	"flag"
)

type bigram struct {
	word_one,  word_two string
}

type Bigram []*bigram

// Methods required by sort.Interface.
func (s Bigram) Len() int {
	return len(s)
}
func (s Bigram) Less(i, j int) bool {
	if s[i].word_one < s[j].word_one {
		return true
	} else if s[i].word_one == s[j].word_one {
		return s[i].word_two < s[j].word_two
	}
	return false
}

func (s Bigram) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}


func FindSingletons(key string, bg Bigram, a map[bigram]int) ([]bigram) {
	singletons := make([]bigram, 0)

	i := rankBigrams(key, bg)

	if i != -1 {
		for ; i < len(bg); i++ {
			if bg[i].word_one == key {
				if a[*bg[i]] == 1 {
					singletons = append(singletons, *bg[i])
				}
			} else {
				break
			}
		}
	}
	return singletons
}

func GetNumerator(key string, bg Bigram, a map[bigram]float64) float64 {
	sum := float64(0)

	i := rankBigrams(key, bg)

	if i != -1 {
		for ; i < len(bg); i++ {
		    if (bg[i].word_one == key){
				sum += a[*bg[i]]
			} else {
				break
			}
		}
	}
	return 1.0 - sum
}

func rankStrings(key string, a []string) int {
	lo := 0
	hi := len(a) - 1
	for lo <= hi {
		// Key is in a[lo..hi] or not present.
		mid := lo + (hi - lo) / 2
		if key < a[mid] {
			hi = mid - 1
		} else if key > a[mid] {
			lo = mid + 1
		} else {
			return mid
		}
	}
	return -1;
}

func rankBigrams(key string, a Bigram) int {
	lo := 0
	hi := len(a) - 1
	for lo <= hi {
		mid := lo + (hi - lo) / 2
		if key < a[mid].word_one {
			hi = mid - 1
		} else if key > a[mid].word_one {
			lo = mid + 1
		} else {
			return mid
		}
	}
	return -1;
}

func BigramCount(counts map[bigram]int, s string) map[bigram]int {

	substrings := strings.Fields(s)

	for i := 0; i < len(substrings) - 1; i++ {
		temp_bigram := bigram{substrings[i], substrings[i + 1]}
		_, ok := counts[temp_bigram]

		if ok == true {
			counts[temp_bigram] += 1
		} else {
			counts[temp_bigram] = 1
		}
	}

	return counts
}

func GetMLESum(key string, bg Bigram, ug map[string]float64) float64 {

	i := rankBigrams(key, bg)
	sum := float64(0)

	if i == -1 {
		return sum
	} else {
		for ; i < len(bg); i++ {
			if bg[i].word_one == key {
				sum += ug[bg[i].word_two]
			} else {
				break
			}
		}
		return sum
	}
}

func Subtractγ(key string, γ float64, bgBigram Bigram, bg map[bigram]float64) {
	i := rankBigrams(key, bgBigram)

	if i != -1 {
		for ; i < len(bg); i++ {
			if bgBigram[i].word_one == key {
				bg[*bgBigram[i]] = bg[*bgBigram[i]] * γ
			}
		}
	}
}

func WordCount(s string) (map[string]int, int) {

	substrings := strings.Fields(s)
	wc := 0
	counts := make(map[string]int)

	for _, word := range substrings {
		wc++
		_, ok := counts[word]

		if ok == true {
			counts[word] += 1
		} else {
			counts[word] = 1
		}
	}

	return counts, wc
}

func trace(s string) (string, time.Time) {
	log.Println("START:", s)
	return s, time.Now()
}

func un(s string, startTime time.Time) {
	endTime := time.Now()
	log.Println("  END:", s, "ElapsedTime in seconds:", endTime.Sub(startTime))
}

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
}

var lm = flag.String("lm", "<lm>", "language model output file")
var text = flag.String("text", "<training_file>", "training file")

func main() {
	flag.Parse()
	if flag.NFlag() != 2 {
		flag.Usage()
		os.Exit(2)
	}

	filename := *text
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	lmOutFile, e := os.Create(*lm)
	if e != nil {
		fmt.Println(e)
		os.Exit(2)
	}
	s := string(content)
	unigrams_count, wordCount := WordCount(s)

	unigrams_probability := make(map[string]float64)
	for key, value := range unigrams_count {
		unigrams_probability[key] = float64(value) / float64(wordCount)
	}

	bigrams_count := make(map[bigram]int)

	file, _ := os.Open(filename)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		bigrams_count = BigramCount(bigrams_count, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	unigrams_sorted := make([]string, len(unigrams_count))
	for key, _ := range unigrams_count {
		unigrams_sorted = append(unigrams_sorted, key)
	}
	sort.Strings(unigrams_sorted)

	bigrams_sorted := make(Bigram, len(bigrams_count))

	i := 0
	for key, _ := range bigrams_count {
		bigrams_sorted[i] = &bigram{key.word_one, key.word_two}
		i++
	}

	sort.Sort(bigrams_sorted)

	bigrams_probability := make(map[bigram]float64)
	n1 := 0
	n2 := 0
	for key, value := range bigrams_count {
		if value == 1 {
			n1++
		} else if value == 2 {
			n2++
		}

		bigrams_probability[key] = float64(value)/float64(unigrams_count[key.word_one])
	}

	unigrams_backoff := make(map[string]float64)
	γ := .99

	fmt.Fprintln(lmOutFile, "unigrams:")
	for key, value := range unigrams_count {

		MLESumSecondWord := GetMLESum(key, bigrams_sorted, unigrams_probability)

		singletons := FindSingletons(key, bigrams_sorted, bigrams_count)
		α_h := float64(0)

		if len(singletons) == 0 {
			Subtractγ(key, γ, bigrams_sorted, bigrams_probability)
			α_h = (1.0 - γ) / (1.0 - MLESumSecondWord)
		} else {
			// replace bigrams_probability of v with Good-Turing estimate (for singletons)
			for _, v := range singletons {
				bigrams_probability[v] = (2.0 * float64(n2)) / (float64(n1) * float64(value))
			}
			α_h = GetNumerator(key, bigrams_sorted, bigrams_probability) / (1.0 - MLESumSecondWord)
		}
		unigrams_backoff[key] = α_h
		fmt.Fprintln(lmOutFile, α_h, key, math.Log2(unigrams_probability[key]))


	}

	fmt.Fprintln(lmOutFile, "bigrams:")
	for key, value := range bigrams_probability {
		fmt.Fprintln(lmOutFile, math.Log2(value), key.word_one, key.word_two)
	}

}
