package mikrotik

import (
	"context"
)

type Scheduler struct {
	ID        string `json:".id,omitempty"`
	Disabled  string `json:"disabled,omitempty"`
	Comment   string `json:"comment,omitempty"`
	Name      string `json:"name"`
	StartDate string `json:"start-date"`
	StartTime string `json:"start-time"`
	Interval  string `json:"interval"`
	Policy    string `json:"policy"`
	OnEvent   string `json:"on-event"`
}

func (a *Adaptor) CreateScheduler(c context.Context, scheduler Scheduler) (*Scheduler, error) {
	var createdScheduler Scheduler

	err := a.httpClient.Put(
		c,
		SchedulerPath,
		scheduler,
		&createdScheduler,
	)
	if err != nil {
		return nil, err
	}

	return &createdScheduler, nil
}

func (a *Adaptor) DeleteScheduler(c context.Context, schedulerID string) error {
	err := a.httpClient.Delete(
		c,
		SchedulerPath+"/"+schedulerID,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}
