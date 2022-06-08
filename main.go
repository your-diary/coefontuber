//usr/bin/env go run "$0" "$@"; exit

package main

import "fmt"
import "log"
import "regexp"
import "os"
import "strings"

import "github.com/chzyer/readline"

import "coefontuber/coefont"
import "coefontuber/config"
import "coefontuber/play"

const (
	configFile   = "./config.json"
	apiURL       = "https://api.coefont.cloud/v1/text2speech"
	dictCategory = "category"
	prompt       = "\033[31m>>\033[0m "
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
	var prefixRegex = regexp.MustCompile(`^!([^ ]+)( (.*))?$`)

	for {

		var line, err = rl.Readline()
		if err != nil {
			break
		}

		if line == "" {
			fmt.Println()
			continue
		} else if line == "!list" {
			var prefixes = make([]string, len(config.CustomPrefixList))
			for i, v := range config.CustomPrefixList {
				prefixes[i] = v.Prefix
			}
			fmt.Printf("Registered prefixes: %v\n\n", prefixes)
			continue
		}

		var matches []string = prefixRegex.FindStringSubmatch(line)
		var additionalArgs []string = nil
		if len(matches) != 0 {

			var prefix = matches[1]

			if prefix == "dict" {

				if matches[2] == "" {
					coefont.GetDict(common)
				} else {
					var tokens = strings.Split(matches[3], " ")
					if len(tokens) != 2 {
						fmt.Printf("Usage\n  !dict\n  !dict <word> <reading>\n  !dict del <word>\n\n")
						continue
					}
					if tokens[0] == "del" {
						coefont.DeleteDict(
							coefont.DeleteDictRequest{
								Text:     tokens[1],
								Category: dictCategory,
							},
							common,
						)
					} else {
						coefont.PostDict(
							coefont.PostDictRequest{
								Text:     tokens[0],
								Category: dictCategory,
								Yomi:     tokens[1],
							},
							common,
						)
					}
				}

				fmt.Println()
				continue

			} else {

				line = matches[3]
				if line == "" {
					fmt.Printf("No argument to `!` command.\n\n")
					continue
				}

				var args, ok = config.CustomPrefixMap[prefix]
				if !ok {
					fmt.Printf("Unknown prefix: %v\n\n", prefix)
					continue
				}
				additionalArgs = args

			}

		}

		var req = coefont.Text2SpeechRequest{
			FontUUID: config.Coefont.FontUUID,
			Text:     line,
			Speed:    config.Coefont.Speed,
		}

		var resultChannel = make(chan string)
		batonIn = batonOut
		batonOut = make(chan struct{})

		go coefont.Text2Speech(req, common, resultChannel)
		go play.Play(resultChannel, batonIn, batonOut, additionalArgs)

		if isFirst {
			close(batonIn)
			isFirst = false
		}

		fmt.Println()

	}

}
