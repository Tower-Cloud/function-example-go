# Tower function example — Go

A Go function that builds and serves as-is. Fork it, or copy `go.mod` and `main.go` into
your own repository.

## Contract

Tower compiles `package main` at the root of the source path and runs the resulting binary.
That binary must:

- listen on the port in `os.Getenv("PORT")`
- bind `0.0.0.0`, not `localhost`
- keep running

Nothing else in the repository is consulted, and there is no handler signature or Tower
package to import.

## Port

The port is a setting on the function, `8080` by default. Tower sets `PORT` to that value,
so read it rather than hardcoding a number.

The port is internal. It does not appear in the function's URL, and changing it does not
change the URL.

## Go version

The runtime you pick on the function decides the Go version. It is not a starting point that
`go.mod` can raise.

Keep the `go` directive at or below the runtime you select. Asking for a newer Go fails the
build with `go.mod requires go >= <version>`, rather than quietly building with that newer
version:

```
go: go.mod requires go >= 1.26 (running go 1.25.9; GOTOOLCHAIN=local)
```

A `toolchain` directive is ignored rather than honoured, so it cannot raise the version
either — and cannot rescue a `go` directive that is too high. This file says `go 1.25` and
sets no `toolchain`, which builds on every runtime currently offered.

## Common failures

Both of these build successfully and then fail to serve.

Work done directly in `main`, with no server. The process exits 0, and Tower reports that
the function exited by itself:

```go
func main() {
	fmt.Println("hello")
}
```

A server bound to localhost, reachable only from inside its own instance:

```go
srv := &http.Server{Addr: "127.0.0.1:" + port}
```

## Deploy

1. Fork this repository into an organization connected to Tower.
2. Create a function with `deploymentMode: "source"`, `runtime: "go1.26"`,
   `visibility: "public"` and this repository as the source. Leave the source path empty.

   `visibility` is required and has no default, so leaving it out fails the request.
   `"public"` is what gives the function a URL. A private function builds and runs
   normally but is reachable only from inside Tower, which leaves nothing to curl.
3. Release the build. Builds are not released automatically unless release-on-push is on.

```sh
curl https://<function-url>/          # {"message":"Hello from Tower.", ...}
curl https://<function-url>/healthz   # ok
```

To serve only some paths, add `httpRules` — anything not listed is refused before it
reaches the function. Leaving it out serves every path, which is what this example does.

## Notes

- `/healthz` is not special. Tower does not probe it, and a function without one is
  healthy exactly the same; it is here because it is a convenient thing to curl.
- No dependencies. The standard library only.
- Chi, Gin, Echo and others work the same way; the contract is the port and staying up.
- No workflow file. Enabling auto-deploy opens a pull request that adds one.
