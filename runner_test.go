package _exec

import (
	"testing"
)

func TestRunner(t *testing.T) {
	t.Log(NewRunner().AddCommand("echo 'Hello world!'").CombinedOutput())
	NewRunner().AddCommand("echo 'Hello world!'").PipeOutput()
	t.Log(NewRunner().SetFiles("./test.sh").CombinedOutput())
	NewRunner().SetFiles("./test2.sh").PipeOutput()
}
