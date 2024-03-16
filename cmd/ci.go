package cmd

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"github.com/deepmap/oapi-codegen/v2/pkg/securityprovider"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/ci"
)

const v1Endpoint = "/api/v1"

type csrfClient struct {
	httpClient *http.Client
	csrfToken  string
}

func (c *csrfClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-CSRF-Token", c.csrfToken)
	return c.httpClient.Do(req)
}

func initHttpCsrfClient() (*csrfClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating cookie jar: %w", err)
	}
	httpClient := csrfClient{
		httpClient: &http.Client{
			Jar: jar,
		},
		csrfToken: "",
	}

	url, err := url.Parse(viper.GetString("api") + v1Endpoint)
	if err != nil {
		return nil, fmt.Errorf("error parsing url: %w", err)
	}
	optionsRequest, err := httpClient.httpClient.Do(&http.Request{
		Method: http.MethodOptions,
		URL:    url,
	})
	if err != nil {
		return nil, fmt.Errorf("error doing options request: %w", err)
	}

	httpClient.csrfToken = optionsRequest.Header.Get("X-CSRF-Token")

	return &httpClient, nil
}

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Signal the Server that a build should be started and wait for it to finish",
	Long:  `This command connects to the API server and listens to tasks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKeyProvider, err := securityprovider.NewSecurityProviderApiKey("header", "X-Worker-Token", viper.GetString("worker-token"))
		if err != nil {
			return fmt.Errorf("error creating security provider: %w", err)
		}

		httpClient, err := initHttpCsrfClient()
		if err != nil {
			return fmt.Errorf("error creating http client: %w", err)
		}

		client, err := generated.NewClientWithResponses(viper.GetString("api")+"/api/v1", generated.WithRequestEditorFn(apiKeyProvider.Intercept), generated.WithHTTPClient(httpClient))
		if err != nil {
			return fmt.Errorf("error creating client: %w", err)
		}

		severity, err := ci.ProjectRunAndWaitResults(cmd.Context(), client, viper.GetInt("project"))
		if err != nil {
			return fmt.Errorf("error running project: %w", err)
		}

		if severity > viper.GetInt("severity") {
			slog.Error("Severity is higher than allowed, failing build", "severity", severity, "allowed", viper.GetInt("severity"))
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	ciCmd.Flags().String("api", "http://localhost:5000", "API Server URL")
	ciCmd.Flags().Int("project", 0, "The project ID to scan")
	if err := ciCmd.MarkFlagRequired("project"); err != nil {
		panic(err)
	}
	ciCmd.Flags().String("worker-token", "", "Worker token")
	if err := ciCmd.MarkFlagRequired("worker-token"); err != nil {
		panic(err)
	}
	ciCmd.Flags().Int("severity", 2, "Minimum severity to fail the build. 0 - low, 1 - medium, 2 - high")

	rootCmd.AddCommand(ciCmd)
}
