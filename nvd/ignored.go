package nvd

var ignoredCVEs = []string{
	"CVE-2009-2943",
	"CVE-2010-3781",
}

func IsCVEIgnored(cveID string) bool {
	for _, ignored := range ignoredCVEs {
		if ignored == cveID {
			return true
		}
	}
	return false
}
