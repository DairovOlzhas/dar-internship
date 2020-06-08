package main

import (
	"fmt"
	"github.com/dairovolzhas/dar-internship/task2/passwordStrength"
	"log"
)

func main() {
	config := passwordStrength.Config{
		MinEditDistFromInputs: 0,
		RegExps: map[string]int{
			`[[:ascii:]]{8,}`:0,
			`[[:digit:]]{1,}`:0,
			`[[:upper:]]{1,}`:0,
			`[[:lower:]]{1,}`:0,
			`[[:punct:]]{1,}`:10,
		},
		SearchInDictionary: true,
		Entropy: true,
 		PathToDict: "passwordStrength/dictionary.txt",
	}
	strength := map[int]string{
		0: "Very Weak",
		1: "Weak",
		2: "Reasonable",
		3: "Strong",
		4: "Very Strong",
	}
	ps := passwordStrength.NewPasswordStrength(config)
	for {
		var s string
		fmt.Print("Type password:\n")
		fmt.Scanf("%s", &s)
		strgth, err := ps.Calc(s,[]string{"example@example.com", "username", "surname", "01081970", "87771234567"})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(strength[strgth])
	}
}
