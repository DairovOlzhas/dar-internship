package passwordStrength

import (
	"bufio"
	"math"
	"os"
	"regexp"
	"sort"
	"unicode"
)

// password strength levels
const (
	VeryWeak   = 0
	Weak       = 1
	Reasonable = 2
	Strong     = 3
	VeryStrong = 4
)


// NewPasswordStrength returns passwordStrength class.
func NewPasswordStrength(config Config) *PasswordStrength {
	ps := &PasswordStrength{
		config:     config,
	}
	return ps
}


// Calc returns password strength.
// Takes password and user inputs list (for instance, first name, second name, email, etc.)
// 0 - Very Weak; 	might keep out family members
// 1 - Weak; 		should keep out most people, often good for desktop login passwords
// 2 - Reasonable; 	fairly secure passwords for network and company passwords
// 3 - Strong; 		can be good for guarding financial information
// 4 - Very Strong; often overkill.
func (ps *PasswordStrength) Calc(password string, userInputs []string) (int, error) {
	if ps.config.SearchInDictionary {
		if !ps.dictLoaded {
			err := ps.loadDict()
			if err != nil {
				return 0, err
			}
		}
		if ps.inDict(password) {
			return VeryWeak, nil
		}
	}

	var maxScore = 0
	var score = 0
	var maxStrength = 0
	var strength = 0
	// For loop calculates how far the password is from user inputs.
	for _, input := range userInputs {
		maxScore += len(input)
		distance := dist(password, input)

		if distance < ps.config.MinEditDistFromInputs {
			return VeryWeak, nil
		}
		score += len(password) - distance
	}

	for regex, points := range ps.config.RegExps {
		re := regexp.MustCompile(regex)
		maxScore+=points
		if re.MatchString(password) { // if password matches regexp
			score += points
		} else if points == 0 { // if password doesn't match regexp and it's must required regexp
			return VeryWeak, nil
		} // if password doesn't match regexp and regexp is not required then nothing happens
	}

	if maxScore > 0 {
		maxStrength += 4
		switch {
		case 100.0*score/maxScore  < 20:
			strength = VeryWeak
		case 100.0*score/maxScore < 40:
			strength = Weak
		case 100.0*score/maxScore < 75:
			strength = Reasonable
		case 100.0*score/maxScore < 90:
			strength = Strong
		default: //  >89
			strength = VeryStrong
		}
	}
	if ps.config.Entropy {
		maxStrength += 4
		strength += ps.entropy(password)
	}
	if maxStrength == 0 {
		return 4, nil
	}
	return 4*strength/maxStrength, nil
}

// entropy returns password strength based on entropy.
// Password entropy is a measurement of how unpredictable a password is.
// More information at https://www.pleacher.com/mp/mlessons/algebra/entropy2.html.
func (ps *PasswordStrength) entropy(password string) int {
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
	case entropy < 28: 	// 0-27
		return VeryWeak
	case entropy < 36: 	// 28-35
		return Weak
	case entropy < 60: 	// 36-59
		return Reasonable
	case entropy < 127:	// 60-126
		return Strong
	default:			// >126
		return VeryStrong
	}
}


// Loads dictionary from txt file.
func (ps *PasswordStrength) loadDict() error {
	file, err := os.Open(ps.config.PathToDict)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		ps.dictionary = append(ps.dictionary, scanner.Text())
	}
	sort.Strings(ps.dictionary)
	ps.dictLoaded = true
	return nil
}


// inDict returns password existence in dictionary.
func (ps *PasswordStrength) inDict(pass string) bool {
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