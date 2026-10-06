package tests

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// clickhouseImage is the same image the local compose stack uses, so it is
// normally already present in the Docker cache.
const clickhouseImage = "clickhouse/clickhouse-server:24.8-alpine"

// The image ships the "default" user restricted to the container's own
// loopback interfaces, so a host-side client (testcontainers maps the port)
// gets an authentication failure. Setting CLICKHOUSE_USER/CLICKHOUSE_PASSWORD
// makes the entrypoint drop that user and create one reachable from ::/0.
const (
	clickhouseUser     = "pos"
	clickhousePassword = "pospass"
)

// SetupStatsEnv starts ClickHouse (once per suite) and points the global viper
// config at both test containers. The stats services resolve their
// dependencies through viper (database.NewClientWithPrefix and
// clickhouse.NewClient), so this must run before any of them is constructed.
//
// It is deliberately not called from SetupTestSuite: only the stats suites
// need ClickHouse, and every other suite should not pay for its startup.
func (s *BaseTestSuite) SetupStatsEnv() {
	s.startClickHouse()
	s.configureStatsEnv()
}

func (s *BaseTestSuite) startClickHouse() {
	if s.ts.CHContainer == nil {
		ctr, err := testcontainers.GenericContainer(s.Ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        clickhouseImage,
				ExposedPorts: []string{"9000/tcp", "8123/tcp"},
				Env: map[string]string{
					"CLICKHOUSE_USER":     clickhouseUser,
					"CLICKHOUSE_PASSWORD": clickhousePassword,
				},
				WaitingFor: wait.ForAll(
					wait.ForListeningPort("9000/tcp").WithStartupTimeout(90*time.Second),
					wait.ForHTTP("/ping").WithPort("8123/tcp").WithStartupTimeout(90*time.Second),
				),
			},
			Started: true,
		})
		s.Require().NoError(err)
		s.ts.CHContainer = ctr
	}

	host, err := s.ts.CHContainer.Host(s.Ctx)
	s.Require().NoError(err)
	port, err := s.ts.CHContainer.MappedPort(s.Ctx, "9000/tcp")
	s.Require().NoError(err)
	s.ts.CHAddr = fmt.Sprintf("%s:%s", host, port.Port())
}

// configureStatsEnv mirrors the Postgres container settings into the viper keys
// the services read. Production runs five separate PostgreSQL instances, each
// behind its own PgBouncer; the test deliberately collapses all five contexts
// onto the single testcontainer. That exercises the same per-prefix resolution
// path (and the same "all keys must be set" guard) without paying for five
// containers.
func (s *BaseTestSuite) configureStatsEnv() {
	u, err := url.Parse(s.ts.DBURL)
	s.Require().NoError(err)

	host := u.Hostname()
	port := u.Port()
	user := u.User.Username()
	pass, _ := u.User.Password()
	dbname := strings.TrimPrefix(u.Path, "/")

	viper.Set("DB_DRIVER", "pgx")
	viper.Set("DB_USERNAME", user)
	viper.Set("DB_PASSWORD", pass)

	for _, prefix := range []string{"DB_IDENTITY", "DB_MERCHANT", "DB_CATALOG", "DB_SALES", "DB_EMAIL"} {
		viper.Set(prefix+"_DRIVER", "pgx")
		viper.Set(prefix+"_HOST", host)
		viper.Set(prefix+"_PORT", port)
		viper.Set(prefix+"_USERNAME", user)
		viper.Set(prefix+"_PASSWORD", pass)
		viper.Set(prefix+"_NAME", dbname)
	}

	viper.Set("CLICKHOUSE_ADDR", s.ts.CHAddr)
	viper.Set("CLICKHOUSE_DATABASE", "default")
	viper.Set("CLICKHOUSE_USERNAME", clickhouseUser)
	viper.Set("CLICKHOUSE_PASSWORD", clickhousePassword)
}

// ClickHouseAddr returns the native-protocol address of the started container.
func (s *BaseTestSuite) ClickHouseAddr() string {
	return s.ts.CHAddr
}
