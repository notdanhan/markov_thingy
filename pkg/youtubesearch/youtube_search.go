package youtubesearch

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Youtube Search API wrapper
// Author Daniel Hannon
// Version 1

type YoutubePageInfo struct {
	TotalResults   int `json:"totalResults"`   // Number of results in total from the query
	ResultsPerPage int `json:"resultsPerPage"` // Amount of results returned by the search
}

type YoutubeVideoId struct {
	Kind    string `json:"kind"`    // The type of data provided
	VideoId string `json:"videoId"` // The ID of the video
}

type YoutubeVideoThumbs struct {
	Url    string `json:"url"`    // The URL of the thumbnail
	Width  int    `json:"width"`  // width of the thumbnail in px
	Height int    `json:"Height"` // Height of the thumbnail in px
}

type YoutubeVideoSnippet struct {
	PublishedAt          string                        `json:"publishedAt"`          // UTC timestamp of time of video being published
	ChannelId            string                        `json:"channelId"`            // The ID of the channel that uploaded the video
	Title                string                        `json:"title"`                // The Title of the video uploaded
	Description          string                        `json:"string"`               // The description of the video
	Thumbnails           map[string]YoutubeVideoThumbs `json:"thumbnails"`           // Information related to the thumbnails of the videos
	ChannelTitle         string                        `json:"channelTitle"`         // The name of the channel that uploaded the video
	LiveBroadcastContent string                        `json:"liveBroadcastContent"` // Live broadcast content (I don't actually know what this means)
	PublishTime          string                        `json:"publishTime"`          // The time the video was published
}

type YoutubeVideoData struct {
	Kind    string              `json:"kind"`    // type of content returned
	Etag    string              `json:"etag"`    // for replayability
	Id      YoutubeVideoId      `json:"id"`      // The information related directly to the video
	Snippet YoutubeVideoSnippet `json:"snippet"` // The data related to the video I.E thumbs and shit
}

type YoutubeApiResponse struct {
	Kind          string             `json:"kind"`          // The Kind of response from the API
	Etag          string             `json:"etag"`          // The Etag of the video (for caching :) )
	NextPageToken string             `json:"nextPageToken"` // Token for next page
	RegionCode    string             `json:"regionCode"`    // The Region of the search
	PageInfo      YoutubePageInfo    `json:"pageInfo"`      // Info related to the query
	Items         []YoutubeVideoData `json:"items"`         // The items returned
}

type YoutubeSearchCacheItem struct {
	Date  time.Time          `json:"Date"`
	Query string             `json:"Query"`
	Body  YoutubeApiResponse `json:"Body"`
}

type YoutubeApiHandler struct {
	ApiKey     string                            // The Api Key used
	Logger     *log.Logger                       // Logger to write diagnostics to
	listener   chan bool                         // A channel for killing the cache
	terminated chan bool                         // Kills the application
	cacheKeys  map[string]string                 // Cache keys
	cache      map[string]YoutubeSearchCacheItem // The actual search cache
	dataAge    float64                           // Maximum allowed age of the query
	lock       sync.Mutex                        // Maps are not thread safe so this locks them
}

func (yt *YoutubeApiHandler) initCache() {
	yt.Logger.Println("[Youtube API] Initalizing Cache for youtube API")
	yt.terminated = make(chan bool)
	yt.listener = make(chan bool)
	yt.cacheKeys = map[string]string{}
	yt.cache = make(map[string]YoutubeSearchCacheItem)
	// Purge data after an hour why not
	yt.dataAge = 3600.00
	go func() {
		defer close(yt.listener)
	main_loop:
		for {
			yt.lock.Lock()
			currT := time.Now()
			for k, v := range yt.cache {
				if v.Date.Sub(currT).Seconds() > yt.dataAge {
					yt.Logger.Println("[Youtube API] Expired cached item", v.Body.Etag, "Removed.")
					delete(yt.cacheKeys, v.Query)
					delete(yt.cache, k)
				}
			}
			yt.lock.Unlock()
			// Check the cache for stale results every 5 minutes
		time_loop:
			for {
				select {
				case <-yt.listener:
					yt.Logger.Println("[Youtube API] Cache loop terminated")
					break main_loop
				case <-time.After(5 * time.Minute):
					yt.Logger.Println("[Youtube API] Checking Cache")
					break time_loop
				}
			}
		}
		yt.terminated <- true
	}()
}

