package config

import "os"
import "io"
import "log"
import "encoding/json"

type CoefontConfig struct {
	AccessKey    string  `json:"access_key"`
	ClientSecret string  `json:"client_secret"`
	FontUUID     string  `json:"font_uuid"`
	Speed        float64 `json:"speed"`
}

type VoicevoxConfig struct {
	Enabled               bool     `json:"enabled"`
	ShouldSkipNonJapanese bool     `json:"should_skip_non_japanese"`
	APIKeys               []string `json:"api_keys"`
	URL                   string   `json:"url"`
	Speaker               string   `json:"speaker"`
	Speed                 float64  `json:"speed"`
}

type ReadlineConfig struct {
	VimMode     bool   `json:"vim_mode"`
	HistoryFile string `json:"history_file"`
}

type CustomPrefixConfig struct {
	Prefix string   `json:"prefix"`
	Args   []string `json:"args"`
}

type Config struct {
	Coefont          CoefontConfig        `json:"coefont"`
	Voicevox         VoicevoxConfig       `json:"voicevox"`
	Readline         ReadlineConfig       `json:"readline"`
	OutputDir        string               `json:"output_dir"`
	TimeoutSec       int                  `json:"timeout_sec"`
	CustomPrefixList []CustomPrefixConfig `json:"custom_prefix_list"`
	CustomPrefixMap  map[string][]string
}

func ReadConfigFile(filepath string) (config Config, err error) {

	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("Failed to open the file [ %v ]: %v\n", filepath, err)
		return config, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Failed to read the file [ %v ]: %v\n", filepath, err)
		return config, err
	}
	defer file.Close()

	err = json.Unmarshal(content, &config)
	if err != nil {
		return config, err
	}

	config.CustomPrefixMap = map[string][]string{}
	for _, v := range config.CustomPrefixList {
		config.CustomPrefixMap[v.Prefix] = v.Args
	}

	return config, nil

}
