# Marshaler (WIP)

`Marshal` and `Unmarshal` generators for Go.

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

This library will generate an UnmarshalJSON function that can take arbitrary JSON and produce `Input`:

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

- [ ] Unmarshal nested structs (get the skipped test to work)
- [ ] Get tests running from a temporary directory
- [ ] Support the json tag
- [ ] Add MarshalJSON support using the writer
- [ ] Resolve nested named types with the parser
- [ ] Support `Valid() error` that gets called while Unmarshaling
- [ ] Pull in tests from other libraries
- [ ] Encode nil maps and nil structs as empty objects
- [ ] Fallback to `json.{Decode,Encode}` (?)
- [ ] Bundle into Bud
- [ ] Re-organize the package structure to allow more marshalers (e.g. form)

If you have the itch, I'd very much appreciate your help! I plan to work on this here and there over the next couple months. Your PRs would speed up this timeline significantly!

## Why?

- Nice speed improvement over the reflection-based alternatives.
- Validate while you're unmarshaling.
- More control: nil maps can be changed to empty objects.

## Prior Art

- [megajson](https://github.com/benbjohnson/megajson): No longer in development. No built-in validation. No nested structure support.
- [ffjson](https://github.com/pquerna/ffjson): More features. No updates since 2019. Recent issues unanswered. No built-in validation. Build-time reflection doesn't jibe well with Bud.
- [go-codec](https://github.com/ugorji/go): Actively maintained. More features. Probably a better choice for the time being. No built-in validation. Build-time reflection doesn't jibe well with Bud.

## Thanks

The scanner and writer were originally written by [Ben Johnson](https://twitter.com/benbjohnson) in [megajson](https://github.com/benbjohnson/megajson).

## License

MIT
