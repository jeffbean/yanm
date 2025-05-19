package monitor

import "time"

type options struct {
	pingInterval         time.Duration
	networkInterval      time.Duration
	pingTriggerThreshold time.Duration
}

type Option interface {
	apply(*options)
}

type pingIntervalOption struct {
	interval time.Duration
}

func (o *pingIntervalOption) apply(opts *options) {
	opts.pingInterval = o.interval
}

func WithPingInterval(interval time.Duration) Option {
	return &pingIntervalOption{interval}
}

type networkIntervalOption struct {
	interval time.Duration
}

func (o *networkIntervalOption) apply(opts *options) {
	opts.networkInterval = o.interval
}

func WithNetworkInterval(interval time.Duration) Option {
	return &networkIntervalOption{interval}
}

type pingTriggerThresholdOption struct {
	threshold time.Duration
}

func (o *pingTriggerThresholdOption) apply(opts *options) {
	opts.pingTriggerThreshold = o.threshold
}

func WithPingTriggerThreshold(threshold time.Duration) Option {
	return &pingTriggerThresholdOption{threshold}
}
