// Package morningpost provides a client type and functions for interacting with
// the HackerNews API. It also provides a Summarizer interface so that packages
// which define implementation types for other news sources can be imported.
package morningpost

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// HNClient provides a client for interacting with the HackerNews API. For details
// about the HackerNews API, please see https://github.com/HackerNews/API.
type HNClient struct {
	BaseURL    string
	HttpClient *http.Client
}

// NewHNClient returns a client that is ready to interact with the HackerNews
// API at https://hacker-news.firebaseio.com.
func NewHNClient() *HNClient {
	return &HNClient{
		BaseURL: "https://hacker-news.firebaseio.com",
		HttpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

/*
Summary returns the 10 newest HackerNews story items as a string of
line-separated story titles and URLs like:

	Story Title 1
	http://story-title-1.com

	Story Title 2
	https://story-title2.com

An error is returned if the client has a problem generating the list of newest
story IDs or generating the details for a particular story.
*/
func (h *HNClient) Summary() (string, error) {
	storyIDs, err := h.NewestStories()
	if err != nil {
		return "", err
	}
	maxNumStories := 10
	if len(storyIDs) < maxNumStories {
		maxNumStories = len(storyIDs)
	}
	summary := "Latest HackerNews Stories\n=========================\n\n"
	for i := 0; i < maxNumStories; i++ {
		story, err := h.Story(storyIDs[i])
		if err != nil {
			return "", err
		}
		summary += story.Title + "\n" + story.Url + "\n\n"
	}
	return summary, nil
}

// NewestStories queries the HackerNews API for the newest story items and
// returns a slice of ints representing the item IDs of these stories. An error
// is returned if there is a problem communicating with the API, if an invalid
// HTTP response code is received, or if the response cannot be parsed into an
// int slice.
func (h *HNClient) NewestStories() ([]int, error) {
	resp, err := h.HttpClient.Get(h.BaseURL + "/v0/newstories.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got unexpected response code %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	newest, err := ParseHNNewestStoriesResponse(data)
	if err != nil {
		return nil, err
	}
	return newest, nil
}

// Story queries the HackerNews API for the item with the given id and returns
// an HNStory struct representing the story. An error is returned if there is a
// problem communicating with the API, if an invalid HTTP reponse code is
// received, or if the response cannot be parsed into a HNStory struct.
func (h *HNClient) Story(id int) (HNStory, error) {
	resp, err := h.HttpClient.Get(fmt.Sprintf("%s/v0/item/%d.json", h.BaseURL, id))
	if err != nil {
		return HNStory{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return HNStory{}, fmt.Errorf("got unexpected response code %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return HNStory{}, err
	}
	story, err := ParseHNStoryResponse(data)
	if err != nil {
		return HNStory{}, err
	}
	return story, nil
}

// ParseHNNewestStoriesResponse accepts a slice of bytes representing a response
// to a query of the HackerNews API's newest stories endpoint and returns a
// slice of ints containing the item IDs of the newest stories. An error is
// returned if there is a problem parsing the response data into a slice of ints.
func ParseHNNewestStoriesResponse(data []byte) ([]int, error) {
	var hnResp []int
	err := json.Unmarshal(data, &hnResp)
	if err != nil {
		return nil, fmt.Errorf("invalid API response %s: %w", data, err)
	}
	return hnResp, nil
}

// HNStory represents a HackerNews API story item.
type HNStory struct {
	Title string
	Url   string
}

// ParseHNStoryResponse accepts a slice of bytes representing a response to a
// query of the HackerNews API's item endpoint and returns an HNStory struct.
// An error is returned if there is a problem parsing the response data into an
// HNStory struct.
func ParseHNStoryResponse(data []byte) (HNStory, error) {
	var hns HNStory
	err := json.Unmarshal(data, &hns)
	if err != nil {
		return HNStory{}, fmt.Errorf("invalid API response: %s: %w", data, err)
	}
	return hns, nil
}

// Summarizer is the interface that wraps the basic Summary method.
//
// Summary returns a news summary as a string that should be suitable for
// reading by human beings. An error is returned for any problems building
// the summary (e.g. problems communicating with an API, problems parsing API
// responses, etc.)
type Summarizer interface {
	Summary() (string, error)
}

// WriteSummaries accepts an io.Writer w and a variable number of Summarizers
// representing news sources, retrieves the summaries from the Summarizers and
// writes the summaries to w. An error is returned for any call to a Summarizer's
// Summary() method that returns an error. If a call to a Summarizer's Summary()
// method returns an error, it does not stop subsequent Summarizers in the
// summaries list from being processed.
func WriteSummaries(w io.Writer, summaries ...Summarizer) error {
	var errs []error
	for _, sum := range summaries {
		s, err := sum.Summary()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		fmt.Fprintln(w, s)
	}
	return errors.Join(errs...)
}

// Main prints the newest HackerNews stories to standard output and returns an
// int exit code. Any non-zero exit code is accompanied with an error message
// written to the stderr steam.
func Main() int {
	hnClient := NewHNClient()
	err := WriteSummaries(os.Stdout, hnClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
