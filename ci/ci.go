package ci

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"errors"

	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/scanner"
)

func waitForScan(ctx context.Context, client generated.ClientWithResponsesInterface, scan *generated.Scan) (int, error) {
	slog.InfoContext(ctx, "Waiting for scan to finish", "scan", scan.Id)

	for {
		response, err := client.GetScanIdWithResponse(ctx, int64(scan.Id))
		if err != nil {
			return 0, fmt.Errorf("cannot get scan: %w", err)
		}

		slog.DebugContext(ctx, "Received scan status from server", "status_code", response.StatusCode(), "scan", scan.Id)

		switch response.StatusCode() {
		case http.StatusOK:
			slog.DebugContext(ctx, "Received scan status from server", "status", response.JSON200.Scan.Status, "severity", response.JSON200.Scan.MaximumSeverity, "error", response.JSON200.Scan.Error, "scan", scan.Id)
			if response.JSON200.Scan.Error != "" {
				return 0, errors.New("received error from remote server: " + response.JSON200.Scan.Error)
			}
			if response.JSON200.Scan.Status == int(models.SCAN_FINISHED) {
				createdAt, err := time.Parse(time.RFC3339, response.JSON200.Scan.CreatedAt)
				if err != nil {
					return 0, err
				}
				endedAt, err := time.Parse(time.RFC3339, response.JSON200.Scan.EndedAt)
				if err != nil {
					return 0, err
				}

				for _, result := range response.JSON200.Results {
					switch result.Severity {
					case int(scanner.SEVERITY_WARNING):
						slog.InfoContext(ctx, "Found problem", "scan", scan.Id, "severity", result.Severity, "title", result.Message, "source", result.ScanSource)
					case int(scanner.SEVERITY_MEDIUM):
						slog.WarnContext(ctx, "Found problem", "scan", scan.Id, "severity", result.Severity, "title", result.Message, "source", result.ScanSource)
					case int(scanner.SEVERITY_HIGH):
						slog.ErrorContext(ctx, "Found problem", "scan", scan.Id, "severity", result.Severity, "title", result.Message, "source", result.ScanSource)
					}
				}

				slog.InfoContext(ctx, "Scan finished", "scan", scan.Id, "time", fmt.Sprint(endedAt.Sub(createdAt).Milliseconds())+"ms")
				return response.JSON200.Scan.MaximumSeverity, nil
			}
		default:
			body := response.Body
			return 0, errors.New("received unknown status code from remote server: " + string(body))
		}

		time.Sleep(2 * time.Second)
		slog.InfoContext(ctx, "Scan not finished yet, waiting 2 seconds...")
	}
}

func ProjectRunAndWaitResults(ctx context.Context, client generated.ClientWithResponsesInterface, projectID int) (int, error) {
	slog.InfoContext(ctx, "Starting project run", "project", projectID)

	response, err := client.PostProjectIdRunWithResponse(ctx, int64(projectID))
	if err != nil {
		return 0, err
	}

	slog.InfoContext(ctx, "Submitted project run request", "project", projectID)

	maximumSeverity := 0

	switch response.StatusCode() {
	case http.StatusOK:
		if !response.JSON200.Success {
			return 0, errors.New("success is not false")
		}
		for _, scan := range response.JSON200.ScanGroup.Scans {
			severity, err := waitForScan(ctx, client, &scan)
			if err != nil {
				return maximumSeverity, err
			}

			if severity > maximumSeverity {
				maximumSeverity = severity
			}
		}
	default:
		return maximumSeverity, errors.New("unknown error")
	}

	return maximumSeverity, nil
}
