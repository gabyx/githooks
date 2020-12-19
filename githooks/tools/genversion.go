// +build tools

package main

import (
	"bytes"
	"go/format"
	"os"
	"path"
	cm "rycus86/githooks/common"
	"rycus86/githooks/git"
	"text/template"
)

var pkg = "build"
var verFile = "build/version.go"

var versionTpl = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
package {{ .Package }}

import 	(
	"github.com/hashicorp/go-version"
	cm "rycus86/githooks/common"
)

var BuildCommit = "{{ .Commit }}"
var BuildVersion = "{{ .Version }}"
var BuildTag = "{{ .Tag }}"

func GetBuildVersion() *version.Version {
	ver, _ := version.NewVersion(BuildVersion)
	cm.DebugAssert(ver != nil, "Wrong build version")
	return ver
}
`))

func main() {

	gitx := git.Ctx()

	root, err := git.Ctx().Get("rev-parse", "--show-toplevel")
	if err != nil {
		panic(err)
	}

	verFile = path.Join(root, "githooks", verFile)

	commitSHA, err := git.GetCommitSHA(gitx, "HEAD")
	cm.AssertNoErrorPanicF(err, "GetCommitSHA failed.")

	ver, tag, err := git.GetVersion(gitx, "HEAD")
	cm.AssertNoErrorPanicF(err, "GetVersion failed.")

	// Create or overwrite the go file from template
	var buf bytes.Buffer
	err = versionTpl.Execute(&buf, struct {
		Package string
		Version string
		Tag     string
		Commit  string
	}{
		Package: pkg,
		Version: ver.String(),
		Tag:     tag,
		Commit:  commitSHA,
	})
	cm.AssertNoErrorPanicF(err, "Setting template failed.")

	// Format
	src, err := format.Source(buf.Bytes())
	cm.AssertNoErrorPanicF(err, "Formatting template failed.")

	// Write to disk (in the Current Working Directory)
	f, err := os.Create(verFile)
	cm.AssertNoErrorPanicF(err, "Opening template file.")
	defer f.Close()

	_, err = f.Write(src)
	cm.AssertNoErrorPanicF(err, "Opening template file.")
}
