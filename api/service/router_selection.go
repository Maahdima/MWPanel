package service

import (
	"errors"

	"github.com/maahdima/mwp/api/dataservice/model"
	"gorm.io/gorm"
)

var ErrServerSelectionRequired = errors.New("select a server explicitly when more than one router is configured")

func resolveServerID(db *gorm.DB, requested uint) (uint, error) {
	if requested != 0 {
		var server model.Server
		if err := db.Where("is_active = ?", true).First(&server, requested).Error; err != nil {
			return 0, err
		}
		return server.ID, nil
	}
	var servers []model.Server
	if err := db.Where("is_active = ?", true).Order("id ASC").Limit(2).Find(&servers).Error; err != nil {
		return 0, err
	}
	if len(servers) == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	if len(servers) > 1 {
		return 0, ErrServerSelectionRequired
	}
	return servers[0].ID, nil
}

// ForServer binds the entire peer operation, including its queue and scheduler,
// without changing the shared service. Database identifiers remain global; router
// object IDs and addresses are always interpreted within this server.
func (w *WgPeer) ForServer(id uint) (*WgPeer, error) {
	adaptor, err := w.mikrotikAdaptor.ForServer(id)
	if err != nil {
		return nil, err
	}
	bound := *w
	bound.serverID = id
	bound.mikrotikAdaptor = adaptor
	bound.scheduler = NewScheduler(adaptor)
	bound.queue = NewQueue(adaptor)
	return &bound, nil
}

func (w *WgPeer) peerQuery() *gorm.DB {
	query := w.db
	if w.serverID != 0 {
		query = query.Where("server_id = ?", w.serverID)
	}
	if w.tenantID != nil {
		query = query.Where("reseller_id = ?", *w.tenantID)
	}
	return query
}

func (i *WgInterface) ForServer(id uint) (*WgInterface, error) {
	adaptor, err := i.mikrotikAdaptor.ForServer(id)
	if err != nil {
		return nil, err
	}
	peers, err := i.peerService.ForServer(id)
	if err != nil {
		return nil, err
	}
	bound := *i
	bound.serverID = id
	bound.mikrotikAdaptor = adaptor
	bound.peerService = peers
	return &bound, nil
}
