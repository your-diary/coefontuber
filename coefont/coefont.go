package coefont

import "bytes"
import "crypto/hmac"
import "crypto/sha256"
import "encoding/hex"
import "encoding/json"
import "fmt"
import "io"
import "log"
import "net/http"
import "os"
import "time"

type Common struct {
	AccessKey    string
	ClientSecret string
	URL          string
	TimeoutSec   int
	OutputDir    string
}

/*-------------------------------------*/

/* header */

func createHeader(common Common, currentUnixSeconds int64, requestBody []byte) http.Header {

	//makes a signature
	var mac = hmac.New(sha256.New, []byte(common.ClientSecret))
	var message = fmt.Sprintf("%d%s", currentUnixSeconds, requestBody)
	mac.Write([]byte(message))
	var signature = hex.EncodeToString(mac.Sum(nil))

	return map[string][]string{
		"Content-Type":      []string{"application/json"},
		"Authorization":     []string{common.AccessKey},
		"X-Coefont-Date":    []string{fmt.Sprintf("%d", currentUnixSeconds)},
		"X-Coefont-Content": []string{signature},
	}

}

/*-------------------------------------*/

/* POST /text2speech */

type Text2SpeechRequest struct {
	FontUUID string  `json:"coefont"`
	Text     string  `json:"text"`
	Speed    float64 `json:"speed"`
}

func Text2Speech(req Text2SpeechRequest, common Common, resultChannel chan<- string) {

	defer close(resultChannel)

	var requestBody, err = json.Marshal(req)
	if err != nil {
		log.Printf("Failed to jsonalize the request body: %v\n", err)
		return
	}

	request, err := http.NewRequest(http.MethodPost, common.URL, bytes.NewReader(requestBody))
	if err != nil {
		log.Printf("Failed to create a first POST request: %v\n", err)
		return
	}

	var currentTime = time.Now().Unix()

	request.Header = createHeader(common, currentTime, requestBody)

	var client = &http.Client{
		Timeout: time.Duration(common.TimeoutSec) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	//The first request is sent to `api.coefont.cloud`.
	//The response is expected to 302 Found (i.e. redirect).
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Failed to send a first POST request: %v\n", err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusFound {
		log.Printf("The response isn't `302 Found`.\n")
		// b, _ := io.ReadAll(response.Body)
		// log.Println(string(b))
		return
	}

	//The second request is sent to `s3.amazonaws.com` from which we get the resultant .wav file.
	var redirectURL = response.Header.Get("Location")
	request, err = http.NewRequest(http.MethodGet, redirectURL /* body = */, nil)
	if err != nil {
		log.Printf("Failed to create a second GET request: %v\n", err)
		return
	}
	response, err = client.Do(request)
	if err != nil {
		log.Printf("Failed to send a second GET request: %v\n", err)
		return
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Failed to read the request body of a second GET request: %v\n", err)
		return
	}

	var filename = fmt.Sprintf("%v/%v_%v.wav", common.OutputDir, currentTime, req.Text)
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

/*-------------------------------------*/
