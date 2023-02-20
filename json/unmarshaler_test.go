package json_test

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	_ "github.com/hexops/valast"

	"github.com/livebud/marshaler/internal/imports"

	"github.com/lithammer/dedent"
	"github.com/livebud/marshaler/internal/finder"
	"github.com/livebud/marshaler/json"
	"github.com/matryer/is"
	"golang.org/x/mod/modfile"
)

func findGoMod(absdir string) (abs string, err error) {
	path := filepath.Join(absdir, "go.mod")
	// Check if this path exists, otherwise recursively traverse towards root
	if _, err = os.Stat(path); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		nextDir := filepath.Dir(absdir)
		if nextDir == absdir {
			return "", fs.ErrNotExist
		}
		return findGoMod(filepath.Dir(absdir))
	}
	return filepath.EvalSymlinks(absdir)
}

func redent(s string) string {
	return strings.TrimSpace(dedent.Dedent(s)) + "\n"
}

func goRun(ctx context.Context, cacheDir, appDir string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "run", "-mod", "mod", ".")
	stdout := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = appDir
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return stdout.String(), nil
}

type Test struct {
	Dir    string
	Files  map[string]string
	Input  string
	Expect string
}

const goMod = `
module app.com

require (
	github.com/livebud/marshaler v0.0.0
)
`

type State struct {
	Imports   []*imports.Import
	Input     string
	Unmarshal string
}

var mainGen = template.Must(template.New("main.go").Parse(`
package main

{{- if $.Imports }}

import (
	{{- range $import := $.Imports }}
	{{$import.Name}} "{{$import.Path}}"
	{{- end }}
)
{{- end }}

func main() {
	var in Input
	if err := UnmarshalJSON([]byte(` + "`" + `{{ .Input }}` + "`" + `), &in); err != nil {
		fmt.Fprintf(os.Stdout, "%s\n", err)
		return
	};
	actual, err := json.Marshal(in)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%s\n", err)
		return
	}
	fmt.Fprintf(os.Stdout, "%s", string(actual))
}

{{ $.Unmarshal }}
`))

func runTest(t testing.TB, test Test) {
	t.Helper()
	is := is.New(t)
	ctx := context.Background()
	if test.Dir == "" {
		test.Dir = t.TempDir()
	}
	if test.Files == nil {
		test.Files = map[string]string{}
	}
	if test.Files["go.mod"] == "" {
		test.Files["go.mod"] = goMod
	}
	// Parse the go.mod file
	modFile, err := modfile.Parse("go.mod", []byte(test.Files["go.mod"]), nil)
	is.NoErr(err)
	// Replace the marshaler dependency with the local version
	absDir, err := filepath.Abs(".")
	is.NoErr(err)
	myModDir, err := findGoMod(absDir)
	is.NoErr(err)
	is.NoErr(modFile.AddReplace("github.com/livebud/marshaler", "", myModDir, ""))
	test.Files["go.mod"] = string(modfile.Format(modFile.Syntax))
	// Write the test files out
	for path, code := range test.Files {
		fullPath := filepath.Join(test.Dir, path)
		is.NoErr(os.MkdirAll(filepath.Dir(fullPath), 0755))
		is.NoErr(os.WriteFile(fullPath, []byte(redent(code)), 0644))
	}
	// Setup the marshaler
	finder := finder.New(test.Dir)
	imports := imports.Imports{}
	// Setup the unmarshaler
	unmarshaler := &json.Unmarshaler{
		TargetPath: modFile.Module.Mod.Path,
		Find:       finder.Find,
		Import:     imports.Import,
	}
	// Generate the unmarshaler
	unmarshal, err := unmarshaler.Generate("app.com", "Input")
	if err != nil {
		is.Equal(err.Error(), test.Expect)
		return
	}
	// Add main.go's imports
	_, err = imports.Import("fmt")
	is.NoErr(err)
	_, err = imports.Import("os")
	is.NoErr(err)
	_, err = imports.Import("encoding/json")
	is.NoErr(err)
	// Generate the main.go file
	mainGo := new(bytes.Buffer)
	is.NoErr(mainGen.Execute(mainGo, &State{
		Imports:   imports,
		Unmarshal: string(unmarshal),
		Input:     test.Input,
	}))
	// fmt.Println(mainGo.String())
	// Write the main.go file out
	mainPath := filepath.Join(test.Dir, "main.go")
	is.NoErr(os.WriteFile(mainPath, []byte(redent(mainGo.String())), 0644))
	// Run the main.go file
	stdout, err := goRun(ctx, t.TempDir(), test.Dir)
	is.NoErr(err)
	is.Equal(stdout, test.Expect)
}

func TestString(t *testing.T) {
	runTest(t, Test{
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
	runTest(t, Test{
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
	runTest(t, Test{
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
	runTest(t, Test{
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
