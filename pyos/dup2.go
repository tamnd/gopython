package pyos

// Dup2 ports CPython's fallback dup2() implementation from Python/dup2.c.
//
// On success it returns fd2. On failure it returns -1 and the underlying
// platform error.
func Dup2(fd1, fd2 int) (int, error) {
	return dup2(fd1, fd2)
}
