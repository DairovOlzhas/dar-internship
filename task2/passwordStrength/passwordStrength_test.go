package passwordStrength

import (
	"testing"
)

var testCases = []struct{
	name 	   string
	config     Config
	password   string
	userInputs []string
	expected   int
}{
	{
		name: "1",
		config: Config{
			RegExps: map[string]int{
				`[[:ascii:]]{8,}`:0,
				`[[:digit:]]{1,}`:0,
				`[[:upper:]]{1,}`:0,
				`[[:lower:]]{1,}`:0,
			},
		},
		password: "password",
		expected: 0,
	},
	{
		name: "2",
		config: Config{
			RegExps: map[string]int{
				`[[:ascii:]]{8,}`:0,
				`[[:digit:]]{1,}`:0,
				`[[:upper:]]{1,}`:0,
				`[[:lower:]]{1,}`:0,
			},
		},
		password: "Passw0rd",
		expected: 4,
	},
	{
		name: "3",
		config: Config{
			RegExps: map[string]int{
				`[[:ascii:]]{8,}`:1,
				`[[:digit:]]{1,}`:1,
				`[[:upper:]]{1,}`:1,
				`[[:lower:]]{1,}`:1,
			},
		},
		password: "Passw0rd",
		expected: 4,
	},
	{
		name: "4",
		config: Config{
			RegExps: map[string]int{
				`[[:ascii:]]{8,}`:1,
				`[[:digit:]]{1,}`:1,
				`[[:upper:]]{1,}`:1,
				`[[:lower:]]{1,}`:1,
			},
		},
		password: "Password",
		expected: 3,
	},
	{
		name: "5",
		config: Config{
			RegExps: map[string]int{
				`[[:ascii:]]{8,}`:1,
				`[[:digit:]]{1,}`:10,
				`[[:upper:]]{1,}`:1,
				`[[:lower:]]{1,}`:1,
			},
		},
		password: "Password",
		expected: 1,
	},
	{
		name: "6",
		config: Config{
			RegExps: map[string]int{
				`[[:ascii:]]{8,}`:1,
				`[[:digit:]]{1,}`:10,
				`[[:upper:]]{1,}`:1,
				`[[:lower:]]{1,}`:1,
			},
		},
		password: "_./%&#",
		expected: 0,
	},
	{
		name: "7",
		config: Config{
			RegExps: map[string]int{
				`[[:ascii:]]{8,}`:1,
				`[[:digit:]]{1,}`:1,
				`[[:upper:]]{1,}`:1,
				`[[:lower:]]{1,}`:1,
				`[[:punct:]]{1,}`:10,
			},
		},
		password: "_./%&#",
		expected: 2,
	},
	{
		name: "8",
		config: Config{
			RegExps: map[string]int{
				`[[:ascii:]]{8,}`:0,
				`[[:digit:]]{1,}`:1,
				`[[:upper:]]{1,}`:1,
				`[[:lower:]]{1,}`:1,
				`[[:punct:]]{1,}`:10,
			},
		},
		password: "_./%&#",
		expected: 0,
	},
	{
		name: "9",
		config: Config{
			MinEditDistFromInputs: 2,
		},
		password: "asdfghjkl",
		userInputs: []string{"asdfghj"},
		expected: 4,
	},
	{
		name: "10",
		config: Config{
			MinEditDistFromInputs: 2,
		},
		password: "asdfghjkl",
		userInputs: []string{"asdfghjk"},
		expected: 0,
	},
	{
		name: "11",
		config: Config{
			MinEditDistFromInputs: 80,
		},
		password: "asdfghjkl",
		userInputs: []string{"12345678"},
		expected: 0,
	},
	{
		name: "12",
		config: Config{
			RegExps: map[string]int{
				`[[:ascii:]]{1,}`:0,
			},
			MinEditDistFromInputs: 3,
		},
		password: "l",
		userInputs: []string{"2"},
		expected: 0,
	},
	{
		name: "13",
		config: Config{
			MinEditDistFromInputs: 1,
		},
		password: "123456",
		userInputs: []string{"123455"},
		expected: 3,
	},
	{
		name: "14",
		config: Config{
			SearchInDictionary: true,
			PathToDict: "dictionary.txt",
		},
		password: "olzhas",
		expected: 0,
	},
	{
		name: "15",
		config: Config{
			SearchInDictionary: true,
			PathToDict: "dictionary.txt",
		},
		password: "ODL<CSDFxx",
		expected: 4,
	},
	{
		name: "16",
		config: Config{
			Entropy: true,
		},
		password: "ODL<CSDFxx",
		expected: 3,
	},
	{
		name: "17",
		config: Config{
			Entropy: true,
		},
		password: "1",
		expected: 0,
	},
	{
		name: "18",
		config: Config{
			MinEditDistFromInputs: 0,
			RegExps: map[string]int{
			},
			SearchInDictionary: false,
			Entropy: false,
			PathToDict: "dictionary.txt",
		},
		password: "",
		userInputs: []string{},
		expected: 4,
	},
}

func TestPasswordStrength_Calc(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ps := NewPasswordStrength(tc.config)
			got, err := ps.Calc(tc.password, tc.userInputs)
			if err != nil{
				t.Fatal(err)
			}
			if got != tc.expected {
				t.Errorf("Got %d, but expected %d", got, tc.expected)
			}
		})
	}
}

func BenchmarkPasswordStrength_Calc(b *testing.B) {
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			ps := NewPasswordStrength(tc.config)
			if tc.config.SearchInDictionary {
				err := ps.loadDict()
				if err != nil {
					b.Fatal(err)
				}
			}
			b.ResetTimer()
			for i:=0; i < b.N; i++ {
				_, err := ps.Calc(tc.password, tc.userInputs)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}