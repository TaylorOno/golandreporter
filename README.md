# Goland Ginkgo Reporter

Ginkgo does not output results in a format which can be converted into JSON 
(via [`go tool test2json`](https://golang.org/cmd/test2json/), which is used by
Goland to get a list of output.

This reporter will trick test2json into outputting the Ginkgo specs similar to
`go test` output.

![Ginkgo output in Goland's "Run" window](https://gist.githubusercontent.com/TaylorOno/48f9a7a243f0d0b89f89d176060ede33/raw/5bdfb5935293dfdfab7c92a6822ecbbb4e2fccc2/screenshot.png)

## Usage

In your suite replace `RunSpecs(t, "Integration Suite")` with the following:

```go
golandReporter := golandreporter.NewGolandReporter()
RunSpecsWithCustomReporters(t, "Integration Suite", []Reporter{golandReporter})
```

If you want to retain normal Ginkgo formatting when using it from the CLI the
best option is to use an environment variable in your Run Configuration, and
use it like this:

```go
RunSpecsWithCustomReporters(t, "Integration Suite", []Reporter{golandreporter.NewAutoGolandReporter()})
```

