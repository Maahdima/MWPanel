package common

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils/httphelper"
)

type MwpClient struct {
	db      *gorm.DB
	mu      sync.RWMutex
	clients map[connectionKey]*httphelper.Client
	logger  *zap.Logger
}

func NewMwpClient(db *gorm.DB) *MwpClient {
	return &MwpClient{
		db:      db,
		mu:      sync.RWMutex{},
		clients: make(map[connectionKey]*httphelper.Client),
		logger:  zap.L().Named("mwpClients"),
	}
}

// IsConnected Check if the specified client is connected (not nil)
func (c *MwpClient) IsConnected(serverName *string) bool {
	client := c.GetClient(serverName)
	if client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return client.Get(ctx, DeviceIdentityPath, nil) == nil
}

// GetClient preserves the legacy API but selects the first active persisted server
// deterministically. Map iteration must never select a router for an operation.
func (c *MwpClient) GetClient(serverName *string) *httphelper.Client {
	var server model.Server
	query := c.db.Where("is_active = ?", true)
	if serverName != nil {
		query = query.Where("name = ?", *serverName)
	}
	if err := query.Order("id ASC").First(&server).Error; err != nil {
		return nil
	}
	client, err := c.clientForRecord(server)
	if err != nil {
		return nil
	}
	return client
}

// ClientForServer never falls back to another router when the requested server is
// missing or inactive. Callers must enforce resource authorization before binding.
func (c *MwpClient) ClientForServer(serverID uint) (*httphelper.Client, error) {
	if serverID == 0 {
		return nil, fmt.Errorf("server ID is required")
	}
	var server model.Server
	if err := c.db.Where("is_active = ?", true).First(&server, serverID).Error; err != nil {
		return nil, err
	}
	return c.clientForRecord(server)
}

func NewServerHTTPClient(server model.Server) (*httphelper.Client, error) {
	protocol := "http"
	if server.IsSSL {
		protocol = "https"
	}
	return httphelper.NewClient(httphelper.Config{
		BaseURL:  fmt.Sprintf("%s://%s/rest", protocol, net.JoinHostPort(server.IPAddress, strconv.Itoa(server.APIPort))),
		Username: server.Username, Password: server.Password, Timeout: 30 * time.Second,
	})
}

// Connection keys include persisted settings so edits never return a stale client.
type connectionKey struct {
	ID       uint
	Host     string
	Port     int
	Username string
	Password string
	TLS      bool
}

func (c *MwpClient) clientForRecord(server model.Server) (*httphelper.Client, error) {
	key := connectionKey{server.ID, server.IPAddress, server.APIPort, server.Username, server.Password, server.IsSSL}
	c.mu.Lock()
	defer c.mu.Unlock()
	if client := c.clients[key]; client != nil {
		return client, nil
	}
	client, err := NewServerHTTPClient(server)
	if err != nil {
		return nil, err
	}
	// Already bound adaptors retain their own immutable client during configuration edits.
	for old := range c.clients {
		if old.ID == key.ID {
			delete(c.clients, old)
		}
	}
	c.clients[key] = client
	return client, nil
}

func (c *MwpClient) InitClient() {
	var servers []model.Server
	if err := c.db.Where("is_active = ?", true).Find(&servers).Error; err != nil {
		c.logger.Error("failed to fetch servers", zap.Error(err))
		return
	}
	for _, server := range servers {
		if _, err := c.clientForRecord(server); err != nil {
			c.logger.Error("failed to initialize router client", zap.Uint("server_id", server.ID), zap.Error(err))
		}
	}
}
