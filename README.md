# Instago

Instago is a unofficial Go SDK for develop on Instagram.

It is not an official product, has no relationship with Meta, and is not affiliated with Instagram. You are solely responsible for your use of this library.

**About this project.** instago is an independent library built for fun and as a personal technical challenge. It was **not created with commercial use in mind** and is not a commercial product, is not offered for sale, and is not intended as a substitute for any official Instagram tooling.

## Install

```bash
go get github.com/felipeinf/instago
```

The module path is `github.com/felipeinf/instago`, but the **Go package name is `ig`**. Example: `import ig "github.com/felipeinf/instago"` and `ig.NewClient()`.

## Documentation

- **[docs/](docs/README.md)** — API overview, minimal examples, and type list (easy to browse on GitHub).
- **[pkg.go.dev](https://pkg.go.dev/github.com/felipeinf/instago)** — full reference and package overview (`doc.go`).

## Quick start

```go
package main

import (
	"fmt"
	"log"

	"github.com/felipeinf/instago"
)

func main() {
	c := ig.NewClient()
	if err := c.Login("username", "password", ""); err != nil {
		log.Fatal(err)
	}
	if err := c.DumpSettings("session.json"); err != nil {
		log.Fatal(err)
	}
	me, err := c.AccountInfo()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("logged in as", me.Username)
}
```

Use `LoadSettings("session.json", false)` instead of `Login` when you already have a saved session. More patterns: [docs/api.md](docs/api.md).

## License

[MIT](LICENSE): you may use, modify, and redistribute the code freely; the license includes a standard “as is” warranty disclaimer. The separate [legal disclaimer](#legal-disclaimer) at the end of this file also applies.

## Legal disclaimer

**What this is.** The following is a general **informational disclaimer** about the software and the project. It is **not** tailored to your situation or jurisdiction and **does not** create an attorney–client relationship.

This notice is written with **U.S. law** in mind; your rights and duties may differ elsewhere.

instago is **not** affiliated with, endorsed by, maintained by, or sponsored by **Meta Platforms, Inc.**, **Instagram**, or their affiliates. *Instagram* is a trademark of Meta Platforms, Inc.

The software is provided **“AS IS”** and **“AS AVAILABLE”**, without warranties of any kind, whether express or implied, including implied warranties of merchantability, fitness for a particular purpose, and non-infringement. **Your use is at your sole risk.** To the fullest extent permitted by law, the authors and contributors disclaim liability for any direct, indirect, incidental, special, consequential, or exemplary damages, and for any loss of data, account access, goodwill, or other intangible losses, arising out of or related to your use or inability to use this software—including claims arising from violation of Instagram’s terms, policies, or applicable law, or from misuse by you or third parties.

The project exists for **personal learning and experimentation** only. You are solely responsible for compliance with Instagram’s terms and all laws that apply to you. The maintainers do not encourage or support unlawful, abusive, or harmful use.

*The above is for general information only. It is **not** legal advice.*
