package tests

import (
	"context"
	"fmt"
	"github.com/ory/dockertest/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("mongo", "7.0.0", []string{})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	err = pool.Retry(func() error {
		var err error
		db, err = mongo.Connect(
			context.Background(),
			options.Client().ApplyURI(
				fmt.Sprintf("mongodb://localhost:%s/", resource.GetPort("27017/tcp")),
			),
		)
		if err != nil {
			return err
		}
		return db.Ping(context.Background(), nil)
	})
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

//const Mongo
//
//type testClient struct {
//	client  *http.Client
//	baseURL string
//}
//
//func setup() *testClient {
//	conn, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoConn))
//	a := app.New()
//	server := httpserver.New(slog.Default(), ":18080", gin.ReleaseMode. )
//	testServer := httptest.NewServer(server.Handler)
//
//	return &testClient{
//		client:  testServer.Client(),
//		baseURL: testServer.URL,
//	}
//}

//type ExampleTestSuite struct {
//	suite.Suite
//	VariableThatShouldStartAtFive int
//	Client                        *http.Client
//	BaseURL                       string
//	Method                        string
//}
//
//func (suite *ExampleTestSuite) SetupTest() {
//	suite.VariableThatShouldStartAtFive = 5
//}
//
//func (suite *ExampleTestSuite) TestExample() {
//	assert.Equal(suite.T(), 5, suite.VariableThatShouldStartAtFive)
//}
//
//func TestExampleTestSuite(t *testing.T) {
//	suite.Run(t, new(ExampleTestSuite))
//}
