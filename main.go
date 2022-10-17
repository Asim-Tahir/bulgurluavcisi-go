package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-co-op/gocron"
)

const (
	redditURL          = "https://www.reddit.com"
	authRedditURL      = "https://oauth.reddit.com"
	redditAccUsername  = "bulgurluavcisi"
	redditAccPassword  = "password"
	redditClientId     = "client_id"
	redditClientSecret = "client_secret"
)

type TokenReq struct {
	AccessToken string `json:"access_token"`
}

type Posts struct {
	Kind string `json:"kind"`
	Data data   `json:"data"`
}

type data struct {
	After     string     `json:"after"`
	Dist      int        `json:"dist"`
	ModHash   string     `json:"modhash"`
	GeoFilter string     `json:"geo_filter"`
	Children  []children `json:"children"`
	Before    string     `json:"before"`
}

type children struct {
	Kind string       `json:"kind"`
	Data childrenData `json:"data"`
}

type childrenData struct {
	Subreddit  string  `json:"subreddit"`
	Author     string  `json:"author"`
	CreatedUTC float64 `json:"created_utc"`
	Id         string  `json:"id"`
}

type KarmaList struct {
	Kind string  `json:"kind"`
	Data []kData `json:"data"`
}

type kData struct {
	Subreddit string `json:"sr"`
}

func main() {
	token := ""
	startTime := time.Now()
	var lastMsgs []string

	s := gocron.NewScheduler(time.UTC)

	s.Every(86400).Seconds().Do(func() {
		token = getToken()
		fmt.Println("Retrieved token:", token)
	})

	s.Every(10).Seconds().Do(func() {
		if len([]rune(token)) < 1 {
			token = getToken()
			fmt.Println(token)
		}
		data := getData(token)

		for _, v := range data.Data.Children {
			post := v.Data
			if post.CreatedUTC > float64(startTime.Unix()) {
				if contains(lastMsgs, post.Id) {
				} else {
					a := checkAuthor(post.Author, token)
					if contains(a, "burdurland") {
						sendReply("amk bulgurlusu", post.Id, token)
						fmt.Println("Commented https://redd.it/" + post.Id)
					}
					if contains(a, "burdurban") {
						sendReply("amk bulgurlusu", post.Id, token)
						fmt.Println("Commented https://redd.it/" + post.Id)
					}
					lastMsgs = append(lastMsgs, post.Id)
				}
			}
		}
	})

	fmt.Println("Starting loop...")
	s.StartBlocking()
}

func getToken() string {
	bodyReq := []byte("grant_type=password&username=" + redditAccUsername + "&password=" + redditAccPassword)

	resp, _ := http.NewRequest("POST", redditURL+"/api/v1/access_token", bytes.NewBuffer(bodyReq))

	resp.Header.Add("Authorization", "Basic "+b64.URLEncoding.EncodeToString([]byte(redditClientId+":"+redditClientSecret)))
	client := &http.Client{}

	res, err := client.Do(resp)
	if err != nil {
		fmt.Println(err)
	}

	body, _ := io.ReadAll(res.Body)
	res.Body.Close()

	var tokenReq TokenReq

	json.Unmarshal(body, &tokenReq)
	return tokenReq.AccessToken
}

func getData(token string) Posts {
	resp, _ := http.NewRequest("GET", authRedditURL+"/r/KGBTR/new.json", nil)
	resp.Header.Add("Authorization", "bearer "+token)
	client := &http.Client{}

	res, err := client.Do(resp)
	if err != nil {
		fmt.Println(err)
	}

	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	var posts Posts

	json.Unmarshal(body, &posts)
	return posts
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func checkAuthor(author string, token string) []string {
	var srList []string
	resp, _ := http.NewRequest("GET", authRedditURL+"/u/"+author+"/top_karma_subreddits.json", nil)
	resp.Header.Add("Authorization", "bearer "+token)
	client := &http.Client{}

	res, err := client.Do(resp)
	if err != nil {
		fmt.Println(err)
	}

	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	var karmalist KarmaList

	json.Unmarshal(body, &karmalist)
	for _, v := range karmalist.Data {
		srList = append(srList, v.Subreddit)
	}
	return srList
}

func sendReply(message string, postId string, token string) {
	bodyReq := []byte("text=" + message + "&thing_id=t3_" + postId)
	resp, _ := http.NewRequest("POST", authRedditURL+"/api/comment", bytes.NewBuffer(bodyReq))

	resp.Header.Add("Authorization", "bearer "+token)
	client := &http.Client{}

	client.Do(resp)
}
