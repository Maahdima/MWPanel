package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/utils"
	"github.com/maahdima/mwp/api/utils/httphelper"
)

type MwpClients struct {
	db      *gorm.DB
	mu      sync.RWMutex
	clients map[string]*httphelper.Client
	logger  *zap.Logger
}

func NewMwpClients(db *gorm.DB) *MwpClients {
	return &MwpClients{
		db:      db,
		mu:      sync.RWMutex{},
		clients: make(map[string]*httphelper.Client),
		logger:  zap.L().Named("mwpClients"),
	}
}

// IsConnected Check if the specified client is connected (not nil)
func (c *MwpClients) IsConnected(serverName *string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var server model.Server
	var name string

	if serverName == nil {
		if err := c.db.First(&server).Error; err != nil {
			c.logger.Error("Failed to fetch server from database", zap.Error(err))
			return false
		}
		name = server.Name
	} else {
		name = utils.DerefString(serverName)

		if err := c.db.Where("name = ?", name).First(&server).Error; err != nil {
			c.logger.Error("Failed to fetch server by name", zap.String("serverName", name), zap.Error(err))
			return false
		}
	}

	client, ok := c.clients[name]
	if !ok || client == nil {
		c.logger.Error("Client not found in mwp clients", zap.String("serverName", name))
		return false
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	uri := fmt.Sprintf("http://%s:%d/rest/system/identity", server.IPAddress, server.APIPort)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, uri, nil)
	if err != nil {
		c.logger.Error("Failed to create request to mikrotik REST API", zap.String("serverName", name), zap.Error(err))
		return false
	}

	req.SetBasicAuth(server.Username, server.Password)

	resp, err := httpClient.Do(req)
	if err != nil {
		c.logger.Error("HTTP request to mikrotik REST API failed", zap.String("serverName", name), zap.Error(err))
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Unexpected status code from mikrotik REST API",
			zap.String("serverName", name),
			zap.Int("status", resp.StatusCode))
		return false
	}

	return true
}

// GetClient Get client for a specific server
func (c *MwpClients) GetClient(serverName *string) *httphelper.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if serverName == nil {
		// get the first item in the client map
		for _, client := range c.clients {
			if client != nil {
				return client
			}

			return nil
		}
	}

	if client, ok := c.clients[utils.DerefString(serverName)]; ok {
		return client
	}

	zap.L().Error("client not found in mwp clients", zap.String("serverName", utils.DerefString(serverName)))
	return nil
}

// SetClient Set or update the client for a specific server
func (c *MwpClients) SetClient(serverData *schema.CreateServerRequest) {
	c.mu.Lock()
	defer c.mu.Unlock()

	apiPort := utils.ParseStringToInt(serverData.APIPort)
	isSSL := *serverData.IsSSL

	var protocol string
	if isSSL {
		protocol = "https"
	} else {
		protocol = "http"
	}

	client, err := httphelper.NewClient(httphelper.Config{
		BaseURL:            fmt.Sprintf("%s://%s:%d/rest", protocol, serverData.IPAddress, apiPort),
		Username:           serverData.Username,
		Password:           serverData.Password,
		InsecureSkipVerify: !isSSL,
	})
	if err != nil {
		c.logger.Panic("Failed to create HTTP client", zap.Error(err))
	}

	c.clients[serverData.Name] = client
}

// InitClient Set or update the client for a specific server
func (c *MwpClients) InitClient() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var servers []model.Server
	if err := c.db.Find(&servers).Error; err != nil {
		c.logger.Error("Failed to fetch servers from database", zap.Error(err))
		return
	}

	for _, server := range servers {
		client, err := httphelper.NewClient(httphelper.Config{
			BaseURL:            fmt.Sprintf("%s://%s:%d/rest", "http", server.IPAddress, server.APIPort),
			Username:           server.Username,
			Password:           server.Password,
			InsecureSkipVerify: true,
		})
		if err != nil {
			c.logger.Panic("Failed to create HTTP client", zap.Error(err))
		}

		c.clients[server.Name] = client
	}
}

// DeleteClient Remove the client for a specific server
func (c *MwpClients) DeleteClient(serverName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.clients, serverName)
}
