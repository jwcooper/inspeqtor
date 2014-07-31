package inspeqtor

import (
	"inspeqtor/services"
)

/*
 Core Inspeqtor types, interfaces, etc.
*/
const (
	CYCLE_TIME = 15
	ONE_HOUR   = 3600
	SLOTS      = ONE_HOUR / CYCLE_TIME
)

/*
  A service is a logical named entity we wish to monitor, "mysql".
  A logical service maps onto a physical process with a PID.
  PID 0 means the process did not exist during that cycle.
*/
type Service struct {
	Name    string
	PID     services.ProcessId
	Status  services.Status
	Rules   []*Rule
	Metrics interface{}

	// Upon bootup, we scan each init system looking for the service
	// and cache which init system manages it for our lifetime.
	Manager services.InitSystem
}

type Host struct {
	Name    string
	Rules   []*Rule
	Metrics interface{}
}

type Operator uint8

const (
	LT Operator = iota
	GT
)

type RuleStatus uint8

const (
	Undetermined RuleStatus = iota
	Ok
	Tripped
	Recovered
	Unchanged
)

type Rule struct {
	MetricFamily string
	MetricName   string
	Op           Operator
	Threshold    int64
	CycleCount   uint8
	Status       RuleStatus
	Actions      []*Action
}

type Alert struct {
	*Service
	*Rule
}

func (r Rule) Metric() string {
	s := r.MetricFamily
	if r.MetricName != "" {
		s += "(" + r.MetricName + ")"
	}
	return s
}

type Action interface {
	Name() string
	Setup(map[string]string) error
	Trigger(alert *Alert) error
}