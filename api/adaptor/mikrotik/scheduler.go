package mikrotik

import (
	"context"

	"github.com/maahdima/mwp/api/common"
)

type Scheduler struct {
	ID        string  `json:".id,omitempty"`
	Disabled  string  `json:"disabled,omitempty"`
	Comment   *string `json:"comment,omitempty"`
	Name      string  `json:"name,omitempty"`
	StartDate *string `json:"start-date,omitempty"`
	StartTime *string `json:"start-time,omitempty"`
	Interval  *string `json:"interval,omitempty"`
	Policy    *string `json:"policy,omitempty"`
	OnEvent   *string `json:"on-event,omitempty"`
}

func (a *Adaptor) CreateScheduler(c context.Context, scheduler Scheduler) (*Scheduler, error) {
	var createdScheduler Scheduler

	httpClient, err := a.client()

	if err != nil {
		return nil, err
	}

	err = httpClient.Put(
		c,
		common.SchedulerPath,
		scheduler,
		&createdScheduler,
	)

	if err != nil {
		return nil, err
	}

	return &createdScheduler, nil
}

func (a *Adaptor) UpdateScheduler(c context.Context, schedulerID string, scheduler Scheduler) (*Scheduler, error) {
	var updatedScheduler Scheduler

	httpClient, err := a.client()

	if err != nil {
		return nil, err
	}

	err = httpClient.Patch(
		c,
		common.SchedulerPath+"/"+schedulerID,
		scheduler,
		&updatedScheduler,
	)

	if err != nil {
		return nil, err
	}

	return &updatedScheduler, nil
}

func (a *Adaptor) DeleteScheduler(c context.Context, schedulerID string) error {
	httpClient, err := a.client()

	if err != nil {
		return err
	}

	err = httpClient.Delete(
		c,
		common.SchedulerPath+"/"+schedulerID,
		nil,
	)

	if err != nil {
		return err
	}

	return nil
}

func (a *Adaptor) FetchScheduler(ctx context.Context, id string) (*Scheduler, error) {
	client, err := a.client()

	if err != nil {
		return nil, err
	}

	var scheduler Scheduler

	if err := client.Get(ctx, common.SchedulerPath+"/"+id, &scheduler); err != nil {
		return nil, err
	}

	return &scheduler, nil
}

func (a *Adaptor) FetchSchedulers(ctx context.Context) ([]Scheduler, error) {
	client, err := a.client()

	if err != nil {
		return nil, err
	}

	var rows []Scheduler

	err = client.Get(ctx, common.SchedulerPath, &rows)

	return rows, err
}
