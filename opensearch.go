package csv2opensearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/opensearch-project/opensearch-go"
)

// Writer indexes batches of documents into OpenSearch.
// Note: It relies on OpenSearch for document ID generation.
type Writer struct {
	client         *opensearch.Client
	host, index    string
	user, password string
}

// WithBasicAuth sets up the credentials for basic authentication to OpenSearch.
func WithBasicAuth(user, password string) func(*Writer) {
	return func(w *Writer) {
		w.user = user
		w.password = password
	}
}

// NewWriter creates a new connection to the OpenSearch cluster.
//
// It accepts a list of options used to further configure the client, e.g. setting the username and password.
// See https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis.
//
// Note: It does NOT work for TLS enabled clusters.
func NewWriter(host string, index string, options ...func(*Writer)) (*Writer, error) {
	w := Writer{host: host, index: index}
	for _, option := range options {
		option(&w)
	}

	cfg := opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Addresses: []string{host},
		Username:  w.user,
		Password:  w.password,
	}

	client, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenSearch client: %v", err)
	}
	w.client = client

	log.Printf("Writing data to index %q", index)

	return &w, nil
}

// Write issues a Bulk request using the `_index` action type for each document.
//
// See https://opensearch.org/docs/1.2/opensearch/rest-api/document-apis/bulk/.
func (w *Writer) Write(_ context.Context, docs []string) error {
	req, err := buildBulkRequest(w.index, docs)
	if err != nil {
		return fmt.Errorf("failed to build bulk request: %v", err)
	}

	res, err := w.client.Bulk(strings.NewReader(req))
	if err != nil {
		return fmt.Errorf("failed to write to index: %v", err)
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("received non-200 status code: %v", res)
	}

	return nil
}

func buildBulkRequest(index string, docs []string) (string, error) {
	b := strings.Builder{}
	for i := range docs {
		if docs[i] == "" { // skip empty rows
			continue
		}
		b.WriteString(fmt.Sprintf("{\"index\": {\"_index\": \"%s\"}}\n", index))
		b.WriteString(docs[i])
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String(), nil
}
