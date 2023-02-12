package decoder_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/lithammer/dedent"
	"github.com/livebud/codec/generator"
	"github.com/livebud/codec/generator/decoder"
	"github.com/matryer/is"
)

func writeFS(fsys fs.FS, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if err := os.Mkdir(filepath.Join(dir, path), 0755); err != nil {
				if os.IsExist(err) {
					return nil
				}
				return err
			}
		}
		code, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, path), code, 0644)
	})
}

type Test struct {
	Dir      string
	Files    map[string]string
	Selector *generator.Selector
	Input    string
	Expect   string
}

const goMod = `
module app.com
`

const mainGo = `
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	result, err := run()
	if err != nil {
		fmt.Fprint(os.Stdout, err.Error())
		return
	}
	fmt.Fprint(os.Stdout, string(result))
}

func run() ([]byte, error) {
	input := new(Input)
	if err := UnmarshalJSON([]byte(%q), input); err != nil {
		return nil, err
	}
	result, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	return result, nil
}
`

func runTest(t *testing.T, test *Test) {
	t.Helper()
	is := is.New(t)
	if test.Files == nil {
		test.Files = map[string]string{}
	}
	fsys := fstest.MapFS{}
	for name, content := range test.Files {
		fsys[name] = &fstest.MapFile{
			Data: []byte(dedent.Dedent(content)),
		}
	}
	if test.Dir == "" {
		// TODO: support running tests in a tmp dir
		// test.Dir = t.TempDir()
		test.Dir = "_tmp"
	}
	decgen := decoder.New(fsys)
	if test.Selector == nil {
		test.Selector = &generator.Selector{
			Dir:    ".",
			Type:   "Input",
			Target: ".",
		}
	}
	generatedCode, err := decgen.Generate(test.Selector)
	if err != nil {
		is.Equal(test.Expect, err.Error())
	}
	fsys["actual.go"] = &fstest.MapFile{
		Data: generatedCode,
	}
	fsys["main.go"] = &fstest.MapFile{
		Data: []byte(fmt.Sprintf(mainGo, test.Input)),
	}
	// TODO: add this back in
	// if fsys["go.mod"] == nil {
	// 	fsys["go.mod"] = &fstest.MapFile{
	// 		Data: []byte(goMod),
	// 	}
	// }
	is.NoErr(writeFS(fsys, test.Dir))
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = test.Dir
	actual := new(bytes.Buffer)
	cmd.Stdout = actual
	cmd.Stderr = os.Stderr
	is.NoErr(cmd.Run())
	is.Equal(test.Expect, actual.String())
}

func TestString(t *testing.T) {
	runTest(t, &Test{
		Files: map[string]string{
			"input.go": `
				package main
				type Input string
			`,
		},
		Input:  `"hello"`,
		Expect: `"hello"`,
	})
}

func TestEmptyStruct(t *testing.T) {
	runTest(t, &Test{
		Files: map[string]string{
			"input.go": `
				package main
				type Input struct {
				}
			`,
		},
		Input:  `{}`,
		Expect: `{}`,
	})
}

func TestStruct(t *testing.T) {
	runTest(t, &Test{
		Files: map[string]string{
			"input.go": `
				package main
				type Input struct {
					B string
					C int
					D float64
					E bool
					F map[string]string
					G []int
					H *string
				}
			`,
		},
		Input:  `{"B":"foo","C":1,"D":1.1,"E":true,"F":{"foo":"bar"},"G":[1,2,3],"H":"hello"}`,
		Expect: `{"B":"foo","C":1,"D":1.1,"E":true,"F":{"foo":"bar"},"G":[1,2,3],"H":"hello"}`,
	})
}

func TestNestedStruct(t *testing.T) {
	t.Skip("TODO: not finished")
	runTest(t, &Test{
		Dir: "_tmp",
		Files: map[string]string{
			"input.go": `
				package main
				type Input struct {
					B string
					C int
					D float64
					E bool
					F map[string]string
					G []int
					H *string
					I *struct{
						B string
						C int
						D float64
						E bool
						F map[string]string
						G []int
						H *string
					}
				}
			`,
		},
		Input:  `{"B":"foo","C":1,"D":1.1,"E":true,"F":{"foo":"bar"},"G":[1,2,3],"H":"hello",I:{"B":"foo","C":1,"D":1.1,"E":true,"F":{"foo":"bar"},"G":[1,2,3],"H":"hello"}}`,
		Expect: `{"B":"foo","C":1,"D":1.1,"E":true,"F":{"foo":"bar"},"G":[1,2,3],"H":"hello"}`,
	})
}
