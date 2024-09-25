package backoff

import (
	"context"
	log "github.com/sirupsen/logrus"
)

type Logger interface {
	log.FieldLogger
	WithContext(ctx context.Context) *log.Entry
}
