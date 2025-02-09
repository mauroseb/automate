package integration_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"path"

	"github.com/chef/automate/api/interservice/data_lifecycle"
	"github.com/chef/automate/components/event-feed-service/pkg/persistence"
	"github.com/chef/automate/components/event-feed-service/pkg/server"
	"github.com/chef/automate/lib/grpc/secureconn"
	"github.com/chef/automate/lib/tls/certs"
	"google.golang.org/grpc"
)

// Suite helps you manipulate various stages of your tests. It provides
// common functionality like initialization and deletion hooks and more.
// If you have some functionality that repeats across multiple tests,
// consider putting it here so that we have it available to the Feed Service
// at a global level.
//
// This struct holds:
// * A FeedServiceClient for making requests against the FeedServiceServer
// * A PurgeClient for making requests against the PurgeServer
// * An Elasticsearch client for ES queries to build up and tear down test cases
//   => Docs: https://godoc.org/gopkg.in/olivere/elastic.v5
type Suite struct {
	feedClient       *server.EventFeedServer
	feedBackend      persistence.FeedStore
	purgeClient      data_lifecycle.PurgeClient
	esClient         *olivere.Client
	indices          []string
	types            []string
	cleanup          func() error
	elasticsearchUrl string
}

// Initialize the test suite
//
// TODO: add check for Elasticsearch connectivity.
// If we can't connect, we'll skip the tests and
// print an error message
func NewSuite(url string) (*Suite, error) {
	s := new(Suite)

	s.elasticsearchUrl = url

	/* cert, err := tls.LoadX509KeyPair("/hab/svc/automate-opensearch/config/root-ca.pem", "/hab/svc/automate-opensearch/config/root-ca-key.pem")
	if err != nil {
		return nil, err
	}
	caCert, err := ioutil.ReadFile("/hab/svc/automate-opensearch/config/root-ca.pem")
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}
	tlsConfig.BuildNameToCertificate() */
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}

	esClient, err := olivere.NewClient(
		olivere.SetURL(s.elasticsearchUrl),
		olivere.SetSniff(false),
		olivere.SetHttpClient(client),
		olivere.SetBasicAuth("admin", "admin"),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "connecting to opensearch (%s)", url)
	}
	s.esClient = esClient
	s.indices = []string{persistence.IndexNameFeeds}

	s.feedBackend = persistence.NewFeedStore(esClient)
	err = s.feedBackend.InitializeStore(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "initializing feed store backend")
	}

	factory, err := secureConnFactoryHab()
	if err != nil {
		return nil, errors.Wrap(err, "loading hab grpc conn factory")
	}

	conn, err := factory.DialContext(
		context.Background(),
		"event-feed-service",
		"localhost:10134",
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "initializing gRPC clients")
	}

	s.feedClient = server.New(s.feedBackend)
	s.purgeClient = data_lifecycle.NewPurgeClient(conn)
	s.cleanup = conn.Close

	return s, nil
}

// Set up global test fixtures
func (s *Suite) GlobalSetup() {
}

// Tear down global test fixtures
func (s *Suite) GlobalTeardown() {
	defer s.cleanup() // nolint errcheck

	// Make sure we clean them all!
	toDelete := s.verifyIndices(s.indices...)
	// if there are no valid Indices, stop processing
	if len(toDelete) == 0 {
		return
	}
	for i, v := range toDelete {
		if v == ".opendistro_security" {
			toDelete = append(toDelete[:i], toDelete[i+1:]...)
			break
		}
	}
	_, err := s.esClient.DeleteIndex(toDelete...).Do(context.Background())
	if err != nil {
		fmt.Printf("Could not 'delete' ES indices: '%s'\nError: %s", s.indices, err)
		os.Exit(3)
	}
}

// DeleteAllDocuments will clean every single document from all ES Indices
//
// You should call this method on every single test as the following example:
// ```
//  func TestGrpcFunc(t *testing.T) {
//    // Here we are ingesting a number of nodes
//    suite.IngestNodes(nodes)
//
//    // Immediately after the ingestion add the hook to clean all documents,
//    // by using `defer` you will ensure that the next test will have clean
//    // data regardless if this test passes or fails
//    defer suite.DeleteAllDocuments()
//  }
// ```
func (s *Suite) DeleteAllDocuments() {
	// ES Query to match all documents
	q := olivere.RawStringQuery("{\"match_all\":{}}")

	// Make sure we clean them all!
	indices := s.indices
	for i, v := range indices {
		if v == ".opendistro_security" {
			indices = append(indices[:i], indices[i+1:]...)
			break
		}
	}
	_, err := s.esClient.DeleteByQuery().
		Index(indices...).
		Type(s.types...).
		Query(q).
		IgnoreUnavailable(true).
		Refresh("true").
		WaitForCompletion(true).
		Do(context.Background())

	if err != nil {
		fmt.Printf("Could not 'clean' ES documents from indices: '%v'\nError: %s", indices, err)
		os.Exit(3)
	}
}

func (s *Suite) RefreshIndices(indices ...string) {
	// Verify that the provided indices exists, if not remove them
	indices = s.verifyIndices(indices...)

	// If there are no valid Indices, stop processing
	if len(indices) == 0 {
		return
	}

	_, err := s.esClient.Refresh(indices...).Do(context.Background())
	if err != nil {
		fmt.Printf("Could not 'refresh' ES documents from indices: '%v'\nError: %s", indices, err)
		os.Exit(3)
	}
}

// verifyIndices receives a list of indices and verifies that they exist.
// If an index doesn't exist, it is removed from the list. Only existing
// indices are returned.
func (s *Suite) verifyIndices(indices ...string) []string {
	var validIndices = make([]string, 0)

	for _, index := range indices {
		if s.indexExists(index) {
			validIndices = append(validIndices, index)
		}
	}

	return validIndices
}

func (s *Suite) indexExists(i string) bool {
	exists, _ := s.esClient.IndexExists(i).Do(context.Background())
	return exists
}

func secureConnFactoryHab() (*secureconn.Factory, error) {
	certs, err := loadCertsHab()
	if err != nil {
		return nil, errors.Wrap(err, "loading event-feed-service TLS certs")
	}

	return secureconn.NewFactory(*certs), nil
}

// uses the certs in a running hab env
func loadCertsHab() (*certs.ServiceCerts, error) {
	dirname := "/hab/svc/event-feed-service/config"
	log.Infof("certs dir is %s", dirname)

	cfg := certs.TLSConfig{
		CertPath:       path.Join(dirname, "service.crt"),
		KeyPath:        path.Join(dirname, "service.key"),
		RootCACertPath: path.Join(dirname, "root_ca.crt"),
	}

	return cfg.ReadCerts()
}
