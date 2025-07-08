package service

import (
	"context"
	"go.uber.org/zap"
	"mikrotik-wg-go/adaptor/mikrotik"
	"mikrotik-wg-go/utils"
)

type Queue struct {
	mikrotikAdaptor *mikrotik.Adaptor
	logger          *zap.Logger
}

func NewQueue(mikrotikAdaptor *mikrotik.Adaptor) *Queue {
	return &Queue{
		mikrotikAdaptor: mikrotikAdaptor,
		logger:          zap.L().Named("QueueService"),
	}
}

func (q *Queue) createQueue(peerName, peerAllowedAddress string, downloadBandwidth, uploadBandwidth *string) (*string, error) {
	normalizedDownload := downloadBandwidth
	if normalizedDownload == nil {
		normalizedDownload = utils.Ptr("0")
	}
	normalizedUpload := uploadBandwidth
	if normalizedUpload == nil {
		normalizedUpload = utils.Ptr("0")
	}

	maxLimit := *normalizedDownload + "/" + *normalizedUpload
	wgQueue := mikrotik.Queue{
		Comment:  utils.Ptr(queueComment + peerName),
		Name:     utils.Ptr(queueName + peerName),
		Target:   utils.Ptr(peerAllowedAddress),
		MaxLimit: &maxLimit,
	}

	createdQueue, err := q.mikrotikAdaptor.CreateSimpleQueue(context.Background(), wgQueue)
	if err != nil {
		q.logger.Error("failed to create simple queue for wireguard peer", zap.Error(err))
		return nil, err
	}

	return createdQueue.ID, nil
}

func (q *Queue) updateQueue(queueID, downloadBandwidth, uploadBandwidth *string) error {
	normalizedDownload := downloadBandwidth
	if normalizedDownload == nil {
		normalizedDownload = utils.Ptr("0")
	}
	normalizedUpload := uploadBandwidth
	if normalizedUpload == nil {
		normalizedUpload = utils.Ptr("0")
	}

	maxLimit := *normalizedDownload + "/" + *normalizedUpload
	queue := mikrotik.Queue{
		MaxLimit: &maxLimit,
	}

	_, err := q.mikrotikAdaptor.UpdateSimpleQueue(context.Background(), *queueID, queue)
	if err != nil {
		q.logger.Error("failed to update simple queue for wireguard peer", zap.String("queueId", *queueID), zap.Error(err))
		return err
	}

	return nil
}

func (q *Queue) deleteQueue(queueID *string) error {
	if queueID == nil {
		return nil
	}

	err := q.mikrotikAdaptor.DeleteSimpleQueue(context.Background(), *queueID)
	if err != nil {
		q.logger.Error("failed to delete simple queue", zap.String("queueId", *queueID), zap.Error(err))
		return err
	}

	return nil
}
