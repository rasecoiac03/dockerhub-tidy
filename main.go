package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rasecoiac03/dockerhub-tidy/config"
)

type LoginResponse struct {
	Token string `json:"token"`
}

type RepositoriesResponse struct {
	Next         string                `json:"next"`
	Repositories []*RepositoryResponse `json:"results"`
}

type RepositoryResponse struct {
	Name        string    `json:"name"`
	LastUpdated time.Time `json:"last_updated"`
	IsPrivate   bool      `json:"is_private"`
	PullCount   int64     `json:"pull_count"`
}

func main() {
	auth := getAuth()
	if auth.Token == "" {
		panic("token cannot be empty")
	}

	requestURL := fmt.Sprintf("%s/v2/repositories/vivareal/?page=1&page_size=100",
		config.GetEnv("DOCKERHUB_BASE_URL"))
	repositories := getRepositories(auth.Token, requestURL)
	fmt.Println("repositories count:", len(repositories))

	// sort reverse by LastUpdated field
	sort.Slice(repositories, func(i, j int) bool {
		return repositories[j].LastUpdated.Before(repositories[i].LastUpdated)
	})

	// Format in tab-separated columns with a tab stop of 8.
	const padding = 1
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.Debug)

	fmt.Fprintln(w, "REPOSITORY\tPRIVATE\tPULL COUNT\tLAST UPDATED\t")
	for _, r := range repositories {
		fmt.Fprintf(w, "%s\t%v\t%d\t%s\t\n", r.Name, r.IsPrivate, r.PullCount, r.LastUpdated)
	}
	w.Flush()
}

func getRepositories(token, requestURL string) []*RepositoryResponse {
	// maximum page size avaiable at the moment
	request, _ := http.NewRequest(http.MethodGet, requestURL, nil)
	request.Header.Set("Authorization", fmt.Sprintf("JWT %s", token))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(fmt.Sprintf("error getting repositories, err: %v\n", err))
	}
	defer response.Body.Close()

	repositoriesResponse := &RepositoriesResponse{}
	rb, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(rb, &repositoriesResponse)
	if err != nil {
		panic(err)
	}

	repositories := repositoriesResponse.Repositories
	if repositoriesResponse.Next != "" {
		repositories = append(repositories, getRepositories(token, repositoriesResponse.Next)...)
	}

	return repositories
}

func getAuth() *LoginResponse {
	body := fmt.Sprintf("{\"username\": \"%s\", \"password\": \"%s\"}",
		config.GetEnv("DOCKER_USERNAME"), config.GetEnv("DOCKER_PASSWORD"))
	request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v2/users/login/",
		config.GetEnv("DOCKERHUB_BASE_URL")), strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(fmt.Sprintf("error connecting, err: %v\n", err))
	}
	defer response.Body.Close()

	loginResponse := &LoginResponse{}
	lr, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(lr, &loginResponse)

	return loginResponse
}
