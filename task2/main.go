package main

import (
	"fmt"
	"github.com/dairovolzhas/dar-internship/task2/passwordStrength"
)



var (
	regex = `.`
)
// password strength levels

func main() {
	config := passwordStrength.Config{
		RegexReq: map[string]int{
			``:3,
		},
	}
	passwordStrength.PasswordStrength("asdfasfdas", )
}
