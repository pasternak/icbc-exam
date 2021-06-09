package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	loginURL        = "https://onlinebusiness.icbc.com/deas-api/v1/webLogin/webLogin"
	appointmentsURL = "https://onlinebusiness.icbc.com/deas-api/v1/web/getAvailableAppointments"
)

var (
	// ICBC params
	lastName      string
	licenseNumber string
	keyword       string
	locationID    int
	examType      string
	startDate     string
	endDate       string

	// pushover params
	token string
	user  string
)

type Login struct {
	LastName      string `json:"drvrLastName"`
	LicenceNumber string `json:"licenceNumber"`
	Keyword       string `json:"keyword,omitempty"`
}

type Exam struct {
	LocationID        int    `json:"aPosID"`
	Type              string `json:"examType"`
	Date              string `json:"examDate"`
	IgnoreReserveTime bool   `json:"ignoreReserveTime"`
	DaysOfWeek        string `json:"prfDaysOfWeek"`
	Time              string `json:"prfPartsOfDay"`
	LastName          string `json:"lastName"`
	LicenseNumber     string `json:"licenseNumber"`
}

type Header struct {
	Name  string
	Value string
}

type ICBCExam interface {
	payload(...Header) (*http.Request, error)
	query(req *http.Request) (map[string]interface{}, error)
}

func (l *Login) payload(h ...Header) (*http.Request, error) {
	method := "PUT"

	headers := map[string]string{
		"Sec-Ch-Ua":        "\" Not;A Brand\";v=\"99\", \"Google Chrome\";v=\"91\", \"Chromium\";v=\"91\"",
		"Pragma":           "no-cache",
		"Sec-Ch-Ua-Mobile": "?0",
		"User-Agent":       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36",
		"Content-Type":     "application/json",
		"Accept":           "application/json, text/plain, */*",
		"Cache-Control":    "no-cache, no-store",
		"Referer":          "https://onlinebusiness.icbc.com/webdeas-ui/login;type=driver",
		"Expires":          "0",
	}

	for _, opt := range h {
		headers[opt.Name] = opt.Value
	}

	return preparePayload(l, method, loginURL, headers)
}

func (l *Login) query(req *http.Request) (map[string]interface{}, http.Header, error) {
	var decodedBody map[string]interface{}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&decodedBody)

	return decodedBody, resp.Header, err
}

func (e *Exam) query(req *http.Request) (map[string]interface{}, http.Header, error) {
	var decodedBody []map[string]interface{}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&decodedBody)
	if err != nil {
		return decodedBody[0], resp.Header, err
	}
	if len(decodedBody) == 0 {
		var empty map[string]interface{}
		return empty, resp.Header, fmt.Errorf("no appointments found matching the search criteria.")
	}
	return decodedBody[0], resp.Header, nil
}

func (e *Exam) payload(h ...Header) (*http.Request, error) {
	method := "POST"

	headers := map[string]string{
		"Sec-Ch-Ua":        "\" Not;A Brand\";v=\"99\", \"Google Chrome\";v=\"91\", \"Chromium\";v=\"91\"",
		"Accept":           "application/json, text/plain, */*",
		"Referer":          "https://onlinebusiness.icbc.com/webdeas-ui/booking",
		"Sec-Ch-Ua-Mobile": "?0",
		"User-Agent":       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36",
		"Content-Type":     "application/json",
	}

	for _, opt := range h {
		headers[opt.Name] = opt.Value
	}

	return preparePayload(e, method, appointmentsURL, headers)

}

func preparePayload(in interface{}, method, url string, headers map[string]string) (*http.Request, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(data)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for key, val := range headers {
		req.Header.Set(key, val)
	}

	return req, nil
}

func pushover(token, user, message string) {
	resp, err := http.PostForm("https://api.pushover.net/1/messages.json", url.Values{
		"token":   {token},
		"user":    {user},
		"message": {message},
	})
	if err != nil {
		log.Println(err)
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
		}
		bodyString := string(bodyBytes)
		log.Println(bodyString)
	}
}

type array []string

func (a *array) String() string {
	return "locationIDs"
}

func (a *array) Set(value string) error {
	*a = append(*a, value)
	return nil
}

var locationIDs array

func main() {
	flag.StringVar(&lastName, "last-name", "", "Last Name")
	flag.StringVar(&licenseNumber, "license-number", "", "Licence number (yellow paper)")
	flag.StringVar(&keyword, "keyword", "", "The keyword used to authenticate")
	flag.Var(&locationIDs, "location-id", "The location ID, by default: Burnaby LoUghed HWY")
	flag.StringVar(&examType, "exam-type", "5-R-1", "The type of the exam, by default 5-R-1")
	flag.StringVar(&startDate, "start-date", time.Now().Format("2006-01-02"), "The type of the exam, by default 5-R-1")
	flag.StringVar(&endDate, "end-date", "", "The type of the exam, by default 5-R-1")

	flag.StringVar(&token, "pushover-token", "", "PushOver token")
	flag.StringVar(&user, "pushover-user", "", "PushOver user")

	flag.Parse()

	// Perform the login to the ICBC portal
	login := Login{
		LastName:      lastName,
		LicenceNumber: licenseNumber,
		Keyword:       keyword,
	}

	req, err := login.payload()
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Println("logging into ICBC portal.")
	//headers, err := makeRequest(req, &loginData)
	_, headers, err := login.query(req)
	if err != nil {
		log.Fatalf("failed logging into the portal: %s", err.Error())
	}

	// query the first free spot for a given location
	exam := Exam{
		Type:              examType,
		Date:              startDate,
		IgnoreReserveTime: false,
		DaysOfWeek:        "[0,1,2,3,4,5,6]",
		Time:              "[0,1]",
		LastName:          lastName,
		LicenseNumber:     licenseNumber,
	}

	for _, locationID := range locationIDs {
		if exam.LocationID, err = strconv.Atoi(locationID); err != nil {
			panic(err)
		}
		log.Printf("querying free appointmens for location: %d\n", exam.LocationID)
		req, err = exam.payload(Header{"Authorization", headers["Authorization"][0]})

		body, _, err := exam.query(req)
		if err != nil {
			log.Println(err)
			continue
		}
		appointments := body["appointmentDt"].(map[string]interface{})
		if endDate != "" {
			ed, err := time.Parse("2006-01-02", endDate)
			if err != nil {
				panic(err)
			}
			red, err := time.Parse("2006-01-02", appointments["date"].(string))
			if err != nil {
				panic(err)
			}
			if red.After(ed) {
				log.Println("no appointments found matching the search criteria.")
				continue
			}
		}
		message := fmt.Sprintf("Found appointment:\n\tlocation: %d, date: %s on %s, time: %s", exam.LocationID, appointments["date"].(string), appointments["dayOfWeek"].(string), body["startTm"].(string))
		log.Println(message)

		if token != "" && user != "" {
			pushover(token, user, message)
		}
	}
}
