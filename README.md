# Golang Coding challenge

The dev team at [Sliide](https://sliide.com/) created this coding challenge to help assess your coding and problem solving skills
Along with this file, you should find an archive with the code od the project to complete.

# The Project

This is a simple http service that simulates a news API.

### Content
- The content itself is fetched from multiple providers (those could be 3rd party APIs, internal services, or database connections).

- Content providers are represented by the `Provider` type. And the API has a mapping between providers and `Clients` that are used to fetch content.

### Content configuration
- The API has configuration, which represents the repeating sequence of providers to use. if the sequence is [Provider1, Provider2, Provider3] and the user requests 5 articles, the response should contain items from [Provider1, Provider2, Provider3, Provider1, Provider2] in that order.

- In addition, if a provider fails to deliver content, the configuration might contain a fallback to use instead.

- In the case both the main provider and the fallback fail (or if the main provider fails and there is no fallback), the API should respond with all the items before that point.
So, for example, if the configuration calls for [1,1,2,3] and 2 fails, the response should only contain [1,1]

# The Interface

The API responds to GET requests, with 2 URL parameters:
- `count` represents the number of items desired
- `offset` represents the number of items previously requested. The configuration should be offset by this number.

The expected response is a list of content items, each one being a JSON representation of the `ContentItem` struct, found in `content.go`

Example request/response:
```
Request:
http '127.0.0.1:8080/?count=3&offset=10'

Response:
HTTP/1.1 200 OK
Content-Length: 385
Content-Type: application/json
Date: Thu, 24 Sep 2020 10:47:11 GMT

[
    {
        "expiry": "2020-09-24T11:47:11.204318471+01:00",
        "id": "5577006791947779410",
        "link": "",
        "source": "1",
        "summary": "",
        "title": "title"
    },
    {
        "expiry": "2020-09-24T11:47:11.204324536+01:00",
        "id": "8674665223082153551",
        "link": "",
        "source": "1",
        "summary": "",
        "title": "title"
    },
    {
        "expiry": "2020-09-24T11:47:11.204326896+01:00",
        "id": "6129484611666145821",
        "link": "",
        "source": "2",
        "summary": "",
        "title": "title"
    }
]

```

# Instructions

1. Complete the `ServeHTTP` method in server.go in accordance with the specifications above.
2. Run existing tests, and make sure they all pass
3. Add a few tests to capture missing edge-cases. For example, test that the fallbacks are respected.

Hints:
- You can run the server simply with `go run .` in the projects directory.
- Tests are run with `go test` in the current directory.
- Try to keep to the standard library as much as possible
- Latency is crucial for this application, so fetching the items sequentially one at a time might not be good enough
