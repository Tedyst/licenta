package git

var defaultIgnoreFileNameIncluding = [...]string{
	"__pycache__",
}

func GetDefaultIgnoreFileNames() []string {
	return defaultIgnoreFileNameIncluding[:]
}
