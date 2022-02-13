//usr/bin/env go run "$0" "$@"; exit

package main

import "fmt"
import "log"
import "os"

import "github.com/chzyer/readline"

import "coefontuber/coefont"
import "coefontuber/config"
import "coefontuber/play"

const (
	configFile = "./config.json"
	apiURL     = "https://api.coefont.cloud/v1/text2speech"
	prompt     = "\033[31m>>\033[0m "
)

func main() {

	var config, err = config.ReadConfigFile(configFile)
	if err != nil {
		log.Printf("Failed to read the config file [ %v ]: %v\n", configFile, err)
		os.Exit(1)
	}

	var common = coefont.Common{
		AccessKey:    config.Coefont.AccessKey,
		ClientSecret: config.Coefont.ClientSecret,
		URL:          apiURL,
		TimeoutSec:   config.TimeoutSec,
		OutputDir:    config.OutputDir,
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:      prompt,
		VimMode:     config.Readline.VimMode,
		HistoryFile: config.Readline.HistoryFile,
	})
	if err != nil {
		log.Printf("Failed to initialize GNU Readline: %v\n", err)
		os.Exit(1)
	}
	defer rl.Close()

	var batonIn chan struct{}
	var batonOut chan struct{} = make(chan struct{})
	var isFirst = true

	for {

		var line, err = rl.Readline()
		if err != nil {
			break
		}

		if line != "" {

			var req = coefont.Request{
				FontUUID: config.Coefont.FontUUID,
				Text:     line,
				Speed:    config.Coefont.Speed,
			}

			var resultChannel = make(chan string)
			batonIn = batonOut
			batonOut = make(chan struct{})

			go coefont.APICall(req, common, resultChannel)
			go play.Play(resultChannel, batonIn, batonOut)

			if isFirst {
				close(batonIn)
				isFirst = false
			}

		}

		fmt.Println()

	}

}
