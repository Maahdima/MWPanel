package mikrotik

import (
	"context"
	"github.com/labstack/echo/v4"
)

type Queue struct {
	ID       string `json:".id,omitempty"`
	Disabled string `json:"disabled,omitempty"`
	Comment  string `json:"comment,omitempty"`
	Name     string `json:"name"`
	Target   string `json:"target"`
	MaxLimit string `json:"max-limit"`
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

func (a *Adaptor) DeleteSimpleQueue(c echo.Context, queueID string) error {
	err := a.httpClient.Delete(
		c.Request().Context(),
		QueuePath+"/"+queueID,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}
