package redis

import (
	"bufio"
	"context"
	"errors"
	"strings"

	r "github.com/redis/go-redis/v9"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/nvd"
	"github.com/tedyst/licenta/scanner"
)

type redisScanResult struct {
	severity scanner.Severity
	message  string
	detail   string
}

func (result *redisScanResult) Severity() scanner.Severity {
	return result.severity
}

func (result *redisScanResult) Detail() string {
	return result.detail
}

type redisScanner struct {
	db *r.Client
}

func (sc *redisScanner) GetScannerName() string {
	return "redis"
}
func (sc *redisScanner) GetScannerID() int32 {
	return models.SCAN_REDIS
}
func GetScannerID() int32 {
	return models.SCAN_REDIS
}
func (sc *redisScanner) GetNvdProductType() nvd.Product {
	return nvd.REDIS
}
func (sc *redisScanner) ShouldNotBePublic() bool {
	return true
}
func (sc *redisScanner) Ping(ctx context.Context) error {
	return sc.db.Ping(ctx).Err()
}
func (sc *redisScanner) CheckPermissions(ctx context.Context) error {
	return nil
}
func (sc *redisScanner) GetVersion(ctx context.Context) (string, error) {
	version := sc.db.Do(ctx, "INFO", "SERVER").String()

	rd := bufio.NewReader(strings.NewReader(version))
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			break
		}
		if strings.HasPrefix(line, "redis_version:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "redis_version:")), nil
		}
	}

	return "", errors.New("Could not find version")
}

func NewScanner(ctx context.Context, db *r.Client) (scanner.Scanner, error) {
	sc := &redisScanner{
		db: db,
	}

	return sc, nil
}
