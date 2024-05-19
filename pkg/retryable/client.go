// Package retryable provides a retryable HTTP client with configurable options
// for request delay, random jitter, and exponential backoff.
package retryable

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/cholland1989/go-delay/pkg/sleep"
	"github.com/cholland1989/go-retryable/pkg/unofficial"
)

// ErrRetryable defines a retryable error.
var ErrRetryable = errors.New("retryable error")

// ErrNonRetryable defines a non-retryable error.
var ErrNonRetryable = errors.New("non-retryable error")

// DefaultClient is the default retryable HTTP client.
var DefaultClient = &Client{
	Client:          *http.DefaultClient,
	RetryStatus:     DefaultStatus,
	RetryCount:      20,
	RetryDelay:      500 * time.Millisecond,
	RetryMultiplier: 1.5,
	RetryJitter:     0.5,
	RetryTimeout:    60 * time.Minute,
	RequestDelay:    10 * time.Millisecond,
	RequestJitter:   0.5,
	RequestTimeout:  5 * time.Minute,
	RequestSize:     2 * 1024 * 1024 * 1024,
	ResponseSize:    2 * 1024 * 1024 * 1024,
}

// DefaultStatus contains the default retryable status codes.
var DefaultStatus = []int{
	http.StatusRequestTimeout,
	http.StatusConflict,
	unofficial.StatusEnhanceYourCalm,
	http.StatusLocked,
	http.StatusTooEarly,
	http.StatusTooManyRequests,
	unofficial.StatusRequestHeaderFieldsTooLarge,
	http.StatusInternalServerError,
	http.StatusBadGateway,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
	http.StatusInsufficientStorage,
	unofficial.StatusBandwidthLimitExceeded,
	unofficial.StatusWebServerReturnedAnUnknownError,
	unofficial.StatusWebServerIsDown,
	unofficial.StatusConnectionTimedOut,
	unofficial.StatusOriginIsUnreachable,
	unofficial.StatusTimeoutOccurred,
	unofficial.StatusRailgunError,
	unofficial.StatusSiteIsOverloaded,
	unofficial.StatusCloudflareError,
	unofficial.StatusNetworkReadTimeout,
	unofficial.StatusNetworkConnectTimeout,
}

// Client is an HTTP client that can automatically retry failed requests, and
// provides a drop-in replacement for [net/http.Client].
type Client struct {
	// Client specifies the base HTTP client.
	http.Client

	// RetryStatus specifies the status codes that are retryable.
	RetryStatus []int

	// RetryCount specifies the maximum number of retries per request.
	RetryCount int

	// RetryDelay specifies the delay between retries.
	RetryDelay time.Duration

	// RetryMultiplier specifies the exponential backoff multiplier for the
	// retry delay. If the retry multiplier is less than one, it will be
	// ignored.
	RetryMultiplier float64

	// RetryJitter specifies the random jitter applied to the retry delay.
	RetryJitter float64

	// RetryTimeout specifies the maximum total duration of retries per request.
	RetryTimeout time.Duration

	// RequestDelay specifies a fixed delay applied to each request.
	RequestDelay time.Duration

	// RequestJitter specifies the random jitter applied to the request delay.
	RequestJitter float64

	// RequestTimeout specifies the maximum duration per request.
	RequestTimeout time.Duration

	// RequestSize specifies the maximum request size in bytes.
	RequestSize int64

	// ResponseSize specifies the maximum response size in bytes.
	ResponseSize int64
}

// CloseIdleConnections closes any connections on its [net/http.Transport]
// which were previously connected from previous requests but are now sitting
// idle in a "keep-alive" state. It does not interrupt any connections
// currently in use.
func (client *Client) CloseIdleConnections() {
	client.Client.CloseIdleConnections()
}

// Get issues a GET to the specified URL.
func (client *Client) Get(url string) (response *http.Response, err error) {
	// Construct and send HTTP request
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to construct request: %w", ErrNonRetryable, err)
	}
	return client.Do(request)
}

// Head issues a HEAD to the specified URL.
func (client *Client) Head(url string) (response *http.Response, err error) {
	// Construct and send HTTP request
	request, err := http.NewRequestWithContext(context.Background(), http.MethodHead, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to construct request: %w", ErrNonRetryable, err)
	}
	return client.Do(request)
}

