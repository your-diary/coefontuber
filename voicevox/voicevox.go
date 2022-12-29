//Unofficial API of VOICEVOX
//|https://voicevox.su-shiki.com/su-shikiapis/|

package voicevox

import "net/url"
import "fmt"
import "io"
import "log"
import "net/http"
import "os"
import "time"

type Common struct {
	APIKey     string
	URL        string
	TimeoutSec int
	OutputDir  string
}

type Request struct {
	Speaker string
	Speed   float64
	Text    string
}

func Text2Speech(req Request, common Common, resultChannel chan<- string) {

	defer close(resultChannel)

	var client = &http.Client{
		Timeout: time.Duration(common.TimeoutSec) * time.Second,
	}

	var query = url.Values{}
	query.Add("key", common.APIKey)
	query.Add("speaker", req.Speaker)
	query.Add("speed", fmt.Sprintf("%f", req.Speed))
	query.Add("text", req.Text)

	var response, err = client.Get(common.URL + "?" + query.Encode())
	if err != nil {
		log.Printf("Failed to send a GET request: %v\n", err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Println("Error response is returned.")
		b, _ := io.ReadAll(response.Body)
		log.Println(string(b))
		return
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Failed to read the response body: %v\n", err)
		return
	}

	var filename = fmt.Sprintf("%v/%v.wav", common.OutputDir, time.Now().UnixNano())
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to create the file [ %v ]: %v\n", filename, err)
		return
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		log.Printf("Failed to write to the file [ %v ]: %v\n", filename, err)
		return
	}

	resultChannel <- filename
	// log.Printf("Save: [ %v ]\n", filename)

}
