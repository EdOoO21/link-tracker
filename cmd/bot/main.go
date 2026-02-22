package main

import (
	"fmt"
	"os"

	app "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application"
	env "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/environment"
)

func main() {
	r := env.NewReader()
	a := app.NewApp()

	config, err := r.GetEnv()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := a.Run(config); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
