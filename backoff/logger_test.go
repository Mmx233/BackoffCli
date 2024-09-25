package backoff

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLogger(t *testing.T) {
	_, ok := interface{}(log.New()).(Logger)
	require.Equal(t, true, ok, "logrus.Logger not meet Logger")
	_, ok = interface{}(log.New().WithContext(context.Background())).(Logger)
	require.Equal(t, true, ok, "logrus.Entry not meet Logger")
}
