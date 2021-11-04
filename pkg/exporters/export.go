package exporters

import (
	"context"
	"fmt"
	kafkainstanceapi "github.com/redhat-developer/app-services-sdk-go/kafkainstance/apiv1internal"
	kafkainstanceclient "github.com/redhat-developer/app-services-sdk-go/kafkainstance/apiv1internal/client"
	"golang.org/x/oauth2/clientcredentials"
	"gopkg.in/errgo.v2/errors"
	"io"
	"net/http"
	"os"
	"strings"
)

type Format int

const (
	Prometheus Format = 0
)

func Export(clientId string, clientSecret string, tokenURL string, bootstrapServers []string, format Format, serve bool, host string, port int) error {
	if (serve) {
		fmt.Printf("Listening on http://%s:%d/data\n", host, port)
		dataHandler := func(w http.ResponseWriter, req *http.Request) {
			err := run(clientId, clientSecret, tokenURL, bootstrapServers, format, w)
			if err != nil {
				http.Error(w, err.Error(), 500)
			}
		}

		healthHandler := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("ok"))
		}

		http.HandleFunc("/data", dataHandler)
		http.HandleFunc("/health", healthHandler)

		return http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
	}
	return run(clientId, clientSecret, tokenURL, bootstrapServers, format, os.Stdout)
}

func run(clientId string, clientSecret string, tokenURL string, bootstrapServers []string, format Format, output io.Writer) error {
	data := make(map[string][]kafkainstanceclient.ConsumerGroup)

	for _, bootstrapServer := range bootstrapServers {
		apiURL := strings.TrimSuffix(bootstrapServer, ":443")
		apiURL = fmt.Sprintf("https://admin-server-%s/rest", apiURL)
		entry, err := getData(clientId, clientSecret, tokenURL, apiURL)
		if err != nil {
			return errors.Wrap(err)
		}
		data[bootstrapServer] = entry
	}
	switch format {
	case Prometheus:
		err := AsPrometheus(data, output)
		if err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}

func getData(clientId string, clientSecret string, tokenURL string, apiURL string) ([]kafkainstanceclient.ConsumerGroup, error) {
	fmt.Printf("Connecting to %s\n", apiURL)
	ctx := context.Background()
	oauth2Config := clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
	}
	tc := oauth2Config.Client(ctx)

	apiClient := kafkainstanceapi.NewAPIClient(&kafkainstanceapi.Config{
		HTTPClient: tc,
		Debug:      false,
		BaseURL:    apiURL,
	})

	res, _, err := apiClient.GroupsApi.GetConsumerGroups(ctx).Execute()
	if err != nil {
		return nil, err
	}
	return res.GetItems(), nil
}
