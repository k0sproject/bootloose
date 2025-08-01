// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type variables map[string][]string

func (v variables) alternatives(name string) []string {
	return v[name]
}

func (v variables) sortedKeys() []string {
	keys := []string{}
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func copyArray(a []string) []string {
	tmp := make([]string, len(a))
	copy(tmp, a)
	return tmp
}

type expandedItem struct {
	expanded    string
	combination []string
}

func uniqueItems(slice []expandedItem) []expandedItem {
	m := make(map[string]struct{})
	r := []expandedItem{}
	for _, i := range slice {
		if _, ok := m[i.expanded]; ok {
			continue
		}
		m[i.expanded] = struct{}{}
		r = append(r, i)
	}
	return r
}

func fixupSingleCombination(s []expandedItem) []expandedItem {
	// When the expansion result in a single string, it's really not the result of
	// a combination of vars, so clear up the combination field.
	if len(s) == 1 {
		s[0].combination = nil
	}
	return s
}

func (v variables) expand(s string) []expandedItem {
	expanded := []expandedItem{}

	if len(v) == 0 {
		return []expandedItem{
			{expanded: s},
		}
	}

	args := [][]string{}
	for _, k := range v.sortedKeys() {
		alts := v.alternatives(k)

		if len(args) == 0 {
			// Populate args for the first time
			cur := [][]string{}
			for _, alt := range alts {
				cur = append(cur, []string{"%" + k, alt})
			}
			args = cur
			continue
		}

		cur := [][]string{}
		for _, a := range args {
			for _, alt := range alts {
				tmp := copyArray(a)
				cur = append(cur, append(tmp, "%"+k, alt))
			}
		}
		args = cur
	}

	for _, a := range args {
		replacer := strings.NewReplacer(a...)
		expanded = append(expanded, expandedItem{
			expanded:    replacer.Replace(s),
			combination: a,
		})
	}
	return fixupSingleCombination(uniqueItems(expanded))
}

func TestVariableExpansion(t *testing.T) {
	v := make(variables)
	v["foo"] = []string{"foo1", "foo2"}
	v["bar"] = []string{"bar1", "bar2", "bar3"}

	// Test a string expansion
	assert.Equal(t, []expandedItem{
		{"foo1-bar1", []string{"%bar", "bar1", "%foo", "foo1"}},
		{"foo2-bar1", []string{"%bar", "bar1", "%foo", "foo2"}},
		{"foo1-bar2", []string{"%bar", "bar2", "%foo", "foo1"}},
		{"foo2-bar2", []string{"%bar", "bar2", "%foo", "foo2"}},
		{"foo1-bar3", []string{"%bar", "bar3", "%foo", "foo1"}},
		{"foo2-bar3", []string{"%bar", "bar3", "%foo", "foo2"}},
	}, v.expand("%foo-%bar"))

	// When a string doesn't need expansion.
	assert.Equal(t, []expandedItem{
		{"foo", nil},
	}, v.expand("foo"))
}

// test is a end to end test, corresponding to one test-$testname.cmd file.
type test struct {
	testname   string          // test name, after variable resolution.
	file       string          // name of the test file (test-*.cmd), without the extentension.
	vars       []string        // user-defined variables list of key, value pairs.
	captures   []*bytes.Buffer // captured output to %1, %2, etc.
	shouldFail bool
}

type byName []test

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].testname < a[j].testname }

func (t *test) name() string {
	return t.testname
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func (t *test) basefile() string {
	ext := filepath.Ext(t.file)
	return t.file[:len(t.file)-len(ext)]
}

func (t *test) shouldSkip() bool {
	return exists(t.basefile() + ".skip")
}

func (t *test) isLong() bool {
	return exists(t.basefile() + ".long")
}

func (t *test) outputDir() string {
	return t.testname + ".got"
}

type cmd struct {
	name    string
	args    []string
	doDefer bool
	stdout  io.Writer
	stderr  io.Writer
}

func (t *test) image() string {
	parts := strings.Split(t.testname, "-")
	return parts[len(parts)-1]
}

// returns the path to to project root (assuming this file is still in the tests/ directory)
func appPath() string {
	_, self, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(self))
}

