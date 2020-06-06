package passwordStrength



type Config struct {
	MinEditDistFromInputs int
	RegExps               map[string]int	// [regexp]=points, points for a regexp, 0 points if required.
	SearchInDictionary    bool 				// If true, will search password in dictionary.
	Entropy				  bool				// If true, will calculate password entropy
	PathToDict			  string			// Location of txt file with list of passwords
}

type PasswordStrength struct {
	config 		Config
	dictionary	[]string	// stores common passwords
	dictLoaded	bool		// loading status of dictionary
}
