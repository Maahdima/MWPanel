package mikrotik

import (
	"context"
)

type Queue struct {
	ID       *string `json:".id,omitempty"`
	Disabled *string `json:"disabled,omitempty"`
	Comment  *string `json:"comment,omitempty"`
	Name     *string `json:"name,omitempty"`
	Target   *string `json:"target,omitempty"`
	MaxLimit *string `json:"max-limit,omitempty"`
}

func (a *Adaptor) CreateSimpleQueue(c context.Context, queue Queue) (*Queue, error) {
	var createdQueue Queue

	err := a.httpClient.Put(
		c,
		QueuePath,
		queue,
		&createdQueue,
	)
	if err != nil {
		return nil, err
	}

	return &createdQueue, nil
}

func (a *Adaptor) UpdateSimpleQueue(c context.Context, queueID string, queue Queue) (*Queue, error) {
	var updatedQueue Queue

	err := a.httpClient.Patch(
		c,
		QueuePath+"/"+queueID,
		queue,
		&updatedQueue,
	)
	if err != nil {
		return nil, err
	}

	return &updatedQueue, nil
}

func (a *Adaptor) DeleteSimpleQueue(c context.Context, queueID string) error {
	err := a.httpClient.Delete(
		c,
		QueuePath+"/"+queueID,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}