// Post issues a POST to the specified URL.
func (client *Client) Post(url string, contentType string, body io.Reader) (response *http.Response, err error) {
	// Construct and send HTTP request
	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to construct request: %w", ErrNonRetryable, err)
	}
	request.Header.Set("Content-Type", contentType)
	return client.Do(request)
}

// PostForm issues a POST to the specified URL, with data's keys and values
// URL-encoded as the request body.
func (client *Client) PostForm(url string, data url.Values) (response *http.Response, err error) {
	// Construct and send HTTP request
	if data != nil {
		return client.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	}
	return client.Post(url, "application/x-www-form-urlencoded", nil)
}

// Do sends an HTTP request and returns an HTTP response, following policy
// (such as redirects, cookies, auth) as configured on the client.
func (client *Client) Do(request *http.Request) (response *http.Response, err error) {
	// Convert panics into an error
	defer client.panicHandler(&err)

	// Extract seekable request body
	err = client.prepareRequestBody(request)
	if err != nil {
		return nil, err
	}

	// Apply retry timeout to context
	ctx := request.Context()
	if client.RetryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, client.RetryTimeout)
		defer cancel()
	}

	// Retry failed requests
	for attempt := 0; attempt <= client.RetryCount; attempt++ {
		// Apply fixed request delay
		err = client.applyRequestDelay(ctx)
		if err != nil {
			return response, err
		}

		// Reset request body
		err = client.resetRequestBody(request)
		if err != nil {
			return response, err
		}

		// Send request and receive response
		response, err = client.sendRequest(ctx, request)
		if err == nil {
			return response, nil
		}

		// Check for non-retryable error
		if !errors.Is(err, ErrRetryable) {
			return response, err
		}

		// Apply exponential retry delay
		if attempt < client.RetryCount {
			err = client.applyRetryDelay(ctx, response, attempt)
			if err != nil {
				return response, err
			}
		}
	}
	return response, err
}

// panicHandler recovers panics and converts them into an error, replacing the
// specified error.
func (client *Client) panicHandler(err *error) {
	// Check for valid error pointer
	if err == nil {
		return
	}

	// Convert panic into error
	cause := recover()
	if cause != nil {
		*err = fmt.Errorf("%w: %v: %s", ErrNonRetryable, cause, string(debug.Stack()))
	}
}

// prepareRequestBody ensures that the request body can be reset between retry
// attempts. If the request body is nil or the GetBody method is already set,
// the request is not modified. Otherwise the request body is read into memory
// and the GetBody method is updated.
func (client *Client) prepareRequestBody(request *http.Request) (err error) {
	// Check for valid request
	if request == nil {
		return fmt.Errorf("%w: invalid request", ErrNonRetryable)
	}

	// Check for valid request body
	if request.Body == nil || request.GetBody != nil {
		return nil
	}

	// Limit request size
	reader := io.Reader(request.Body)
	if client.RequestSize > 0 {
		reader = io.LimitReader(reader, client.RequestSize)
	}

	// Read request body
	buffer, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("%w: unable to read request body: %w", ErrNonRetryable, err)
	}

	// Replace request body
	defer func(buffer []byte) {
		_ = request.Body.Close()
		request.ContentLength = int64(len(buffer))
		request.Body = io.NopCloser(bytes.NewReader(buffer))
		request.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(buffer)), nil
		}
	}(buffer)

	// Discard remaining request body
	size, err := io.Copy(io.Discard, request.Body)
	if err != nil {
		return fmt.Errorf("%w: unable to discard request body: %w", ErrNonRetryable, err)
	}

	// Check for valid request size
	size += client.RequestSize
	if client.RequestSize > 0 && size > client.RequestSize {
		return fmt.Errorf("%w: request size exceeded (%d)", ErrNonRetryable, size)
	}
	return nil
}

// applyRequestDelay applies a fixed backoff with random jitter to each
// request, returning an error if the context is canceled.
func (client *Client) applyRequestDelay(ctx context.Context) (err error) {
	// Sleep for a fixed duration with random jitter
	err = sleep.RandomJitterWithContext(ctx, client.RequestDelay, client.RequestJitter)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNonRetryable, err)
	}
	return nil
}

// resetRequestBody resets the request body so that the request can be retried.
func (client *Client) resetRequestBody(request *http.Request) (err error) {
	// Check for valid request body
	if request == nil || request.GetBody == nil {
		return nil
	}

	// Reset request body
	request.Body, err = request.GetBody()
	if err != nil {
		return fmt.Errorf("%w: unable to reset request body: %w", ErrNonRetryable, err)
	}
	return nil
}

