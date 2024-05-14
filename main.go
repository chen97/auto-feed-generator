package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/tidwall/gjson"
)

type ContentInfo struct {
	Title string `json:"title"`
	Date  string `json:"date"` // Store date as a string for simplicity, or as time.Time for more complex manipulation
	Link  string `json:"link"`
}

type Data struct {
	Contents []ContentInfo `json:"contents"`
}

func fetchLeetCode() []ContentInfo {
	url := "https://leetcode.cn/graphql/noj-go"
	payload := map[string]interface{}{
		"query": "query recentAcSubmissions($userSlug: String!) {recentACSubmissions(userSlug: $userSlug) {submissionId submitTime question { title translatedTitle titleSlug questionFrontendId } } }",
		"variables": map[string]string{
			"userSlug": "chen-hao-v3",
		},
		"operationName": "recentAcSubmissions",
	}
	bodyData, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyData))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var contents []ContentInfo
	items := gjson.Get(string(body), "data.recentACSubmissions").Array()
	for _, item := range items {
		submitTime := item.Get("submitTime").Int()
		title := item.Get("question.title").String()
		date := time.Unix(submitTime, 0).Format("2006-01-02") // Converting Unix timestamp to date string
		link := "https://leetcode.cn/u/chen-hao-v3/"
		contents = append(contents, ContentInfo{Title: title, Date: date, Link: link})
	}
	return contents
}

func fetchMedium() []ContentInfo {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL("https://medium.com/feed/@dochenhao")
	if err != nil {
		panic(err)
	}

	var contents []ContentInfo
	for _, item := range feed.Items {
		date := item.PublishedParsed.Format("2006-01-02") // Ensuring date format consistency
		title := item.Title
		link := item.Link
		contents = append(contents, ContentInfo{Title: title, Date: date, Link: link})
	}
	return contents
}

func main() {
	leetCodeContents := fetchLeetCode()
	mediumContents := fetchMedium()

	allContents := append(leetCodeContents, mediumContents...) // Combining both sources into one slice

	// Optional: Sort by date if required using sort.Slice
	sort.Slice(allContents, func(i, j int) bool {
		return allContents[i].Date < allContents[j].Date
	})

	// To process or store the combined data
	fileData, err := json.MarshalIndent(allContents, "", "  ")
	if err != nil {
		panic(err)
	}

	// Creating a new structure to nest fileData under "feed"
	var rawData interface{}
	if err := json.Unmarshal(fileData, &rawData); err != nil {
		log.Fatalf("Failed to unmarshal data: %v", err)
	}
	wrappedData := map[string]interface{}{
		"feed": rawData,
	}

	// Marshal the new structure to JSON
	finalJSON, err := json.MarshalIndent(wrappedData, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal final JSON: %v", err)
	}

	// Log and write the final JSON to a file
	log.Println("data.json", string(finalJSON))
	if err := os.WriteFile("data.json", finalJSON, 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
}
