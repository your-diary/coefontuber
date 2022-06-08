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

func createHeader(common Common, requestBody []byte) http.Header {

	var currentUnixSeconds = time.Now().Unix()

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

	var requestHeader = createHeader(common, requestBody)
	request.Header = requestHeader

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
		if response.StatusCode == http.StatusBadRequest {
			log.Println("Failed. The input may include a forbidden word.")
		} else {
			log.Println("Failed. The response isn't `302 Found`.")
			// b, _ := io.ReadAll(response.Body)
			// log.Println(string(b))
		}
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
		log.Printf("Failed to read the response body of a second GET request: %v\n", err)
		return
	}

	var filename = fmt.Sprintf("%v/%v_%v.wav", common.OutputDir, requestHeader.Get("X-Coefont-Date"), req.Text)
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

/* GET /dict */

type getDictResponse struct {
	Text string `json:"text"`
	//Category string `json:"category"`
	Yomi string `json:"yomi"`
}

const dictURL = "https://api.coefont.cloud/v1/dict"

func GetDict(common Common) {

	request, err := http.NewRequest(http.MethodGet, dictURL, nil)
	if err != nil {
		log.Printf("Failed to create a GET request: %v\n", err)
		return
	}

	request.Header = createHeader(common, nil)

	var client = &http.Client{
		Timeout: time.Duration(common.TimeoutSec) * time.Second,
	}

	response, err := client.Do(request)
	if err != nil {
		log.Printf("Failed to send a GET request: %v\n", err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Printf("The response isn't `200 OK`.\n")
		return
	}
	content, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Failed to read the response body of a GET request: %v\n", err)
		return
	}

	var dictList []getDictResponse
	json.Unmarshal(content, &dictList)

	fmt.Println(dictList)

}

/*-------------------------------------*/

/* POST /dict */

type PostDictRequest struct {
	Text     string `json:"text"`
	Category string `json:"category"`
	Yomi     string `json:"yomi"`
}

func PostDict(req PostDictRequest, common Common) {

	var requestBody, err = json.Marshal([]PostDictRequest{req})
	if err != nil {
		log.Printf("Failed to jsonalize the request body: %v\n", err)
		return
	}

	request, err := http.NewRequest(http.MethodPost, dictURL, bytes.NewReader(requestBody))
	if err != nil {
		log.Printf("Failed to create a POST request: %v\n", err)
		return
	}

	request.Header = createHeader(common, requestBody)

	var client = &http.Client{
		Timeout: time.Duration(common.TimeoutSec) * time.Second,
	}

	response, err := client.Do(request)
	if err != nil {
		log.Printf("Failed to send a POST request: %v\n", err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Printf("The response isn't `200 OK`.\n")
		// log.Printf("%v\n", response.StatusCode)
		// b, _ := io.ReadAll(response.Body)
		// log.Println(string(b))
		return
	}

	fmt.Println("POST `/dict` succeeded.")

}

/*-------------------------------------*/
