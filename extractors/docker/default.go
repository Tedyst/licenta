package docker

var defaultIgnoreFileNameIncluding = [...]string{
	"LICENSE",
	".so",
	"sbin/",
	"bin/",
	"lib/",
	"lib64/",
	"lib32/",
	"libx32/",
	"libexec/",
	"share/",
	"include/",
	".asc",
	".css",
	".woff2",
	".crt",
	".pem",
	"__pycache__",
	".git",
	".cache",
	"etc/ssl/openssl.cnf",
}

func GetDefaultIgnoreFileNames() []string {
	return defaultIgnoreFileNameIncluding[:]
}
