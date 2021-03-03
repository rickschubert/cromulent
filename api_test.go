package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	SimpleContentRequest = httptest.NewRequest("GET", "/?offset=0&count=5", nil)
	OffsetContentRequest = httptest.NewRequest("GET", "/?offset=5&count=5", nil)
)

func runRequest(t *testing.T, srv http.Handler, r *http.Request) (content []*ContentItem) {
	response := httptest.NewRecorder()
	srv.ServeHTTP(response, r)

	if response.Code != 200 {
		t.Fatalf("Response code is %d, want 200", response.Code)
		return
	}

	decoder := json.NewDecoder(response.Body)
	err := decoder.Decode(&content)
	if err != nil {
		t.Fatalf("couldn't decode Response json: %v", err)
	}

	return content
}

func TestResponseCount(t *testing.T) {
	content := runRequest(t, app, SimpleContentRequest)

	if len(content) != 5 {
		t.Fatalf("Got %d items back, want 5", len(content))
	}

}

func TestResponseOrder(t *testing.T) {
	content := runRequest(t, app, SimpleContentRequest)

	if len(content) != 5 {
		t.Fatalf("Got %d items back, want 5", len(content))
	}

	for i, item := range content {
		if Provider(item.Source) != DefaultConfig[i%len(DefaultConfig)].Type {
			t.Errorf(
				"Position %d: Got Provider %v instead of Provider %v",
				i, item.Source, DefaultConfig[i].Type,
			)
		}
	}
}

func TestOffsetResponseOrder(t *testing.T) {
	content := runRequest(t, app, OffsetContentRequest)

	if len(content) != 5 {
		t.Fatalf("Got %d items back, want 5", len(content))
	}

	for j, item := range content {
		i := j + 5
		if Provider(item.Source) != DefaultConfig[i%len(DefaultConfig)].Type {
			t.Errorf(
				"Position %d: Got Provider %v instead of Provider %v",
				i, item.Source, DefaultConfig[i].Type,
			)
		}
	}
}

type FailingSampleProvider struct {
	Source Provider
}

func (cp FailingSampleProvider) GetContent(userIP string, count int) ([]*ContentItem, error) {
	return nil, errors.New("Unable to fetch the items, sorry")
}

func TestFallbacksAreRespected(t *testing.T) {
	// I am creating my own app config here so that the tests are less prone to
	// error. Reason being is that if the default config where to change, we
	// would also have to touch all the tests. By that logic, I would also have
	// to hard-code the config variables here - I won't do this for now for
	// clarity, but I would do so in real life.
	appAllWorkingProviders := App{
		Config: []ContentConfig{
			config1, config1, config2, config3, config4, config1, config1, config2,
		},
		ContentClients: map[Provider]Client{
			Provider1: SampleContentProvider{Source: Provider1},
			Provider2: SampleContentProvider{Source: Provider2},
			Provider3: SampleContentProvider{Source: Provider3},
		},
	}
	contentWithWorkingSources := runRequest(t, appAllWorkingProviders, SimpleContentRequest)

	if len(contentWithWorkingSources) != 5 {
		t.Fatalf("Got %d items back, want 5", len(contentWithWorkingSources))
	}

	if contentWithWorkingSources[0].Source != string(Provider1) {
		t.Fatalf("Expected provider 1 to be used for the first item.")
	}
	if contentWithWorkingSources[1].Source != string(Provider1) {
		t.Fatalf("Expected provider 1 to be used for the second item.")
	}
	if contentWithWorkingSources[2].Source != string(Provider2) {
		t.Fatalf("Expected provider 2 to be used for the third item.")
	}
	if contentWithWorkingSources[3].Source != string(Provider3) {
		t.Fatalf("Expected provider 3 to be used for the fourth item.")
	}
	if contentWithWorkingSources[4].Source != string(Provider1) {
		t.Fatalf("Expected provider 1 to be used for the fifth item.")
	}

	appWithBadProviders := App{
		Config: []ContentConfig{
			config1, config1, config2, config3, config4, config1, config1, config2,
		},
		ContentClients: map[Provider]Client{
			Provider1: SampleContentProvider{Source: Provider1},
			Provider2: FailingSampleProvider{Source: Provider2},
			Provider3: SampleContentProvider{Source: Provider3},
		},
	}
	contentWithFallbacks := runRequest(t, appWithBadProviders, SimpleContentRequest)
	if contentWithFallbacks[0].Source != string(Provider1) {
		t.Fatalf("Expected provider 1 to be used for the first item.")
	}
	if contentWithFallbacks[1].Source != string(Provider1) {
		t.Fatalf("Expected provider 1 to be used for the second item.")
	}
	if contentWithFallbacks[2].Source != string(Provider3) {
		t.Fatalf("Expected provider 3 to be used for the third item.")
	}
	if contentWithFallbacks[3].Source != string(Provider3) {
		t.Fatalf("Expected provider 3 to be used for the fourth item.")
	}
	if contentWithFallbacks[4].Source != string(Provider1) {
		t.Fatalf("Expected provider 1 to be used for the fifth item.")
	}
}

func TestListGetsCutOffIfSourceAndFallbackFail(t *testing.T) {
	mockAppWithBadResponders := App{
		Config: []ContentConfig{
			config1, config1, {Type: Provider2, Fallback: &Provider2}, config3,
		},
		ContentClients: map[Provider]Client{
			Provider1: SampleContentProvider{Source: Provider1},
			Provider2: FailingSampleProvider{Source: Provider2},
			Provider3: SampleContentProvider{Source: Provider3},
		},
	}
	content := runRequest(t, mockAppWithBadResponders, SimpleContentRequest)

	if len(content) != 2 {
		t.Fatalf("Got %d items back, want 2", len(content))
	}
}
