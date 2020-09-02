package github

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

var httpClient http.Client
var githubApiUrl = os.Getenv("GITHUB_API_URL")
var githubRepo = os.Getenv("GITHUB_REPOSITORY")
var githubToken = os.Getenv("GITHUB_TOKEN")

var ua = "CreatedBy/GongT repo/" + githubRepo + " workflow/" + os.Getenv("GITHUB_WORKFLOW") + " run/" + os.Getenv("GITHUB_RUN_NUMBER")

var requestPerPage = 100

func init() {
	if len(githubApiUrl) == 0 {
		githubApiUrl = "https://api.github.com"
	}
	if len(githubToken) == 0 {
		fmt.Println("missing env: GITHUB_TOKEN")
		os.Exit(1)
	}

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient = http.Client{Transport: customTransport, Timeout: time.Minute}
}

func GetCurrentRepo() string {
	return githubRepo
}

func GithubRequest(request *http.Request) (*http.Response, error) {
	request.Header.Set("User-Agent", ua)
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	request.Header.Set("Authorization", fmt.Sprintf("token %s", githubToken))
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func DoRequest(method, endpoint string, query map[string]string) ([]byte, error) {
	request, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return nil, err
	}

	if query != nil {
		if method == "POST" {
			data, _ := json.Marshal(query)
			request.Body = ioutil.NopCloser(bytes.NewReader(data))
		} else {
			q := make(url.Values)
			for k, v := range query {
				q.Set(k, v)
			}
			request.URL.RawQuery = q.Encode()
		}
	}

	response, err := GithubRequest(request)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return body, fmt.Errorf("failed %v %v: status code: %d", method, endpoint, response.StatusCode)
	}

	return body, nil
}

func ApiUrl(format string, a ...interface{}) string {
	return githubApiUrl + "/repos/" + githubRepo + "/" + fmt.Sprintf(format, a...)
}
