package nvd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

const baseNvdCpeUrl = "https://services.nvd.nist.gov/rest/json/cpes/2.0"
const baseNvdCveUrl = "https://services.nvd.nist.gov/rest/json/cves/2.0"

const baseNvdCpeCpeMatchStringQuery = "cpeMatchString"
const baseNvdCpeCpeNameQuery = "cpeName"

const postgresqlCpe = "cpe:2.3:a:postgresql:postgresql"

var semverRegex = regexp.MustCompile(`(0|[1-9]\d*)\.(0|[1-9]\d*)(\.(0|[1-9]\d*))?(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`)

var ErrRateLimit error = errors.New("request rate limit exceeded")

type Product int

const (
	PRODUCT_UNKNOWN Product = iota
	POSTGRESQL
)

func GetNvdDatabaseType(name string) (Product, error) {
	switch name {
	case "postgres":
		return POSTGRESQL, nil
	default:
		return 0, errors.New("Product does not exist")
	}
}

func DownloadCpe(ctx context.Context, product Product) (io.ReadCloser, error) {
	return DownloadCpeNext(ctx, product, 0)
}

func DownloadCpeNext(ctx context.Context, product Product, startIndex int64) (io.ReadCloser, error) {
	slog.DebugContext(ctx, "Downloading CPEs from NVD", slog.String("product", fmt.Sprint(product)), slog.Int64("startIndex", startIndex))
	start := time.Now()

	u, err := url.Parse(baseNvdCpeUrl)
	if err != nil {
		return nil, err
	}
	u.Scheme = "https"
	q := u.Query()
	switch product {
	case POSTGRESQL:
		q.Set(baseNvdCpeCpeMatchStringQuery, postgresqlCpe)
	default:
		return nil, errors.New("Product does not exist")
	}
	q.Set("startIndex", fmt.Sprint(startIndex))
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 429 {
		return nil, ErrRateLimit
	}

	slog.DebugContext(ctx, "Finished downloading CPEs from NVD", slog.String("product", fmt.Sprint(product)), slog.Int64("startIndex", startIndex), slog.Int64("ms", time.Since(start).Milliseconds()))

	return resp.Body, err
}

func DownloadCVEs(ctx context.Context, product Product, cpe string) (io.ReadCloser, error) {
	return DownloadCVEsNext(ctx, product, cpe, 0)
}

func DownloadCVEsNext(ctx context.Context, product Product, cpe string, startIndex int64) (io.ReadCloser, error) {
	slog.DebugContext(ctx, "Downloading CVEs from NVD", slog.String("product", fmt.Sprint(product)), slog.String("cpe", cpe), slog.Int64("startIndex", startIndex))
	start := time.Now()

	u, err := url.Parse(baseNvdCveUrl)
	if err != nil {
		return nil, err
	}
	u.Scheme = "https"
	q := u.Query()
	q.Set(baseNvdCpeCpeNameQuery, cpe)
	q.Set("startIndex", fmt.Sprint(startIndex))
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 429 {
		return nil, ErrRateLimit
	}

	slog.DebugContext(ctx, "Finished downloading CVEs from NVD", slog.String("product", fmt.Sprint(product)), slog.Int64("startIndex", startIndex), slog.Int64("ms", time.Since(start).Milliseconds()))

	return resp.Body, err
}

func extractCpePostgresqlVersion(titles []NvdCpeTitle) (string, error) {
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
	case POSTGRESQL:
		return extractCpePostgresqlVersion(titles)
	default:
		return "", errors.New("Product does not exist")
	}
}

func ParseCpeAPI(ctx context.Context, reader io.Reader) (NvdCpeAPIResult, error) {
	result := NvdCpeAPIResult{}
	err := json.NewDecoder(reader).Decode(&result)
	return result, err
}

func ParseCveAPI(ctx context.Context, reader io.Reader) (NvdCveAPIResult, error) {
	result := NvdCveAPIResult{}
	err := json.NewDecoder(reader).Decode(&result)
	return result, err
}

func GetCveScore(ctx context.Context, cve NvdCveCve) (float64, error) {
	value := 0.0
	for _, metric := range cve.Metrics.CvssMetricV31 {
		if metric.CvssData.BaseScore > value {
			value = metric.CvssData.BaseScore
		}
	}
	for _, metric := range cve.Metrics.CvssMetricV30 {
		if metric.CvssData.BaseScore > value {
			value = metric.CvssData.BaseScore
		}
	}
	for _, metric := range cve.Metrics.CvssMetricV2 {
		if metric.CvssData.BaseScore > value {
			value = metric.CvssData.BaseScore
		}
	}

	if value == 0.0 {
		return 0.0, errors.New("no score found")
	}

	return value, nil
}
