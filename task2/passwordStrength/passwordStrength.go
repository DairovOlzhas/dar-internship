package passwordStrength

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	SameWithUserInputError = errors.New("Password too same with user inputs!")
	ExistInDictError	   = errors.New("Password exist in dictionary!")
)

// errors
const (
)

// password strength level
const (
	VERYWEAK   = 0
	WEAK       = 1
	MEDIUM     = 2
	STRONG	   = 3
	VERYSTRONG = 4
)

type Config struct {
	UserInputs			 	[]string
	minDistFromInputs		int
	RegexReq             	map[string]int
	NotExistInDictionary 	bool
}


func PasswordStrength(password string, config Config) (int, error) {
	maxScore := 0
	score := 0

	for i, regex := range config.RegexReq {
		maxScore+=regex
		re := regexp.MustCompile(i)
		if re.MatchString(password) {
			score+=regex
		}
	}

	for _, input := range config.UserInputs {
		maxScore += len(input)
		d := dist(password, input)
		if d < config.minDistFromInputs {
			return 0, SameWithUserInputError
		}
		score += len(password) - d
	}

	if indict(password) {
		return 0, ExistInDictError
	}

	strength := 100*score/maxScore
	switch  {
	case strength > 90:
		return VERYSTRONG, nil
	case strength > 75:
		return STRONG, nil
	case strength > 50:
		return MEDIUM, nil
	case strength > 25:
		return WEAK, nil
	default:
		return VERYWEAK, nil
	}
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