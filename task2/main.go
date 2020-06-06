package main

import (
	"fmt"
	"github.com/dairovolzhas/dar-internship/task2/passwordStrength"
)

func main() {


	config := passwordStrength.Config{
		RegExps: map[string]int{
			`[[:alpha:]]{8,}`:1,
		},
		SearchInDictionary: true,
	}
	ps := passwordStrength.NewPasswordStrength(config)
	for {
		var s string
		fmt.Print("Type password:\n")
		fmt.Scanf("%s", &s)
		fmt.Println(ps.Calc(s))
	}

}
