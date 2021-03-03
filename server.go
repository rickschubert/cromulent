package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// App represents the server's internal state.
// It holds configuration about providers and content
type App struct {
	ContentClients map[Provider]Client
	Config         ContentMix
}

type FetchedContentsMap = map[ContentConfig][]*ContentItem

type ContentsOfConfig struct {
	config ContentConfig
	items  []*ContentItem
}

type CountsPerConfig = map[ContentConfig]int

func sendMissingParameterError(missingParam string, w http.ResponseWriter) {
	w.WriteHeader(400)
	w.Write([]byte(fmt.Sprintf("Please provide the '%s' query parameter.", missingParam)))
}

func sendIncorrectParameterError(missingParam string, w http.ResponseWriter) {
	w.WriteHeader(400)
	w.Write([]byte(fmt.Sprintf("Please provide the '%s' query parameter as number.", missingParam)))
}

func sendInternalServerError(err error, w http.ResponseWriter) {
	w.WriteHeader(500)
	// TODO: In production, we probably wouldn't want to forward the error message
	w.Write([]byte(fmt.Sprintf("Something went wrong. Sorry! %s", err.Error())))
}

func getQueryParameter(param string, w http.ResponseWriter, req *http.Request) int {
	parameterGiven := req.URL.Query().Get(param)
	if parameterGiven == "" {
		sendMissingParameterError(param, w)
	}
	paramAsNr, err := strconv.Atoi(parameterGiven)
	if err != nil {
		sendIncorrectParameterError(param, w)
	}
	return paramAsNr
}

// Stretches and repeats a list given a count. I.e. the list [1, 2, 3] would be
// stretched to the following should the desired count be 8: [1, 2, 3, 1, 2, 3, 1, 2]
func stretchContentMixOverCount(config ContentMix, count int, offset int, w http.ResponseWriter) ContentMix {
	configLastIdx := len(config) - 1
	if configLastIdx == -1 {
		sendInternalServerError(errors.New("The app configuration is empty."), w)
	}
	var configIdx = 0
	var stretchedContentMix []ContentConfig
	for i := 0; i < offset+count; i++ {
		if i >= offset {
			stretchedContentMix = append(stretchedContentMix, config[configIdx])
		}
		configIdx++
		if configIdx > configLastIdx {
			configIdx = 0
		}
	}
	return stretchedContentMix
}

// Returns how many items should be fetched per content config
func getContentCountsPerConfig(order ContentMix) CountsPerConfig {
	countsPerConfig := make(CountsPerConfig)
	for _, item := range order {
		if _, containsKey := countsPerConfig[item]; containsKey {
			countsPerConfig[item]++
		} else {
			countsPerConfig[item] = 1
		}
	}
	return countsPerConfig
}

// Fetches the items for each config, using fallback strategy
func fetchItemsForConfig(app App, confItem ContentConfig, req *http.Request, amount int, wg *sync.WaitGroup, channel chan<- ContentsOfConfig) {
	defer wg.Done()
	contents, err := app.ContentClients[confItem.Type].GetContent(req.RemoteAddr, amount)
	if err != nil {
		contents, err = app.ContentClients[*confItem.Fallback].GetContent(req.RemoteAddr, amount)
		if err != nil {
			contents = nil
		}
	}
	channel <- ContentsOfConfig{
		config: confItem,
		items:  contents,
	}
}

func writeJsonResponse(writer http.ResponseWriter, returnList []ContentItem) {
	writer.Header().Add("content-type", "application/json")
	jsonData, err := json.Marshal(returnList)
	if err != nil {
		sendInternalServerError(err, writer)
	}
	writer.Write(jsonData)
}

// Arranges the fetched contents in the expected order
func generateListOfItemsToReturn(order ContentMix, contents FetchedContentsMap) []ContentItem {
	var returnList []ContentItem
	for _, config := range order {
		if contents[config] == nil {
			break
		}
		returnList = append(returnList, *contents[config][0])
		// Delete the item we just plucked off from the slice
		contents[config] = append(contents[config][:0], contents[config][1:]...)
	}
	return returnList
}

// Reads contents sent through channel and puts them on a map
func getMapOfFetchedContents(CountsPerConfig CountsPerConfig, channel <-chan ContentsOfConfig) FetchedContentsMap {
	contents := make(FetchedContentsMap)
	for i := 0; i < len(CountsPerConfig); i++ {
		respo := <-channel
		contents[respo.config] = respo.items
	}
	return contents
}

func (app App) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.String())

	count := getQueryParameter("count", writer, req)
	offset := getQueryParameter("offset", writer, req)

	order := stretchContentMixOverCount(app.Config, count, offset, writer)
	CountsPerConfig := getContentCountsPerConfig(order)

	waitgroup := sync.WaitGroup{}
	waitgroup.Add(len(CountsPerConfig))
	contents := make(chan ContentsOfConfig)

	for config, amount := range CountsPerConfig {
		go fetchItemsForConfig(app, config, req, amount, &waitgroup, contents)
	}
	fetchedContents := getMapOfFetchedContents(CountsPerConfig, contents)
	waitgroup.Wait()

	returnList := generateListOfItemsToReturn(order, fetchedContents)
	writeJsonResponse(writer, returnList)
}
