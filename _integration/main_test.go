package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type spec struct {
	Command    string `yaml:"cmd"`
	Output     string `yaml:"output"`
	ShouldFail bool   `yaml:"should_fail"`
	Generate   bool   `yaml:"generate"`
}

func TestE2E(t *testing.T) {
	matches, err := filepath.Glob("*/test.golden")
	require.NoError(t, err, "failed to find test specs")

	for _, match := range matches {
		testName := strings.Split(match, "/")[0]
		runner := newRunner(t, testName)
		runner.Run()
	}
}

type testRunner struct {
	t        *testing.T
	testName string
	testSpec spec
}

func newRunner(t *testing.T, name string) *testRunner {
	return &testRunner{
		t:        t,
		testName: name,
	}
}

func (r *testRunner) Run() {
	r.t.Run(r.testName, func(t *testing.T) {
		r.t = t

		r.loadSpec()
		t.Logf("test spec: \n%+v\n", r.testSpec)

		output := r.runCommand()
		t.Logf("test output: \n%+v\n", output)

		r.verifyOutput(output)
	})
}

func (r *testRunner) loadSpec() {
	// read test spec
	specData, err := ioutil.ReadFile(fmt.Sprintf("%s/test.golden", r.testName))
	require.NoError(r.t, err, "failed to read test spec file")
	r.t.Logf("spec data: |\n%s", specData)

	var spec spec
	err = yaml.Unmarshal(specData, &spec)
	require.NoError(r.t, err, "failed to parse test spec")

	r.testSpec = spec
}

func (r *testRunner) runCommand() string {
	if r.testSpec.Generate {
		src := "./" + r.testName + "/src/..."
		gen := exec.Command("go", "generate", src)
		out, err := gen.CombinedOutput()
		require.NoError(r.t, err, "%v exited with %s\nfull output:\n%s", gen.Args, err, string(out))
		if err == nil {
			r.t.Logf("%v: success", gen.Args)
		}
	}
	args := strings.Split(r.testSpec.Command, " ")
	first, rest := args[0], args[1:]

	cmd := exec.Command(first, rest...)
	var output, stderr bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &stderr
	cmd.Dir = r.testName

	err := cmd.Run()
	if r.testSpec.ShouldFail {
		require.Error(r.t, err, "command must fail: '%s %s'\nout:\n%s\nerr:\n%s", first, strings.Join(rest, " "), output.String(), stderr.String())
	} else {
		require.NoError(r.t, err, "command failed: '%s %s'\nout:\n%s\nerr:\n%s", first, strings.Join(rest, " "), output.String(), stderr.String())
	}
	return output.String()
}

func (r *testRunner) verifyOutput(output string) {
	expected := strings.Split(r.testSpec.Output, "\n")
	actual := strings.Split(output, "\n")

	sort.Strings(expected)
	sort.Strings(actual)

	assert.Equal(r.t, expected, actual)
}
