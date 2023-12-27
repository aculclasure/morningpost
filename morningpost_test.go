package morningpost_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aculclasure/morningpost"
	"github.com/google/go-cmp/cmp"
)

func TestNewestStories_ReturnsExpectedHackerNewsStoryIDs(t *testing.T) {
	t.Parallel()
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantURI := "/v0/newstories.json"
		gotURI := r.RequestURI
		if wantURI != gotURI {
			t.Fatalf("want request URI %s, got %s", wantURI, gotURI)
		}
		fmt.Fprint(w, "[38776446, 38776437]")
	}))
	defer ts.Close()
	c := morningpost.NewHNClient()
	c.BaseURL = ts.URL
	c.HttpClient = ts.Client()
	want := []int{38776446, 38776437}
	got, err := c.NewestStories()
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestStory_ReturnsExpectedHNStory(t *testing.T) {
	t.Parallel()
	wantStoryID := 38777401
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantURI := fmt.Sprintf("/v0/item/%d.json", wantStoryID)
		gotURI := r.RequestURI
		if wantURI != gotURI {
			t.Fatalf("want request URI %s, got %s", wantURI, gotURI)
		}
		http.ServeFile(w, r, "testdata/hackernews_story_item_response.json")
	}))
	defer ts.Close()
	c := morningpost.NewHNClient()
	c.BaseURL = ts.URL
	c.HttpClient = ts.Client()
	want := morningpost.HNStory{
		Title: "Computer-Based System Safety Essential Reading List",
		Url:   "http://safeautonomy.blogspot.com/p/safe-autonomy.html",
	}
	got, err := c.Story(wantStoryID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestParseHNNewestStoriesResponse_CorrectlyParsesJSONResponse(t *testing.T) {
	t.Parallel()
	data := []byte(`[38776446, 38776437]`)
	want := []int{38776446, 38776437}
	got, err := morningpost.ParseHNNewestStoriesResponse(data)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestParseHNNewestStoriesResponse_ReturnsErrorGivenInvalidResponse(t *testing.T) {
	t.Parallel()
	data := []byte(`["not-an-int"]`)
	_, err := morningpost.ParseHNNewestStoriesResponse(data)
	if err == nil {
		t.Fatal("want error parsing invalid response, got nil")
	}
}

func TestParseHNStoryResponse_CorrectlyParsesJSONResponse(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/hackernews_story_item_response.json")
	if err != nil {
		t.Fatal(err)
	}
	want := morningpost.HNStory{
		Title: "Computer-Based System Safety Essential Reading List",
		Url:   "http://safeautonomy.blogspot.com/p/safe-autonomy.html",
	}
	got, err := morningpost.ParseHNStoryResponse(data)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestParseHNStoryResponse_ReturnsErrorGivenEmptyJSON(t *testing.T) {
	t.Parallel()
	_, err := morningpost.ParseHNStoryResponse([]byte(`[]`))
	if err == nil {
		t.Fatal("expected an error for empty data but got nil")
	}
}

type mockSummarizer struct {
	summary string
	err     error
}

func (m *mockSummarizer) Summary() (string, error) {
	return m.summary, m.err
}

func TestWriteSummaries_CorrectlyWritesSummariesToOutputGivenAllValidSummarizers(t *testing.T) {
	output := new(bytes.Buffer)
	s := []morningpost.Summarizer{
		&mockSummarizer{summary: "news1"},
		&mockSummarizer{summary: "news2"},
		&mockSummarizer{summary: "news3"},
	}
	err := morningpost.WriteSummaries(output, s...)
	if err != nil {
		t.Fatal(err)
	}
	want := "news1\nnews2\nnews3\n"
	got := output.String()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestWriteSummaries_CorrectlyWritesValidSummariesToOutputAndReturnsErrorForInvalidSummaries(t *testing.T) {
	output := new(bytes.Buffer)
	s := []morningpost.Summarizer{
		&mockSummarizer{summary: "news1"},
		&mockSummarizer{err: errors.New("oh no!")},
		&mockSummarizer{summary: "news3"},
	}
	err := morningpost.WriteSummaries(output, s...)
	if err == nil {
		t.Fatal("expected an error for invalid summarizer but got nil")
	}
	want := "news1\nnews3\n"
	got := output.String()
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}
