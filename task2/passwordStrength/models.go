package passwordStrength



type Config struct {
	UserInputs            []string			// User inputs such as first name, second name, email, etc.
	MinEditDistFromInputs int
	RegExps               map[string]int	// [regexp]=points, points for a regexp, 0 points if required.
	SearchInDictionary    bool 				// If true, will search password in dictionary.
}

type PasswordStrength struct {
	config 		Config
	dictionary	[]string	// stores common passwords
	pathToDict	string		// location to txt file with list of passwords
	dictLoaded	bool		// loading status of dictionary
}
