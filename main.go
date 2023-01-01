//usr/bin/env go run "$0" "$@"; exit

package main

import "fmt"
import "log"
import "regexp"
import "os"
import "strings"

import "github.com/chzyer/readline"

import "coefontuber/coefont"
import "coefontuber/voicevox"
import "coefontuber/config"
import "coefontuber/play"
import "coefontuber/util"

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

	for iter := 0; ; iter++ {

		var line, err = rl.Readline()
		if err != nil {
			break
		}

		if line == "" {
			fmt.Println()
			continue
		}

		//special commands
		var matches []string = prefixRegex.FindStringSubmatch(line)
		var additionalArgs []string = nil
		if len(matches) != 0 {

			var prefix = matches[1]

			if prefix == "help" {

				fmt.Printf(`[Special Commands]
!help
!list
!dict
!dict <word> <reading>
!dict del <word>` + "\n\n")
				continue

			} else if prefix == "list" {

				var prefixes = make([]string, len(config.CustomPrefixList))
				for i, v := range config.CustomPrefixList {
					prefixes[i] = v.Prefix
				}
				fmt.Printf("Registered prefixes: %v\n\n", prefixes)

				continue

			} else if prefix == "dict" {

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

		var resultChannel = make(chan string)

		if config.Voicevox.Enabled { //VOICEVOX
			if config.Voicevox.ShouldSkipNonJapanese && !util.IsJapanese(line) {
				fmt.Println()
				continue
			}
			var req = voicevox.Request{
				Speaker: config.Voicevox.Speaker,
				Speed:   config.Voicevox.Speed,
				Text:    line,
			}
			var common = voicevox.Common{
				APIKey:     config.Voicevox.APIKeys[iter%len(config.Voicevox.APIKeys)],
				URL:        config.Voicevox.URL,
				TimeoutSec: config.TimeoutSec,
				OutputDir:  config.OutputDir,
			}
			go voicevox.Text2Speech(req, common, resultChannel)
		} else { //CoeFont
			var req = coefont.Text2SpeechRequest{
				FontUUID: config.Coefont.FontUUID,
				Text:     line,
				Speed:    config.Coefont.Speed,
			}
			go coefont.Text2Speech(req, common, resultChannel)
		}
		batonIn = batonOut
		batonOut = make(chan struct{})
		go play.Play(resultChannel, batonIn, batonOut, additionalArgs)

		if isFirst {
			close(batonIn)
			isFirst = false
		}

		fmt.Println()

	}

}
