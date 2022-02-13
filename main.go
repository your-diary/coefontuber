//usr/bin/env go run "$0" "$@"; exit

package main

import "fmt"
import "log"
import "os"

import "coefontuber/coefont"
import "coefontuber/config"
import "coefontuber/play"

const configFile = "./config.json"
const apiURL = "https://api.coefont.cloud/v1/text2speech"

func main() {

	if len(os.Args) != 2 {
		fmt.Println("./main.go <string>")
		os.Exit(1)
	}
	var text = os.Args[1]

	var config, err = config.ReadConfigFile(configFile)
	if err != nil {
		log.Printf("Failed to read the config file [ %v ]: %v\n", configFile, err)
		return
	}

	var req = coefont.Request{
		FontUUID: config.Coefont.FontUUID,
		Text:     text,
		Speed:    config.Coefont.Speed,
	}

	var common = coefont.Common{
		AccessKey:    config.Coefont.AccessKey,
		ClientSecret: config.Coefont.ClientSecret,
		URL:          apiURL,
		TimeoutSec:   config.TimeoutSec,
		OutputDir:    config.OutputDir,
	}

	var filename = coefont.APICall(req, common)
	if filename != "" {
		play.Play(filename)
	}

}
