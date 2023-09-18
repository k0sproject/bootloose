package tests

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
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

func find(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		switch {
		case err != nil:
			return err
		case info.IsDir():
			return nil
		case strings.HasSuffix(path, "~"):
			return nil
		}
		files = append(files, strings.TrimPrefix(path, dir))
		return nil
	})

	return files, err
}

// test is a end to end test, corresponding to one test-$testname.cmd file.
type test struct {
	testname string   // test name, after variable resolution.
	file     string   // name of the test file (test-*.cmd), without the extentension.
	vars     []string // user-defined variables list of key, value pairs.
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

func (t *test) shouldErrorOut() bool {
	return exists(t.basefile() + ".error")
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
	name string
	args []string
	// should we capture the command output to be tested against the golden
	// output?
	captureOutput bool
	doDefer       bool
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

func (t *test) expandVars(s string) string {
	replacements := copyArray(t.vars)
	replacements = append(replacements,
		"%testOutputDir", t.outputDir(),
		"%testName", t.name(),
		"%image", t.image(),
	)
	replacer := strings.NewReplacer(replacements...)
	return replacer.Replace(s)
}

func (t *test) parseCmd(line string) cmd {
	parts := strings.Split(line, " ")
	goRun := []string{"go", "run", appPath() + "/."}

	// Replace special strings
	for i := range parts {
		parts[i] = t.expandVars(parts[i])
	}

	cmd := cmd{}
	var done bool
	for !done {
		switch parts[0] {
		case "%out":
			cmd.captureOutput = true
			parts = parts[1:]
		case "%defer":
			cmd.doDefer = true
			parts = parts[1:]
		case "footloose":
			parts = append(goRun, parts[1:]...)
		default:
		  done = true
		}
	}

	cmd.name = parts[0]
	cmd.args = parts[1:]
	return cmd

}

func (t *test) run() (string, error) {
	f, err := os.Open(t.file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var capturedOutput strings.Builder

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line[0] == '#' {
			continue
		}
		testCmd := t.parseCmd(line)
		cmd := exec.Command(testCmd.name, testCmd.args...)
		if testCmd.doDefer {
			defer func() { _ = cmd.Run() }()
			continue
		}
		if testCmd.captureOutput {
			output, err := cmd.CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("failed to run command cmd=%s args=%v error:%w output: %s", testCmd.name, testCmd.args, err, string(output))
			}
			capturedOutput.Write(output)
		} else {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return "", err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return capturedOutput.String(), nil
}

func (t *test) goldenOutput() string {
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

func runTest(t *testing.T, test *test) {
	base := test.file
	goldenDir := base + ".golden"
	gotDir := base + ".got"

	if test.shouldSkip() {
		return
	}

	output, err := test.run()

	// 0. Check process exit code.
	if test.shouldErrorOut() {
		_, ok := err.(*exec.ExitError)
		assert.True(t, ok, err.Error())
	} else {
		if err != nil {
			t.Logf("output: %s", output)
		}
		require.NoError(t, err)
	}

	// 1. Compare stdout/err.
	assert.Equal(t, test.goldenOutput(), strings.TrimSpace(string(output)))

	// 2. Compare produced files.
	goldenFiles, _ := find(goldenDir)
	gotFiles, _ := find(gotDir)

	// 2. a) Compare the list of files.
	if !assert.Equal(t, goldenFiles, gotFiles) {
		assert.FailNow(t, "generated files not equivalent; bail")
	}

	// 2. b) Compare file content.
	for i := range goldenFiles {
		golden, err := os.ReadFile(goldenDir + goldenFiles[i])
		require.NoError(t, err)
		got, err := os.ReadFile(gotDir + gotFiles[i])
		require.NoError(t, err)

		assert.Equal(t, string(golden), string(got))
	}
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

			expanded = append(expanded, test{
				testname: testname,
				file:     f,
				vars:     item.combination,
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
