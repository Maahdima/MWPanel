package service

import (
	"context"
	"fmt"
	"strings"
	"time"

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

	startDate, err := toMikrotikDate(expireTime)
	if err != nil {
		s.logger.Error("failed to convert expire date for scheduler", zap.Error(err))
		return nil, err
	}

	scheduler := mikrotik.Scheduler{
		Comment:   utils.Ptr(common.SchedulerComment + peerName),
		Name:      common.SchedulerName + peerName,
		StartDate: startDate,
		StartTime: utils.Ptr(common.SchedulerStartTime),
		Interval:  utils.Ptr(common.SchedulerInterval),
		Policy:    utils.Ptr(common.SchedulerPolicy),
		OnEvent:   utils.Ptr(schedulerOnEvent(peerID)),
	}

	createdScheduler, err := s.mikrotikAdaptor.CreateScheduler(context.Background(), scheduler)
	if err != nil {
		s.logger.Error("failed to create scheduler for wireguard peer", zap.Error(err))
		return nil, err
	}

	return &createdScheduler.ID, nil
}

func (s *Scheduler) updateScheduler(schedulerID *string, peerID string, expireTime *string) error {
	startDate, err := toMikrotikDate(expireTime)
	if err != nil {
		s.logger.Error("failed to convert expire date for scheduler", zap.Error(err))
		return err
	}

	scheduler := mikrotik.Scheduler{
		StartDate: startDate,
		StartTime: utils.Ptr(common.SchedulerStartTime),
		OnEvent:   utils.Ptr(schedulerOnEvent(peerID)),
	}

	_, err = s.mikrotikAdaptor.UpdateScheduler(context.Background(), *schedulerID, scheduler)
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

func schedulerOnEvent(peerID string) string {
	return fmt.Sprintf(common.SchedulerEvent, peerID)
}

func toMikrotikDate(isoDate *string) (*string, error) {
	if isoDate == nil {
		return nil, nil
	}

	raw := strings.TrimSpace(*isoDate)
	if raw == "" {
		return nil, fmt.Errorf("expire date is empty")
	}

	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, fmt.Errorf("invalid expire date %q: %w", raw, err)
	}

	return utils.Ptr(strings.ToLower(parsed.Format("Jan/02/2006"))), nil
}
