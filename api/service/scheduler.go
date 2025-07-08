package service

import (
	"context"
	"go.uber.org/zap"
	"mikrotik-wg-go/adaptor/mikrotik"
	"mikrotik-wg-go/utils"
)

type Scheduler struct {
	mikrotikAdaptor *mikrotik.Adaptor
	logger          *zap.Logger
}

func NewScheduler(mikrotikAdaptor *mikrotik.Adaptor) *Scheduler {
	return &Scheduler{
		mikrotikAdaptor: mikrotikAdaptor,
		logger:          zap.L().Named("SchedulerService"),
	}
}

func (s *Scheduler) createScheduler(peerName, peerID string, expireTime *string) (*string, error) {
	if expireTime == nil {
		return nil, nil
	}

	scheduler := mikrotik.Scheduler{
		Comment:   utils.Ptr(schedulerComment + peerName),
		Name:      utils.Ptr(schedulerName + peerName),
		StartDate: expireTime,
		StartTime: utils.Ptr(schedulerStartTime),
		Interval:  utils.Ptr(schedulerInterval),
		Policy:    utils.Ptr(schedulerPolicy),
		OnEvent:   utils.Ptr(schedulerEvent + peerID),
	}

	createdScheduler, err := s.mikrotikAdaptor.CreateScheduler(context.Background(), scheduler)
	if err != nil {
		s.logger.Error("failed to create scheduler for wireguard peer", zap.Error(err))
		return nil, err
	}

	return createdScheduler.ID, nil
}

func (s *Scheduler) updateScheduler(schedulerID, expireTime *string) error {
	scheduler := mikrotik.Scheduler{
		StartDate: expireTime,
	}

	_, err := s.mikrotikAdaptor.UpdateScheduler(context.Background(), *schedulerID, scheduler)
	if err != nil {
		s.logger.Error("failed to update scheduler for wireguard peer", zap.String("schedulerID", *schedulerID), zap.Error(err))
		return err
	}

	return nil
}

func (s *Scheduler) deleteScheduler(schedulerID *string) error {
	if schedulerID == nil {
		return nil
	}

	err := s.mikrotikAdaptor.DeleteScheduler(context.Background(), *schedulerID)
	if err != nil {
		s.logger.Error("failed to delete scheduler", zap.String("schedulerID", *schedulerID), zap.Error(err))
		return err
	}

	return nil
}
