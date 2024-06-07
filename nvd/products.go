package nvd

import "errors"

type Product int

const (
	PRODUCT_UNKNOWN Product = iota
	POSTGRESQL
	MYSQL
	REDIS
	MONGODB
)

func GetNvdProductType(name string) Product {
	switch name {
	case "postgres":
		return POSTGRESQL
	case "mysql":
		return MYSQL
	case "redis":
		return REDIS
	case "mongodb":
		return MONGODB
	default:
		return PRODUCT_UNKNOWN
	}
}

func GetNvdProductName(t Product) string {
	switch t {
	case POSTGRESQL:
		return "postgres"
	case MYSQL:
		return "mysql"
	case REDIS:
		return "redis"
	case MONGODB:
		return "mongodb"
	default:
		return "unknown"
	}
}

func GetNvdCpeForProduct(t Product) (string, error) {
	switch t {
	case POSTGRESQL:
		return "cpe:2.3:a:postgresql:postgresql", nil
	case MYSQL:
		return "cpe:2.3:a:oracle:mysql", nil
	case REDIS:
		return "cpe:2.3:a:redis:redis", nil
	case MONGODB:
		return "cpe:2.3:a:mongodb:mongodb", nil
	default:
		return "", errors.New("Product does not exist")
	}
}
