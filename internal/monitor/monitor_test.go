package monitor

import (
	"bytes"
	"log/slog"
	"testing"
	"yanm/internal/network/networkmock"
	"yanm/internal/storage/storagemock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

// TestNewNetwork tests the creation of a Network monitor
func TestNewNetwork(t *testing.T) {
	logs := bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	mockCtrl := gomock.NewController(t)
	storageMock := storagemock.NewMockMetricsStorage(mockCtrl)
	networkMock := networkmock.NewMockSpeedTester(mockCtrl)

	network := NewNetwork(logger, storageMock, networkMock)
	require.NotNil(t, network)

}
