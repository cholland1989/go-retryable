# go-retryable [![Documentation][doc-img]][doc] [![Build Status][ci-img]][ci]

Retryable HTTP client with configurable options for request delay, random
jitter, and exponential backoff.

## Installation

```bash
go get github.com/cholland1989/go-retryable
```

This library supports [version 1.20 and later][ver] of Go.

## Usage

```go
import "github.com/cholland1989/go-retryable/pkg/retryable"
import "github.com/cholland1989/go-retryable/pkg/unofficial"
```

Package [`retryable`](https://pkg.go.dev/github.com/cholland1989/go-retryable/pkg/retryable)
provides a retryable HTTP client with configurable options for request delay,
random jitter, and exponential backoff.

```go
request, err := http.NewRequest(http.MethodGet, "https://www.github.com/", nil)
if err != nil {
    log.Fatal(err)
}
response, err := retryable.DefaultClient.Do(request)
if err != nil {
    log.Fatal(err)
}
defer response.Body.Close()
```

Package [`unofficial`](https://pkg.go.dev/github.com/cholland1989/go-retryable/pkg/unofficial)
provides constants for well-known HTTP status codes that are not part of the
official specification.

See the [documentation][doc] for more details.

## License

Released under the [MIT License](LICENSE).

[ci]: https://github.com/cholland1989/go-retryable/actions/workflows/build.yml
[ci-img]: https://github.com/cholland1989/go-retryable/actions/workflows/build.yml/badge.svg
[doc]: https://pkg.go.dev/github.com/cholland1989/go-retryable
[doc-img]: https://pkg.go.dev/badge/github.com/cholland1989/go-retryable
[ver]: https://go.dev/doc/devel/release
