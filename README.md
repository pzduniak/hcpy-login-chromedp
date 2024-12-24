# hcpy-login-chromedp

A proof of concept for Home Connect profile retrieval implemented as a ChromeDP session controller. Intended to be a solution for https://github.com/hcpy2-0/hcpy/issues/116

A mix between the following 2 implementations:

- https://github.com/bruestel/homeconnect-profile-downloader
- https://github.com/hcpy2-0/hcpy/pull/117

In order to perform the dump, `go build` the repo, run the binary and log into HC.
Eventually all the bundles will be dumped into an `output` directory, relative to
the current working directory path.
