package pkg

import (
	"net/http"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

const retryAttempts = 3
const retryDelay = 10

func HttpClient() *http.Client {
	client := retryablehttp.NewClient()
	client.RetryMax = retryAttempts
	client.RetryWaitMin = time.Duration(retryDelay) * time.Second
	client.RetryWaitMax = client.RetryWaitMin
	return client.StandardClient()
}
