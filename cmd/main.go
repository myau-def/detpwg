package main

import "gitea.host/myau-def/detpwg"

func main() {
	// Основные модели
	login := detpwg.Login{
		Service: "service",
		Login:   "login",
		Info:    "account on service",
		Counter: 1,
	}

	alphabetCfg := detpwg.AlphabetConfig{
		Exclude:          "",
		IncludeLowercase: true,
		IncludeUppercase: true,
		IncludeDigits:    true,
		IncludeSpecial:   true,
	}
	alphabet, err := alphabetCfg.BuildAlphabet()
	if err != nil {
		panic(err)
	}

	// И

	key, err := detpwg.GenerateMasterKey("master", "password")
	if err != nil {
		panic(err)
	}

	password, err := detpwg.GeneratePassword(&key, &login, alphabet, 16)
	if err != nil {
		panic(err)
	}

	println(password)
}
