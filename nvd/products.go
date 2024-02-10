package nvd

import "errors"

type Product int

const (
	PRODUCT_UNKNOWN Product = iota
	POSTGRESQL
)

func GetNvdDatabaseType(name string) Product {
	switch name {
	case "postgres":
		return POSTGRESQL
	default:
		return PRODUCT_UNKNOWN
	}
}

func GetNvdDatabaseName(t Product) string {
	switch t {
	case POSTGRESQL:
		return "postgres"
	default:
		return "unknown"
	}
}

func GetNvdCpeForProduct(t Product) (string, error) {
	switch t {
	case POSTGRESQL:
		return "cpe:2.3:a:postgresql:postgresql", nil
	default:
		return "", errors.New("Product does not exist")
	}
}
