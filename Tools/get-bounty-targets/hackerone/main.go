package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	RATE_LIMIT_WAIT_TIME_SEC = 5
	RATE_LIMIT_MAX_RETRIES   = 10
	RATE_LIMIT_HTTP_STATUS   = 429
)

var (
	username = flag.String("u", "", "Your HackerOne username")
	apiKey   = flag.String("key", "", "Your HackerOne API key")
)

type ResponseData struct {
	Data []Program `json:"data"`
}

type Program struct {
	Attributes ProgramAttributes `json:"attributes"`
}

type ProgramAttributes struct {
	Handle string `json:"handle"`
}

func main() {
	flag.Parse()

	if *username == "" || *apiKey == "" {
		fmt.Println("Please provide your HackerOne username (-u) and API key (-key) as flags.")
		return
	}

	authorization := base64.StdEncoding.EncodeToString([]byte(*username + ":" + *apiKey))

	for i := 0; i < RATE_LIMIT_MAX_RETRIES; i++ {
		resp, err := sendHTTPRequest("GET", "https://api.hackerone.com/v1/hackers/programs", authorization)
		if err != nil {
			fmt.Printf("HTTP request failed: %v. Retrying...\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if resp.StatusCode != RATE_LIMIT_HTTP_STATUS {
			body, _ := ioutil.ReadAll(resp.Body)

			// Decode the JSON response body into the ResponseData struct
			var responseData ResponseData
			err = json.Unmarshal(body, &responseData)
			if err != nil {
				fmt.Printf("Failed to parse JSON response: %v\n", err)
				return
			}

			// Iterate over all programs in the response
			for _, program := range responseData.Data {
				handle := program.Attributes.Handle
				resp, err := sendHTTPRequest("GET", "https://api.hackerone.com/v1/hackers/programs/"+handle, authorization)
				if err != nil {
					fmt.Printf("HTTP request failed for handle '%v': %v. Retrying...\n", handle, err)
					time.Sleep(2 * time.Second)
					continue
				}

				if resp.StatusCode != RATE_LIMIT_HTTP_STATUS {
					body, _ := ioutil.ReadAll(resp.Body)

					// Extract the domains from the structured scopes
					var scopeData map[string]interface{}
					err = json.Unmarshal(body, &scopeData)
					if err != nil {
						fmt.Printf("Failed to parse JSON response: %v\n", err)
						return
					}

					scopes := scopeData["relationships"].(map[string]interface{})["structured_scopes"].(map[string]interface{})["data"].([]interface{})
					for _, scope := range scopes {
						scopeAttributes := scope.(map[string]interface{})["attributes"].(map[string]interface{})
						domain := scopeAttributes["asset_identifier"].(string)

						// Remove prefixes and suffixes from the domain name
						domain = removePrefixes(domain, []string{"https://*", "http://*", "https://", "http://", ".", "*."})
						domain = strings.TrimSuffix(domain, "/")
						domain = strings.TrimPrefix(domain, "*.")

						// Check if the domain matches the format [subdomain.]domain.[com]
						matched, _ := regexp.MatchString(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`, domain)
						if matched {
							// Check if the domain does not start with "com."
							if !strings.HasPrefix(domain, "com.") {
								fmt.Println("Domain:", domain)
							}
						}
					}
				} else {
					time.Sleep(RATE_LIMIT_WAIT_TIME_SEC * time.Second)
				}
			}

			break
		} else {
			time.Sleep(RATE_LIMIT_WAIT_TIME_SEC * time.Second)
		}
	}
}

func sendHTTPRequest(method string, url string, authorization string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Basic "+authorization)
	client := &http.Client{}
	return client.Do(req)
}

func removePrefixes(s string, prefixes []string) string {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			s = strings.TrimPrefix(s, prefix)
		}
	}
	return s
}
