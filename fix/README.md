## client api changes

`goirc.go` is a fix compatible with `go tool fix` that attempts to
programaticallly change old client code to work with the newer, shinier
API. It might even work, depending on the complexity of your code. You'll
need to have a Go source tree in the path specified by GOROOT. GOOS and
GOARCH may not be set in your environment; use your head.

### To install:

    cd $GOROOT/src/cmd/fix
    ln -s $GOPATH/src/github.com/fluffle/goirc/fix/goirc.go
    go build
    mv fix $GOROOT/pkg/tool/$GOOS_$GOARCH

### To fix:

    go tool fix -r goirc -diff /path/to/code
    <check diffs>
    go tool fix -r goirc /path/to/code

### Things that aren't fixed by this

This fix doesn't take care of some bad design decisions I made:

  - conn.State is left as-is. If you're using this you'll need to rewrite
    things to get scope into your handlers in a different fashion.
  - conn.ER and conn.ED are left as-is. These should be completely removed
    (and probably shouldn't have been left accessible in the first place).

It's also quite likely that this won't produce the nicest "fixed" code. In
particular, if you're seeing lots of lines like `conn.Config().XXX = "foo"`,
you probably want to consider creating a `Config` struct and then passing
it to `client.Client()`.

