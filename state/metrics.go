package state

import (
	"github.com/go-kit/kit/metrics"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this
	// package.
	MetricsSubsystem = "state"
)

//go:generate go run ../scripts/metricsgen -struct=Metrics

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// Time between BeginBlock and EndBlock in ms.
	BlockProcessingTime metrics.Gauge

	SaveABCIResponse metrics.Gauge

	UpdateState metrics.Gauge

	CommitState metrics.Gauge
	SaveState   metrics.Gauge

	// ConsensusParamUpdates is the total number of times the application has
	// udated the consensus params since process start.
	ConsensusParamUpdates metrics.Counter

	// ValidatorSetUpdates is the total number of times the application has
	// udated the validator set since process start.
	ValidatorSetUpdates metrics.Counter
}