// Close terminates the YoutubeApiHandler
func (yt *YoutubeApiHandler) Close() {
	yt.Logger.Println("[Youtube API] Shutting down. Terminating Cache, Please allow some time.")
	defer close(yt.terminated)
	yt.listener <- true
	<-yt.terminated
	yt.Logger.Println("[Youtube API] Shut down.")
}

// MakeQuery searches to see for results, it also uses Local Caching to reduce bandwidth/api use :)
func (yt *YoutubeApiHandler) MakeQuery(query string) YoutubeApiResponse {
	yt.Logger.Println("[Youtube API] Searching Query \"", query, "\"")
	// need to lock for map accesses
	yt.lock.Lock()
	defer yt.lock.Unlock()
	if v, ok := yt.cacheKeys[query]; ok {
		if data, ok := yt.cache[v]; ok {
			yt.Logger.Println("[Youtube API] Result Was Cached, returning.")
			return data.Body
		}
	}
	yt.Logger.Println("[Youtube API] Nothing found in cache, creating URL")
	params := url.Values{}
	params.Add("type", "video")
	params.Add("maxResults", "10")
	params.Add("part", "snippet")
	params.Add("q", url.QueryEscape(query))
	yt.Logger.Println("[Youtube API] Query String Generated: https://www.googleapis.com/youtube/v3/search/?" + params.Encode())
	params.Add("key", yt.ApiKey)
	queryUrl := "https://www.googleapis.com/youtube/v3/search/?" + params.Encode()
	resp, err := http.Get(queryUrl)
	if err != nil {
		yt.Logger.Println("[Youtube API] Non-Fatal Error, failed to make HTTP request Reason:", err.Error())
		return YoutubeApiResponse{}
	}
	defer resp.Body.Close()
	var ytResponse YoutubeApiResponse
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		yt.Logger.Println("[Youtube API] Error occurred while parsing response", err.Error())
		return YoutubeApiResponse{}
	}
	err = json.Unmarshal(data, &ytResponse)
	if err != nil {
		yt.Logger.Println("[Youtube API] Failed to parse JSON, reason", err.Error(), "\nResponse Data:", string(data))
		return YoutubeApiResponse{}
	}
	yt.Logger.Println("[Youtube API] Response Successfully recieved and parsed, passing result to cache and returning")
	newCacheItem := YoutubeSearchCacheItem{time.Now(), query, ytResponse}
	yt.cache[ytResponse.Etag] = newCacheItem
	yt.cacheKeys[query] = ytResponse.Etag
	return ytResponse
}

func (yt *YoutubeApiHandler) GetRandomVid(query string) string {
	result := yt.MakeQuery(query)
	if result.Items == nil {
		yt.Logger.Println("[Youtube API] failed to get result, returning the funny.")
		return "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	}
	myVid := result.Items[rand.Intn(len(result.Items))]
	return "https://youtube.com/watch?v=" + myVid.Id.VideoId
}

// New creates a new YoutubeApiHandler and sets up the caching facilities
func New(apiKey string, logger *log.Logger) *YoutubeApiHandler {
	if logger == nil {
		logger = log.Default()
	}
	logger.Println("[Youtube API] Created, initalizing")
	myYtApiHandler := YoutubeApiHandler{ApiKey: apiKey, Logger: logger}
	myYtApiHandler.initCache()
	return &myYtApiHandler
}
