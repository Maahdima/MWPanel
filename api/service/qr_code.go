package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils"
)

var (
	peerQrCodesPath string
)

type QRCodeGenerator struct {
	db     *gorm.DB
	logger *zap.Logger
}

func init() {
	appCfg := config.GetAppConfig()
	peerQrCodesPath = filepath.Join(appCfg.PeerFilesDir, "qrcode")
	if err := os.MkdirAll(peerQrCodesPath, os.ModePerm); err != nil {
		panic(fmt.Sprintf("failed to create QR code directory: %v", err))
	}
}

func NewQRCodeGenerator(db *gorm.DB) *QRCodeGenerator {
	return &QRCodeGenerator{
		db:     db,
		logger: zap.L().Named("QRCodeGenerator"),
	}
}

func (q *QRCodeGenerator) GetPeerQRCode(uuid string) (qrcodePath string, err error) {
	var peer model.Peer
	if err = q.db.First(&peer, "uuid = ?", uuid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			q.logger.Error("peer not found in database", zap.String("uuid", uuid))
			return
		}
		q.logger.Error("failed to get peer from database", zap.String("uuid", uuid), zap.Error(err))
		return
	}

	isSharable := utils.IsPeerSharable(peer.IsShared, peer.ShareExpireTime)
	if !isSharable {
		// TODO: return 404
		return "", fmt.Errorf("peer is not shared")
	}

	qrcodePath = fmt.Sprintf("%s/%s.jpeg", peerQrCodesPath, peer.UUID)

	return qrcodePath, nil
}

func (q *QRCodeGenerator) BuildPeerQRCode(config string, uuid string) error {
	qrc, err := qrcode.New(config)
	if err != nil {
		fmt.Printf("could not generate QRCode: %v", err)
		return err
	}

	filePath := fmt.Sprintf("%s/%s.jpeg", peerQrCodesPath, uuid)

	w, err := standard.New(filePath)
	if err != nil {
		fmt.Printf("standard.New failed: %v", err)
		return err
	}

	if err = qrc.Save(w); err != nil {
		fmt.Printf("could not save image: %v", err)
	}

	return nil
}
