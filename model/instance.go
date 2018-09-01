package model

import "time"

//InstanceStatus last known state of an instance
// type InstanceStatus int
//
// var (
// 	//InstanceDied not runnig due to failure
// 	InstanceDied InstanceStatus
// 	//InstanceStopped stopped by request
// 	InstanceStopped = InstanceStatus(10)
// 	//InstanceStarted started
// 	InstanceStarted = InstanceStatus(20)
// )

//NewInstance return a new json instance
func NewInstance(name string) *Instance {
	return &Instance{
		Name:    name,
		Created: time.Now(),
		Status:  InstanceUnavailable,
	}
}

// Instance is a contianer instance
type Instance struct {
	Name    string
	ID      string
	Created time.Time
	Status  InstanceStatus
	// IP      string
	// Port    string
	Url string
}

type InstanceStatus int

const (
	InstanceRunning InstanceStatus = 0 + iota
	InstanceStarting
	InstanceStopped
	InstanceKilled
	InstanceUnavailable
)

func (i InstanceStatus) String() string { return statuses[i] }

var statuses = [...]string{
	"RUNNING",
	"STARTING",
	"STOPPED",
	"KILLED",
	"UNAVAILABLE",
}
