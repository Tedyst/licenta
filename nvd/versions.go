package nvd

import "errors"

func extractCpeSemverVersion(titles []NvdCpeTitle) (string, error) {
	for _, title := range titles {
		extract := semverRegex.FindAllString(title.Title, -1)
		if len(extract) > 0 {
			return extract[0], nil
		}
	}
	return "", errors.New("no version found")
}

func ExtractCpeVersionProduct(product Product, titles []NvdCpeTitle) (string, error) {
	switch product {
	case POSTGRESQL, MYSQL:
		return extractCpeSemverVersion(titles)
	default:
		return "", errors.New("Product does not exist")
	}
}
