package mikrotik

import (
	"context"

	"github.com/maahdima/mwp/api/common"
)

type Queue struct {
	ID       string  `json:".id,omitempty"`
	Disabled string  `json:"disabled,omitempty"`
	Comment  *string `json:"comment,omitempty"`
	Name     string  `json:"name,omitempty"`
	Target   *string `json:"target,omitempty"`
	MaxLimit *string `json:"max-limit,omitempty"`
}

func (a *Adaptor) CreateSimpleQueue(c context.Context, queue Queue) (*Queue, error) {
	var createdQueue Queue

	httpClient := a.mwpClients.GetClient(nil)

	err := httpClient.Put(
		c,
		common.QueuePath,
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

	httpClient := a.mwpClients.GetClient(nil)

	err := httpClient.Patch(
		c,
		common.QueuePath+"/"+queueID,
		queue,
		&updatedQueue,
	)
	if err != nil {
		return nil, err
	}

	return &updatedQueue, nil
}

func (a *Adaptor) DeleteSimpleQueue(c context.Context, queueID string) error {
	httpClient := a.mwpClients.GetClient(nil)

	err := httpClient.Delete(
		c,
		common.QueuePath+"/"+queueID,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}
