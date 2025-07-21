package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/utils"
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

func (s *Scheduler) createScheduler(peerID, peerName string, expireTime *string) (*string, error) {
	if expireTime == nil {
		return nil, nil
	}

	scheduler := mikrotik.Scheduler{
		Comment:   utils.Ptr(common.SchedulerComment + peerName),
		Name:      common.SchedulerName + peerName,
		StartDate: expireTime,
		StartTime: utils.Ptr(common.SchedulerStartTime),
		Interval:  utils.Ptr(common.SchedulerInterval),
		Policy:    utils.Ptr(common.SchedulerPolicy),
		OnEvent:   utils.Ptr(common.SchedulerEvent + peerID),
	}

	createdScheduler, err := s.mikrotikAdaptor.CreateScheduler(context.Background(), scheduler)
	if err != nil {
		s.logger.Error("failed to create scheduler for wireguard peer", zap.Error(err))
		return nil, err
	}

	return &createdScheduler.ID, nil
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
