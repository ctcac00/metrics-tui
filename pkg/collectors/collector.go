package collectors

import "context"

// Collector defines the interface for all metric collectors
type Collector interface {
	// Name returns the name of the collector
	Name() string

	// Collect gathers metrics and returns the data
	// The returned data can be of any type, specific to each collector
	Collect(ctx context.Context) (interface{}, error)

	// Interval returns the recommended update interval for this collector
	Interval() uint // in seconds
}
