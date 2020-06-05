package main

import "github.com/dairovolzhas/dar-internship/task2/passwordStrength"

var (
	regex = `.`
)
// password strength levels

func main() {
	config := passwordStrength.Config{
		RegExpReq: map[string]int{
			``:3,
		},
	}
	passwordStrength.LoadDict()
	passwordStrength.PasswordStrength("asdfasfdas", config)

}