func (t *test) replacer() *strings.Replacer {
	replacements := copyArray(t.vars)
	replacements = append(replacements,
		"%testOutputDir", t.outputDir(),
		"%testName", t.name(),
		"%image", t.image(),
	)
	for i := 1; i <= 9; i++ {
		replacements = append(replacements, "%"+strconv.Itoa(i), strings.TrimSpace(t.captures[i].String()))
	}

	return strings.NewReplacer(replacements...)
}

func (t *test) expandVars(s string) string {
	return t.replacer().Replace(s)
}

func (t *test) parseCmd(tt *testing.T, line string, lineno int) (*cmd, error) {
	parts := strings.Split(line, " ")
	goRun := []string{"go", "run", appPath() + "/."}

	cmd := &cmd{}

replaceOuter:
	for {
		switch parts[0] {
		case "%out":
			cmd.stdout = t.captures[0]
			cmd.stderr = t.captures[0]
			parts = parts[1:]
		case "%defer":
			cmd.doDefer = true
			parts = parts[1:]
			cmd.stdout = os.Stdout
			cmd.stderr = os.Stderr
		case "%1", "%2", "%3", "%4", "%5", "%6", "%7", "%8", "%9":
			numInt, err := strconv.Atoi(parts[0][1:])
			if err != nil {
				return nil, fmt.Errorf("invalid capture number %s: %w", parts[0], err)
			}
			cmd.stderr = os.Stderr
			cmd.stdout = t.captures[numInt]
			parts = parts[1:]
		case "%assert":
			if len(parts) < 2 {
				return nil, fmt.Errorf("assert requires at least one argument")
			}
			t.assert(tt, lineno, parts[1:]...)
			return nil, nil
		case "%error":
			t.shouldFail = true
			parts = parts[1:]
		case "bootloose":
			parts = append(goRun, parts[1:]...)
		default:
			break replaceOuter
		}
	}

	replacer := t.replacer()

	// Replace special strings
	for i := range parts {
		parts[i] = replacer.Replace(parts[i])
	}

	cmd.name = parts[0]
	cmd.args = parts[1:]
	return cmd, nil
}

func (t *test) testString(s ...string) string {
	if len(s) == 0 {
		return ""
	}

	if len(s) == 1 && len(s[0]) == 2 && s[0][0] == '%' {
		numInt, err := strconv.Atoi(s[0][1:])
		if err != nil {
			return ""
		}
		return strings.TrimSpace(t.captures[numInt].String())
	}
	return strings.TrimSpace(strings.Join(s, " "))
}

func (t *test) assert(tt *testing.T, lineno int, args ...string) {
	if len(args) < 2 {
		tt.Fatal("assert requires at least two arguments on line", lineno)
	}

	switch args[0] {
	case "equal", "notequal", "contains", "notcontains":
		if len(args) < 3 {
			tt.Fatalf("assert %s requires at least two arguments on line %d", args[0], lineno)
		}

		strA := t.testString(args[1])
		strB := t.testString(args[2:]...)

		switch args[0] {
		case "equal":
			assert.Equalf(tt, strA, strB, "assert equal at %s:%d failed: %s != %s", t.file, lineno, strA, strB)
		case "notequal":
			assert.NotEqualf(tt, strA, strB, "assert notequal at %s:%d failed: %s == %s", t.file, lineno, strA, strB)
		case "contains":
			assert.Containsf(tt, strA, strB, "assert contains at %s:%d failed: %s does not contain %s", t.file, lineno, strA, strB)
		case "notcontains":
			assert.NotContainsf(tt, strA, strB, "assert notcontains at %s:%d failed: %s contains %s", t.file, lineno, strA, strB)
		}
	case "empty", "notempty":
		if len(args) != 2 {
			tt.Fatalf("assert %s requires exactly one argument", args[0])
		}
		switch args[0] {
		case "empty":
			assert.Emptyf(tt, t.testString(args[1]), "assert empty at %s:%d failed: %s is not empty", t.file, lineno, t.testString(args[1]))
		case "notempty":
			assert.NotEmptyf(tt, t.testString(args[1]), "assert notempty at %s:%d failed: test string is empty", t.file, lineno)
		}
	}
}