// sendRequest sends the request with the configured HTTP client, validates
// the response, and reads the response body into memory.
func (client *Client) sendRequest(ctx context.Context, request *http.Request) (response *http.Response, err error) {
	// Apply request timeout to context
	if client.RequestTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, client.RequestTimeout)
		defer cancel()
	}

	// Send request and receive response
	response, err = client.Client.Do(request.WithContext(ctx))

	// Check that context is valid
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return response, fmt.Errorf("%w: %w", ErrNonRetryable, err)
	}

	// Check for error sending request
	if err != nil {
		return response, fmt.Errorf("%w: unable to send request: %w", ErrRetryable, err)
	}

	// Check for valid response
	if response == nil || response.Body == nil {
		return response, fmt.Errorf("%w: invalid response", ErrRetryable)
	}

	// Read and replace response body
	err = client.prepareResponseBody(response)
	if err != nil {
		return response, err
	}
	return response, nil
}

// prepareResponseBody reads the response body into memory, validates the
// status code, and validates the response size.
func (client *Client) prepareResponseBody(response *http.Response) (err error) {
	// Close response body
	defer func(body io.Closer) {
		_ = body.Close()
	}(response.Body)

	// Limit response size
	reader := io.Reader(response.Body)
	if client.ResponseSize > 0 {
		reader = io.LimitReader(reader, client.ResponseSize)
	}

	// Read response body
	buffer, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("%w: unable to read response body: %w", ErrRetryable, err)
	}

	// Replace response body
	defer func(buffer []byte) {
		response.ContentLength = int64(len(buffer))
		body := bytes.NewReader(buffer)
		response.Body = io.NopCloser(body)
	}(buffer)

	// Discard remaining response body
	size, err := io.Copy(io.Discard, response.Body)
	if err != nil {
		return fmt.Errorf("%w: unable to discard response body: %w", ErrRetryable, err)
	}

	// Check for retryable status code
	for _, status := range client.RetryStatus {
		if status == response.StatusCode {
			return fmt.Errorf("%w: invalid status code (%d)", ErrRetryable, response.StatusCode)
		}
	}

	// Check for non-retryable status code
	if response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("%w: invalid status code (%d)", ErrNonRetryable, response.StatusCode)
	}

	// Check for valid response size
	size += client.ResponseSize
	if client.ResponseSize > 0 && size > client.ResponseSize {
		return fmt.Errorf("%w: response size exceeded (%d)", ErrNonRetryable, size)
	}
	return nil
}

// applyRetryDelay applies an exponential backoff with random jitter to each
// retry, returning an error if the context is canceled. If the retry header
// is present and valid, it is used (without random jitter) instead of an
// exponential backoff.
func (client *Client) applyRetryDelay(ctx context.Context, response *http.Response, attempt int) (err error) {
	// Check for valid retry header
	delay := client.parseRetryDelay(response)
	if delay > 0 {
		// Sleep for a fixed duration without random jitter
		err = sleep.RandomJitterWithContext(ctx, delay, 0.0)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrNonRetryable, err)
		}
		return nil
	}

	// Ensure the retry multiplier is valid when unset
	multiplier := math.Max(client.RetryMultiplier, 1.0)

	// Sleep for an exponential duration with random jitter
	err = sleep.ExponentialBackoffWithContext(ctx, client.RetryDelay, multiplier, client.RetryJitter, attempt)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNonRetryable, err)
	}
	return nil
}

// parseRetryDelay attempts to parse the retry header for either a duration
// in seconds or a date in [time.RFC1123] format, returning a non-zero
// [time.Duration] if the retry header is present and valid.
func (client *Client) parseRetryDelay(response *http.Response) (delay time.Duration) {
	// Check for valid response headers
	if response == nil || response.Header == nil {
		return 0
	}

	// Check for valid retry header
	header := response.Header.Get("Retry-After")
	if header == "" {
		return 0
	}

	// Attempt to parse retry header as duration
	duration, err := strconv.ParseInt(header, 10, 64)
	if err == nil {
		return time.Duration(duration) * time.Second
	}

	// Attempt to parse retry header as date
	date, err := time.Parse(time.RFC1123, header)
	if err == nil {
		return time.Until(date)
	}
	return 0
}
