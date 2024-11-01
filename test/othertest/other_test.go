package othertest

import "testing"

func TestString(t *testing.T) {
	str := "hello world\n\r"
	t.Log(len(str))
	t.Log(len([]byte(str)))
}
