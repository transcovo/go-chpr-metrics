/*
Package tests contains test utilities
*/
package tests

import (
	"io/ioutil"
	"os"

	"github.com/transcovo/go-chpr-logger"
)

/*
Util function to cleanly reset the std out for sequential tests
*/
func setStdout(out *os.File) {
	os.Stdout = out
	logger.ReloadConfiguration()
}

/*
CaptureStdout captures output generated on stdout by the execution of a task.

Use only for tests, it hijacks os.Stdout and is not thread-safe.
*/
func CaptureStdout(task func()) string {
	in, out, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	os.Stdout.Sync()
	oldStdout := os.Stdout

	setStdout(out)

	defer setStdout(oldStdout)

	task()

	out.Close()

	stdout, err := ioutil.ReadAll(in)
	if err != nil {
		panic(err)
	}
	return string(stdout)
}

/*
Util function to store the current value of an Env var and return the function to reset it
*/
func setAndUnsetVarsHelper(name, newValue string) func() {
	initValue := os.Getenv(name)
	os.Setenv(name, newValue)
	return func() { os.Setenv(name, initValue) }
}

/*
WithEnvVars provides a modified env for execution of a task and reset the env afterwards at its precedent state
*/
func WithEnvVars(newVars map[string]string, task func()) {
	for env, newValue := range newVars {
		resetEnv := setAndUnsetVarsHelper(env, newValue)
		defer resetEnv()
	}
	task()
}
