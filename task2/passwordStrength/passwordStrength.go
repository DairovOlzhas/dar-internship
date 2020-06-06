package passwordStrength

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"unicode"
)

var (
	pathToDict			   = "dictionary.txt"
)

// password strength levels
const (
	VeryWeak   = 0
	Weak       = 1
	Reasonable = 2
	Strong     = 3
	VeryStrong = 4
)
// NewPasswordStrength returns passwordStrength class
func NewPasswordStrength(config Config) PasswordStrength {
	ps := PasswordStrength{
		config:     config,
		pathToDict: pathToDict,
	}
	return ps
}

// Calc returns password strength.
// 0 - Very Weak; 	might keep out family members
// 1 - Weak; 		should keep out most people, often good for desktop login passwords
// 2 - Reasonable; 	fairly secure passwords for network and company passwords
// 3 - Strong; 		can be good for guarding financial information
// 4 - Very Strong; often overkill.
func (ps PasswordStrength) Calc(password string) (int, error) {
	maxScore := 0
	score := 0

	// For loop calculates how far the password is from user inputs.
	for _, input := range ps.config.UserInputs {
		maxScore += len(input)
		d := dist(password, input)

		// Has password required distance from user input?
		if d < ps.config.MinEditDistFromInputs {
			return Weak, nil
		}
		score += len(password) - d
	}

	for regex, points := range ps.config.RegExps {
		re := regexp.MustCompile(regex)
		maxScore+=points
		if re.MatchString(password) { // if password matches regexp
			score += points
		} else if points == 0 { // if password doesn't match regexp and it's must required regexp
			return Weak, nil
		} // if password doesn't match regexp and regexp is not required then nothing happens
	}

	if ps.config.SearchInDictionary {
		if !ps.dictLoaded {
			err := ps.loadDict()
			if err != nil {
				return 0, err
			}
		}
		if ps.inDict(password) {
			return Weak, nil
		}
	}

	entropy := ps.entropy(password)
	strength := (100*score/maxScore + 100*entropy/4)/2 // calculating percentage

	switch  {
	case strength > 90:
		return VeryStrong, nil
	case strength > 75:
		return Strong, nil
	case strength > 50:
		return Reasonable, nil
	case strength > 25:
		return Weak, nil
	default:
		return VeryWeak, nil
	}
}

// Loads dictionary from txt file located in pathToDict
func (ps PasswordStrength) loadDict() error {
	file, err := os.Open(pathToDict)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		ps.dictionary = append(ps.dictionary, scanner.Text())
	}
	sort.Strings(ps.dictionary)

	return nil
}

// inDict returns password existence in dictionary
func (ps PasswordStrength) inDict(pass string) bool {
	l,r := 0, len(ps.dictionary)-1

	for l <= r {
		m := (l+r)/2
		if ps.dictionary[m] > pass {
			r = m - 1
		}else if ps.dictionary[m] < pass {
			l = m + 1
		}else{
			return true
		}
	}
	return false
}

// entropy returns password strength based on entropy.
// Password entropy is a measurement of how unpredictable a password is.
// More information at https://www.pleacher.com/mp/mlessons/algebra/entropy2.html
func (ps PasswordStrength) entropy(password string) int {
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
		return VeryWeak
	case entropy < 36:
		return Weak
	case entropy < 60:
		return Reasonable
	case entropy < 127:
		return Strong
	default:
		return VeryStrong
	}
}


// Returns edit distance between two strings.
// Edit distance is minimum number of edits(operations)
// required to convert 'str1' into 'str2'. Operations:
// - Insert
// - Remove
// - Replace
func dist(str1, str2 string) int {
	dp := [][]int{}
	for i:=0; i <= len(str1); i++ {
		row := []int{}
		for j:=0; j <= len(str2); j++ {
			row = append(row, 0)
		}
		dp = append(dp, row)
	}
	fmt.Println(len(dp), len(dp[0]))
	for i:=0; i <= len(str1); i++ {
		for j:=0; j <= len(str2); j++ {
			if min(i,j) == 0 {
				dp[i][j] = j+i
			}else if str1[i-1] == str2[j-1] {
				dp[i][j] = dp[i-1][j-1]
			}else{
				dp[i][j] = 1 + min(dp[i][j-1], min(dp[i-1][j], dp[i-1][j-1]))
			}
		}
	}
	return dp[len(str1)][len(str2)]
}

func min(a,b int) int {
	if a > b {
		return b
	}
	return a
}