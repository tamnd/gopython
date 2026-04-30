package pycore

import "testing"

func TestPyOSSnprintfFits(t *testing.T) {
	buf := make([]byte, 8)
	n := PyOSSnprintf(buf, "x=%d", 12)
	if n != 4 {
		t.Fatalf("n = %d", n)
	}
	if string(buf[:5]) != "x=12\x00" {
		t.Fatalf("buf = %q", buf)
	}
	if buf[len(buf)-1] != 0 {
		t.Fatal("last byte was not nul")
	}
}

func TestPyOSSnprintfTruncatesAndTerminates(t *testing.T) {
	buf := []byte{'?', '?', '?', '?'}
	n := PyOSSnprintf(buf, "abcdef")
	if n != 6 {
		t.Fatalf("n = %d", n)
	}
	if string(buf) != "abc\x00" {
		t.Fatalf("buf = %q", buf)
	}
}

func TestPyOSSnprintfRequiresBuffer(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	PyOSSnprintf(nil, "x")
}
