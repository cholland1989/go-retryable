package retryable

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type MockReader struct {
	state bool
}

func (mock *MockReader) Read(_ []byte) (int, error) {
	mock.state = !mock.state
	if mock.state {
		return 0, io.ErrUnexpectedEOF
	}
	return 0, io.EOF
}

func ExampleClient() {
	request, err := http.NewRequest(http.MethodGet, "https://www.github.com/", nil)
	if err != nil {
		log.Fatal(err)
	}
	response, err := DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	fmt.Println(response.StatusCode)
	// Output: 200
}

func TestClient_CloseIdleConnections(test *testing.T) {
	test.Parallel()

	client := new(Client)
	client.CloseIdleConnections()
}

func TestClient_Get(test *testing.T) {
	test.Parallel()

	client := new(Client)
	response, err := client.Get("https://www.github.com/")
	require.NoError(test, err)
	require.NotNil(test, response)

	response, err = client.Get(string([]byte{0x7F}))
	require.ErrorIs(test, err, ErrNonRetryable)
	require.Nil(test, response)
}

func TestClient_Head(test *testing.T) {
	test.Parallel()

	client := new(Client)
	response, err := client.Head("https://www.github.com/")
	require.NoError(test, err)
	require.NotNil(test, response)

	response, err = client.Head(string([]byte{0x7F}))
	require.ErrorIs(test, err, ErrNonRetryable)
	require.Nil(test, response)
}

func TestClient_Post(test *testing.T) {
	test.Parallel()

	client := new(Client)
	response, err := client.Post("https://www.github.com/", "text/plain", nil)
	require.NoError(test, err)
	require.NotNil(test, response)

	response, err = client.Post(string([]byte{0x7F}), "text/plain", nil)
	require.ErrorIs(test, err, ErrNonRetryable)
	require.Nil(test, response)
}

func TestClient_PostForm(test *testing.T) {
	test.Parallel()

	client := new(Client)
	response, err := client.PostForm("https://www.github.com/", nil)
	require.NoError(test, err)
	require.NotNil(test, response)

	data := make(url.Values)
	response, err = client.PostForm("https://www.github.com/", data)
	require.NoError(test, err)
	require.NotNil(test, response)
}

func TestClient_Do(test *testing.T) {
	test.Parallel()

	client := new(Client)
	response, err := client.Do(nil)
	require.ErrorIs(test, err, ErrNonRetryable)
	require.Nil(test, response)

	request := new(http.Request)
	response, err = client.Do(request)
	require.ErrorIs(test, err, ErrRetryable)
	require.Nil(test, response)

	url, err := url.Parse("https://www.github.com/")
	require.NoError(test, err)
	require.NotNil(test, url)

	request.Method = http.MethodGet
	request.URL = url
	response, err = client.Do(request)
	require.NoError(test, err)
	require.NotNil(test, response)
	require.Greater(test, response.ContentLength, int64(1))

	client.RetryCount = 1
	client.RetryStatus = append(client.RetryStatus, 200)
	response, err = client.Do(request)
	require.ErrorIs(test, err, ErrRetryable)
	require.NotNil(test, response)
	require.Greater(test, response.ContentLength, int64(1))

	client.RetryTimeout = time.Millisecond
	response, err = client.Do(request)
	require.ErrorIs(test, err, ErrNonRetryable)
	require.ErrorIs(test, err, context.DeadlineExceeded)
	require.Nil(test, response)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	response, err = client.Do(request.WithContext(ctx))
	require.ErrorIs(test, err, ErrNonRetryable)
	require.ErrorIs(test, err, context.Canceled)
	require.Nil(test, response)

	client.RetryDelay = time.Second
	response, err = client.Do(new(http.Request))
	require.ErrorIs(test, err, ErrNonRetryable)
	require.ErrorIs(test, err, context.DeadlineExceeded)
	require.Nil(test, response)

	request.GetBody = func() (io.ReadCloser, error) { return nil, io.EOF }
	response, err = client.Do(request)
	require.ErrorIs(test, err, ErrNonRetryable)
	require.ErrorIs(test, err, io.EOF)
	require.Nil(test, response)
}

func TestClient_PanicHandler(test *testing.T) {
	test.Parallel()

	require.Panics(test, func() {
		client := new(Client)
		defer client.panicHandler(nil)
		panic("runtime error")
	})

	err := func() (err error) {
		client := new(Client)
		defer client.panicHandler(&err)
		panic("runtime error")
	}()

	require.ErrorIs(test, err, ErrNonRetryable)
	require.ErrorContains(test, err, "runtime error")
}

