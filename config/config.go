package config

import "os"
import "io"
import "log"
import "encoding/json"

type ReadlineConfig struct {
	VimMode     bool   `json:"vim_mode"`
	HistoryFile string `json:"history_file"`
}

type CoefontConfig struct {
	AccessKey    string  `json:"access_key"`
	ClientSecret string  `json:"client_secret"`
	FontUUID     string  `json:"font_uuid"`
	Speed        float64 `json:"speed"`
}

type Config struct {
	Readline   ReadlineConfig `json:"readline"`
	Coefont    CoefontConfig  `json:"coefont"`
	OutputDir  string         `json:"output_dir"`
	TimeoutSec int            `json:"timeout_sec"`
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
	return config, err

}