func (t *test) run(tt *testing.T) error {
	f, err := os.Open(t.file)
	if err != nil {
		return fmt.Errorf("failed to open command file: %w", err)
	}
	defer func() {
		assert.NoError(tt, f.Close(), "failed to close command file")
	}()

	scanner := bufio.NewScanner(f)
	lineno := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineno++
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		testCmd, err := t.parseCmd(tt, line, lineno)
		if err != nil {
			return fmt.Errorf("failed to parse command: %w", err)
		}
		if testCmd == nil {
			continue
		}
		cmd := exec.Command(testCmd.name, testCmd.args...)
		stdoutBuf := bytes.Buffer{}
		stderrBuf := bytes.Buffer{}
		if testCmd.stdout == nil {
			testCmd.stdout = &stdoutBuf
		} else {
			testCmd.stdout = io.MultiWriter(testCmd.stdout, &stdoutBuf)
		}
		if testCmd.stderr == nil {
			testCmd.stderr = &stderrBuf
		} else {
			testCmd.stderr = io.MultiWriter(testCmd.stderr, &stderrBuf)
		}

		if os.Getenv("DEBUG") != "" {
			tt.Log("Running", testCmd.name, strings.Join(testCmd.args, " "))
			testCmd.stdout = io.MultiWriter(testCmd.stdout, os.Stdout)
			testCmd.stderr = io.MultiWriter(testCmd.stderr, os.Stderr)
		}

		cmd.Stdout = testCmd.stdout
		cmd.Stderr = testCmd.stderr

		if testCmd.doDefer {
			defer func() {
				if os.Getenv("DEBUG") == "" {
					cmd.Stdout = io.Discard
					cmd.Stderr = io.Discard
				} else {
					tt.Log("Running deferred", testCmd.name, strings.Join(testCmd.args, " "))
				}
				_ = cmd.Run()
			}()
			continue
		}
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%w: line %d of %s (`%s %s`) stdout: `%s` stderr: `%s`", err, lineno, t.file, testCmd.name, strings.Join(testCmd.args, " "), stdoutBuf.String(), stderrBuf.String())
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read command file: %w", err)
	}

	return nil
}

func (t *test) expectedOutput() string {
	// testname.golden.output takes precedence.
	golden, err := os.ReadFile(t.testname + ".golden.output")
	if err == nil {
		return strings.TrimSpace(t.expandVars(string(golden)))
	}

	// Expand a generic golden output.
	baseFilename := t.file[:len(t.file)-len(".cmd")]
	data, err := os.ReadFile(baseFilename + ".golden.output")
	if err != nil {
		// not having any golden output isn't an error, it just means the test
		// shouldn't output anything.
		return ""
	}

	return strings.TrimSpace(t.expandVars(string(data)))
}

func (t *test) output() string {
	return strings.TrimSpace(t.captures[0].String())
}

func runTest(t *testing.T, test *test) {
	if test.shouldSkip() {
		return
	}

	err := test.run(t)

	if test.shouldFail {
		require.Error(t, err)
		targetErr := new(exec.ExitError)
		assert.ErrorAs(t, err, &targetErr)
	} else {
		require.NoError(t, err)
	}

	assert.Equal(t, test.expectedOutput(), test.output(), "output does not match expected output")
}

func listTests(t *testing.T, vars variables) []test {
	files, err := filepath.Glob("test-*.cmd")
	require.NoError(t, err)

	// expand variables in file names.
	expanded := []test{}
	for _, f := range files {
		items := vars.expand(f)
		for _, item := range items {
			ext := filepath.Ext(item.expanded)
			testname := item.expanded[:len(item.expanded)-len(ext)]

			var shouldRun bool

			for _, img := range vars["image"] {
				if strings.Contains(testname, img) {
					shouldRun = true
					break
				}
			}

			if !shouldRun {
				continue
			}

			captures := make([]*bytes.Buffer, 10)
			for i := range captures {
				captures[i] = &bytes.Buffer{}
			}
			expanded = append(expanded, test{
				testname: testname,
				file:     f,
				vars:     item.combination,
				captures: captures,
			})
		}
	}

	sort.Sort(byName(expanded))
	return expanded
}

var imageFlag = flag.String("image", "", "Docker image or comma separated images to use for testing")

func TestEndToEnd(t *testing.T) {
	images := strings.Split(*imageFlag, ",")
	if len(images) == 0 {
		t.Fatal("No images specified")
	}
	vars := variables{"image": images}

	tests := listTests(t, vars)

	if len(tests) == 0 {
		t.Fatal("No tests found")
	}

	for _, test := range tests {
		t.Run(test.name(), func(t *testing.T) {
			if test.isLong() && testing.Short() {
				t.Skip("Skipping long running test in short mode")
			}
			runTest(t, &test)
		})
	}
}