func TestClient_PrepareRequestBody(test *testing.T) {
	test.Parallel()

	client := new(Client)
	err := client.prepareRequestBody(nil)
	require.ErrorIs(test, err, ErrNonRetryable)

	request := new(http.Request)
	err = client.prepareRequestBody(request)
	require.NoError(test, err)

	request.Body = io.NopCloser(strings.NewReader("xyz"))
	err = client.prepareRequestBody(request)
	require.NoError(test, err)

	reader, err := request.GetBody()
	require.NoError(test, err)
	buffer, err := io.ReadAll(reader)
	require.NoError(test, err)
	require.Equal(test, "xyz", string(buffer))

	function := &request.GetBody
	err = client.prepareRequestBody(request)
	require.NoError(test, err)
	require.Equal(test, function, &request.GetBody)

	client.RequestSize = 1
	request.GetBody = nil
	err = client.prepareRequestBody(request)
	require.ErrorIs(test, err, ErrNonRetryable)
	require.ErrorContains(test, err, "3")

	reader, err = request.GetBody()
	require.NoError(test, err)
	buffer, err = io.ReadAll(reader)
	require.NoError(test, err)
	require.Equal(test, "x", string(buffer))

	request.Body = io.NopCloser(new(MockReader))
	request.GetBody = nil
	err = client.prepareRequestBody(request)
	require.ErrorIs(test, err, ErrNonRetryable)
	require.ErrorIs(test, err, io.ErrUnexpectedEOF)

	request.GetBody = nil
	err = client.prepareRequestBody(request)
	require.ErrorIs(test, err, ErrNonRetryable)
	require.ErrorIs(test, err, io.ErrUnexpectedEOF)
}

