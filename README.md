# pkghtml

Package pkghtml implements a http.Handler that renders package documentation.

## Usage

```go
package main

import (
    "net/http"
    "github.com/pnelson/pkghtml"
)

func main() {
    http.Handle("/pkg/net/", http.StripPrefix("/pkg/net/", pkghtml.New("net")))
    http.ListenAndServe(":3000", nil)
}
```
