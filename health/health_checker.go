//nolint:wrapcheck
package health

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

type Checker struct {
	checkers []checkerI
}

type checkerI interface {
	check(ctx context.Context) error
}

var ErrInvalidParam = errors.New("param is missing or invalid")

func NewChecker(opts ...func(*Checker) error) (*Checker, error) {
	hc := &Checker{} //nolint:exhaustruct

	for _, o := range opts {
		if err := o(hc); err != nil {
			return nil, err
		}
	}

	return hc, nil
}

func (hc *Checker) Check(ctx context.Context) error {
	const timeout = 30 * time.Second

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for _, c := range hc.checkers {
		if err := c.check(timeoutCtx); err != nil {
			return err
		}
	}

	return nil
}
