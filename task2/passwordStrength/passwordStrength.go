package passwordStrength

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"sort"
	"unicode"
)

var (
	// errors
	pathToDict			   = "dictionary.txt"
	dictionary             []string
	SameWithUserInputError = errors.New("Password too same with user inputs!")
	ExistInDictError       = errors.New("Password exist in dictionary!")
)

// password strength level
const (
	VERYWEAK   = 0
	WEAK       = 1
	REASONABLE = 2
	STRONG	   = 3
	VERYSTRONG = 4
)

type Config struct {
	UserInputs           []string
	minDistFromInputs    int
	RegExpReq            map[string]int
	NotExistInDictionary bool
}

func PasswordStrength(password string, config Config) (int, error) {
	maxScore := 0
	score := 0

	for i, regex := range config.RegExpReq {
		maxScore+=regex
		re := regexp.MustCompile(i)
		if re.MatchString(password) {
			score+=regex
		}
	}

	for _, input := range config.UserInputs {
		maxScore += len(input)
		d := dist(password, input)
		if d > config.minDistFromInputs {
			return 0, SameWithUserInputError
		}
		score += len(password) - d
	}

	if config.NotExistInDictionary && inDict(password) {
		return 0, ExistInDictError
	}

	strength := (100*score/maxScore + entropy(password)/4)/2

	switch  {
	case strength > 90:
		return VERYSTRONG, nil
	case strength > 75:
		return STRONG, nil
	case strength > 50:
		return REASONABLE, nil
	case strength > 25:
		return WEAK, nil
	default:
		return VERYWEAK, nil
	}
}

func entropy(password string) int {
	poolSize := 0
	digit, lower, upper, symbol := false, false, false, false
	for _, c := range password {
		if !digit && unicode.IsDigit(c) {
			digit = true
			poolSize += 10
		}
		if !lower && unicode.IsLower(c) {
			lower = true
			poolSize += 26
		}
		if !upper && unicode.IsUpper(c) {
			upper = true
			poolSize += 26
		}
		if !symbol && unicode.IsSymbol(c) {
			symbol = true
			poolSize += 30
		}
	}

	entropy := (math.Log2(float64(poolSize)))*float64(len(password))
	switch  {
	case entropy < 28:
		return VERYWEAK
	case entropy < 36:
		return WEAK
	case entropy < 60:
		return REASONABLE
	case entropy < 127:
		return STRONG
	default:
		return VERYSTRONG
	}
}

func LoadDict(){
	file, err := os.Open(pathToDict)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		dictionary = append(dictionary, scanner.Text())
	}
	sort.Strings(dictionary)
}

func inDict(pass string) bool {
	l,r := 0, len(dictionary)-1

	for l <= r {
		m := (l+r)/2
		if dictionary[m] > pass {
			l = m + 1
		}else if dictionary[m] < pass {
			r = m - 1
		}else{
			return true
		}
	}
	return false
}

func dist(a,b string) int {
	dp := [][]int{}
	for i:=0; i <= len(a); i++ {
		row := []int{}
		for j:=0; j <= len(b); j++ {
			row = append(row, 0)
		}
		dp = append(dp, row)
	}
	fmt.Println(len(dp), len(dp[0]))
	for i:=0; i <= len(a); i++ {
		for j:=0; j <= len(b); j++ {
			if min(i,j) == 0 {
				dp[i][j] = j+i
			}else if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1]
			}else{
				dp[i][j] = 1 + min(dp[i][j-1], min(dp[i-1][j], dp[i-1][j-1]))
			}
		}
	}
	return dp[len(a)][len(b)]
}

func min(a,b int) int {
	if a > b {
		return b
	}
	return a
}