package health

import (
	"context"
	"database/sql"
	"strings"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type postgresChecker struct {
	dsn string
}

func newPostgresChecker(dsn string) (*postgresChecker, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.Wrap(ErrInvalidParam, "dsn")
	}

	return &postgresChecker{
		dsn: dsn,
	}, nil
}

func (c *postgresChecker) check(ctx context.Context) error {
	conn, err := sql.Open("postgres", c.dsn)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err := conn.PingContext(ctx); err != nil {
		return errors.Wrap(err, "postgres: ping")
	}

	return nil
}
