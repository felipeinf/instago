

# Instago

Instago is a unofficial Go SDK for develop on Instagram.

It is not an official product, has no relationship with Meta, and is not affiliated with Instagram. You are solely responsible for your use of this library.

**About this project.** instago is an independent library built for fun and as a personal technical challenge. It was **not created with commercial use in mind** and is not a commercial product, is not offered for sale, and is not intended as a substitute for any official Instagram tooling.

## Install

```bash
go get github.com/felipeinf/instago
```

The module path is `github.com/felipeinf/instago`, but the **Go package name is `ig`**. Use an import alias so calls stay readable, for example `import ig "github.com/felipeinf/instago"` and then `ig.NewClient()`.

## Quick start

### Login

Create a client, call `Login` with your username and password. If you do not use two-factor authentication, pass an empty string as the third argument.

```go
package main

import (
	"fmt"
	"log"

	"github.com/felipeinf/instago"
)

func main() {
	c := ig.NewClient()
	if err := c.Login("your_username", "your_password", ""); err != nil {
		log.Fatal(err)
	}
	fmt.Println("logged in")
	if err := c.DumpSettings("user.json"); err != nil {
		log.Fatal(err)
	}
}
```

If Instagram requires 2FA, pass the app/SMS code as the third argument.

### Save and restore session

The path is entirely up to you; `user.json` is only an example (often the process working directory). The login example above already saves to that path; you can also load an existing file before calling API methods:

```go
c2 := ig.NewClient()
_ = c2.LoadSettings("user.json", false)
```

### Look up a profile by username

Using a session file from a previous `Login` + `DumpSettings` (or any file you manage yourself):

```go
package main

import (
	"fmt"
	"log"

	"github.com/felipeinf/instago"
)

func main() {
	c := ig.NewClient()
	if err := c.LoadSettings("user.json", false); err != nil {
		log.Fatal(err)
	}
	u, err := c.UserInfoByUsername("google", true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(u.Username, u.FullName, u.FollowerCount)
}
```

The second argument to `UserInfoByUsername` enables the in-memory user cache; use `false` if you need a fresh fetch.

## Documentation

The SDK reference is godoc in the repository. The **package overview** in [`doc.go`](doc.go) lists the API in business order (login → session → users → media → reels → stories → direct → comments → friendship → search → low-level), including defaults for inbox/thread pagination and comment cursors; pkg.go.dev’s symbol index stays alphabetical by name.

Types (`User`, `Media`, `Settings`, …) live together in [`types.go`](types.go). `Client` methods and subpackages (`config`, `encoding`, `password`, `igerrors`) stay next to their implementations.

```bash
go doc github.com/felipeinf/instago
go doc -all github.com/felipeinf/instago
```

For a local browser view, install [pkgsite](https://pkg.go.dev/golang.org/x/pkgsite/cmd/pkgsite) and run it from this module root.

## License

[MIT](LICENSE): you may use, modify, and redistribute the code freely; the license includes a standard “as is” warranty disclaimer. The separate [legal disclaimer](#legal-disclaimer) at the end of this file also applies.

## Legal disclaimer

**What this is.** The following is a general **informational disclaimer** about the software and the project. It is **not** tailored to your situation or jurisdiction and **does not** create an attorney–client relationship.

This notice is written with **U.S. law** in mind; your rights and duties may differ elsewhere.

instago is **not** affiliated with, endorsed by, maintained by, or sponsored by **Meta Platforms, Inc.**, **Instagram**, or their affiliates. *Instagram* is a trademark of Meta Platforms, Inc.

The software is provided **“AS IS”** and **“AS AVAILABLE”**, without warranties of any kind, whether express or implied, including implied warranties of merchantability, fitness for a particular purpose, and non-infringement. **Your use is at your sole risk.** To the fullest extent permitted by law, the authors and contributors disclaim liability for any direct, indirect, incidental, special, consequential, or exemplary damages, and for any loss of data, account access, goodwill, or other intangible losses, arising out of or related to your use or inability to use this software—including claims arising from violation of Instagram’s terms, policies, or applicable law, or from misuse by you or third parties.

The project exists for **personal learning and experimentation** only. You are solely responsible for compliance with Instagram’s terms and all laws that apply to you. The maintainers do not encourage or support unlawful, abusive, or harmful use.

*The above is for general information only. It is **not** legal advice.*