# Marshaler (WIP)

`Unmarshal` and `Marshal` generators for Go. For use in [Bud](https://github.com/livebud/bud).

## Example

Given the following struct:

```go
type Input struct {
  B string
  C int
  D float64
  E bool
  F map[string]string
  G []int
  H *string
}
```

This library will generate an `UnmarshalJSON` function that can take arbitrary JSON and produce `Input`:

```go
func UnmarshalJSON(json []byte, in *Input) error {
  // Generated code
}
```

Then usage would look something like this:

```go
func ReadBody(r io.ReadCloser) (*Input, error) {
  json, err := io.ReadAll(r)
  if err != nil {
    return nil, err
  }
  in := new(Input)
  if err := UnmarshalJSON(json, in); err != nil {
    return nil, err
  }
  return in, nil
}
```

## TODO

This package is still very much WIP. There's a lot more work to do:

- [x] Unmarshal nested structs
- [x] Get tests running from a temporary directory
- [ ] Unmarshal referenced types
- [ ] Support the json tag
- [ ] Support `Valid() error` that gets called while Unmarshaling
- [ ] Pull in tests from other libraries
- [ ] Encode nil maps and nil structs as empty objects
- [ ] Fallback to `json.{Decode,Encode}` (?)
- [ ] Bundle into Bud
- [ ] Add MarshalJSON support using the writer
- [x] Re-organize the package structure to allow more marshalers (e.g. form)

If you have the itch, I'd very much appreciate your help! I plan to work on this here and there over the next couple months. Your PRs would speed up this timeline significantly.

## Why?

- Nice speed improvement over the reflection-based alternatives.
- Validate while you're unmarshaling.
- Nil maps and structs can be changed to empty objects.
- Code can be re-used for other formats like URL-encoded form data or protobufs.

## Development

```
git clone https://github.com/livebud/marshaler
cd marshaler
go mod tidy
go test ./...
```

## Prior Art

- [megajson](https://github.com/benbjohnson/megajson): Simple, easy-to-understand code. Uses static analysis instead of build-time reflection. No longer in development. No built-in validation. No nested structure support.
- [ffjson](https://github.com/pquerna/ffjson): More features. No updates since 2019. Recent issues unanswered. No built-in validation. Build-time reflection doesn't jibe well with Bud.
- [go-codec](https://github.com/ugorji/go): Actively maintained. More features. Probably a better choice for the time being. No built-in validation. Build-time reflection doesn't jibe well with Bud.

## Thanks

The scanner and writer were originally written by [Ben Johnson](https://twitter.com/benbjohnson) in [megajson](https://github.com/benbjohnson/megajson).

## License

MIT
