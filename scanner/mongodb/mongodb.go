package mongodb

import (
	"context"
	"errors"

	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/nvd"
	"github.com/tedyst/licenta/scanner"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongodbScanResult struct {
	severity scanner.Severity
	message  string
	detail   string
}

func (result *mongodbScanResult) Severity() scanner.Severity {
	return result.severity
}

func (result *mongodbScanResult) Detail() string {
	return result.detail
}

type mongodbScanner struct {
	db *mongo.Client
}

func (sc *mongodbScanner) GetScannerName() string {
	return "mongodb"
}
func (sc *mongodbScanner) GetScannerID() int32 {
	return models.SCAN_MONGODB
}
func GetScannerID() int32 {
	return models.SCAN_MONGODB
}
func (sc *mongodbScanner) GetNvdProductType() nvd.Product {
	return nvd.MONGODB
}
func (sc *mongodbScanner) ShouldNotBePublic() bool {
	return true
}
func (sc *mongodbScanner) Ping(ctx context.Context) error {
	return sc.db.Ping(ctx, nil)
}
func (sc *mongodbScanner) CheckPermissions(ctx context.Context) error {
	return nil
}
func (sc *mongodbScanner) GetVersion(ctx context.Context) (string, error) {
	var result bson.M
	err := sc.db.Database("admin").RunCommand(ctx, bson.D{{Key: "buildInfo", Value: 1}}).Decode(&result)
	if err != nil {
		return "", err
	}

	version, ok := result["version"].(string)
	if !ok {
		return "", errors.New("Could not find version")
	}

	return version, nil
}

func NewScanner(ctx context.Context, db *mongo.Client) (scanner.Scanner, error) {
	sc := &mongodbScanner{
		db: db,
	}

	return sc, nil
}