func TestClient_ApplyRequestDelay(test *testing.T) {
	test.Parallel()

	client := new(Client)
	applyRequestDelay := func(ctx context.Context) (time.Duration, error) {
		timestamp := time.Now()
		err := client.applyRequestDelay(ctx)
		return time.Since(timestamp), err
	}

	duration, err := applyRequestDelay(nil)
	require.NoError(test, err)
	require.Less(test, duration, time.Millisecond)

	client.RequestDelay = time.Millisecond
	duration, err = applyRequestDelay(nil)
	require.NoError(test, err)
	require.GreaterOrEqual(test, duration, time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	duration, err = applyRequestDelay(ctx)
	require.ErrorIs(test, err, context.Canceled)
	require.Less(test, duration, time.Millisecond)
}

func TestClient_ResetRequestBody(test *testing.T) {
	test.Parallel()

	client := new(Client)
	request := new(http.Request)
	err := client.resetRequestBody(request)
	require.NoError(test, err)

	reader := io.NopCloser(strings.NewReader("xyz"))
	request.GetBody = func() (io.ReadCloser, error) { return reader, nil }
	err = client.resetRequestBody(request)
	require.NoError(test, err)
	require.Equal(test, reader, request.Body)

	request.GetBody = func() (io.ReadCloser, error) { return nil, io.EOF }
	err = client.resetRequestBody(request)
	require.ErrorIs(test, err, ErrNonRetryable)
	require.ErrorIs(test, err, io.EOF)
	require.Nil(test, request.Body)
}

func TestClient_SendRequest(test *testing.T) {
	test.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	client := new(Client)
	request := new(http.Request)
	response, err := client.sendRequest(ctx, request)
	require.ErrorIs(test, err, ErrRetryable)
	require.Nil(test, response)

	url, err := url.Parse("https://www.github.com/")
	require.NoError(test, err)
	require.NotNil(test, url)

	request.Method = http.MethodGet
	request.URL = url
	response, err = client.sendRequest(ctx, request)
	require.NoError(test, err)
	require.NotNil(test, response)
	require.Greater(test, response.ContentLength, int64(1))

	client.ResponseSize = 1
	response, err = client.sendRequest(ctx, request)
	require.ErrorIs(test, err, ErrNonRetryable)
	require.NotNil(test, response)
	require.Equal(test, int64(1), response.ContentLength)

	cancel()
	response, err = client.sendRequest(ctx, request)
	require.ErrorIs(test, err, ErrNonRetryable)
	require.ErrorIs(test, err, context.Canceled)
	require.Nil(test, response)

	ctx = context.Background()
	client.RequestTimeout = 1
	response, err = client.sendRequest(ctx, request)
	require.ErrorIs(test, err, ErrNonRetryable)
	require.ErrorIs(test, err, context.DeadlineExceeded)
	require.Nil(test, response)
}

func TestClient_PrepareResponseBody(test *testing.T) {
	test.Parallel()

	client := new(Client)
	response := new(http.Response)
	reader := io.NopCloser(strings.NewReader("xyz"))
	response.Body = reader
	err := client.prepareResponseBody(response)
	require.NoError(test, err)
	require.NotEqual(test, response.Body, reader)

	buffer, err := io.ReadAll(response.Body)
	require.NoError(test, err)
	require.Equal(test, "xyz", string(buffer))

	client.ResponseSize = 1
	response.Body = io.NopCloser(strings.NewReader("xyz"))
	err = client.prepareResponseBody(response)
	require.ErrorIs(test, err, ErrNonRetryable)
	require.ErrorContains(test, err, "3")

	buffer, err = io.ReadAll(response.Body)
	require.NoError(test, err)
	require.Equal(test, "x", string(buffer))

	response.Body = io.NopCloser(new(MockReader))
	err = client.prepareResponseBody(response)
	require.ErrorIs(test, err, ErrRetryable)
	require.ErrorIs(test, err, io.ErrUnexpectedEOF)

	err = client.prepareResponseBody(response)
	require.ErrorIs(test, err, ErrRetryable)
	require.ErrorIs(test, err, io.ErrUnexpectedEOF)

	response.StatusCode = 400
	err = client.prepareResponseBody(response)
	require.ErrorIs(test, err, ErrNonRetryable)
	require.ErrorContains(test, err, "400")

	client.RetryStatus = append(client.RetryStatus, 400)
	err = client.prepareResponseBody(response)
	require.ErrorIs(test, err, ErrRetryable)
	require.ErrorContains(test, err, "400")
}

func TestClient_ApplyRetryDelay(test *testing.T) {
	test.Parallel()

	client := new(Client)
	applyRetryDelay := func(ctx context.Context, response *http.Response, attempt int) (time.Duration, error) {
		timestamp := time.Now()
		err := client.applyRetryDelay(ctx, response, attempt)
		return time.Since(timestamp), err
	}

	duration, err := applyRetryDelay(nil, nil, 0)
	require.NoError(test, err)
	require.Less(test, duration, time.Millisecond)

	client.RetryDelay = time.Millisecond
	duration, err = applyRetryDelay(nil, nil, 0)
	require.NoError(test, err)
	require.GreaterOrEqual(test, duration, time.Millisecond)

	client.RetryMultiplier = 0.5
	duration, err = applyRetryDelay(nil, nil, 0)
	require.NoError(test, err)
	require.GreaterOrEqual(test, duration, time.Millisecond)

	client.RetryMultiplier = 2.0
	duration, err = applyRetryDelay(nil, nil, 0)
	require.NoError(test, err)
	require.GreaterOrEqual(test, duration, 2*time.Millisecond)

	response := new(http.Response)
	response.Header = make(http.Header)
	response.Header["Retry-After"] = make([]string, 1)
	duration, err = applyRetryDelay(nil, response, 0)
	require.NoError(test, err)
	require.GreaterOrEqual(test, duration, 2*time.Millisecond)

	response.Header["Retry-After"][0] = "1"
	duration, err = applyRetryDelay(nil, response, 0)
	require.NoError(test, err)
	require.GreaterOrEqual(test, duration, time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	duration, err = applyRetryDelay(ctx, nil, 0)
	require.ErrorIs(test, err, context.Canceled)
	require.Less(test, duration, time.Millisecond)

	duration, err = applyRetryDelay(ctx, response, 0)
	require.ErrorIs(test, err, context.Canceled)
	require.Less(test, duration, time.Millisecond)
}

func TestClient_ParseRetryDelay(test *testing.T) {
	test.Parallel()

	client := new(Client)
	delay := client.parseRetryDelay(nil)
	require.Zero(test, delay)

	response := new(http.Response)
	delay = client.parseRetryDelay(response)
	require.Zero(test, delay)

	response.Header = make(http.Header)
	delay = client.parseRetryDelay(response)
	require.Zero(test, delay)

	response.Header["Retry-After"] = make([]string, 1)
	delay = client.parseRetryDelay(response)
	require.Zero(test, delay)

	response.Header["Retry-After"][0] = "xyz"
	delay = client.parseRetryDelay(response)
	require.Zero(test, delay)

	response.Header["Retry-After"][0] = "1"
	delay = client.parseRetryDelay(response)
	require.Equal(test, time.Second, delay)

	date := time.Now().Add(time.Minute).Format(time.RFC1123)
	response.Header["Retry-After"][0] = date
	delay = client.parseRetryDelay(response)
	require.Greater(test, delay, time.Minute-time.Second)
	require.Less(test, delay, time.Minute)
}
