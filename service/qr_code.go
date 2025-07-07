package service

import (
	"fmt"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"mikrotik-wg-go/dataservice/model"
	"os"
)

type QRCodeGenerator struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewQRCodeGenerator(db *gorm.DB) *QRCodeGenerator {
	return &QRCodeGenerator{
		db:     db,
		logger: zap.L().Named("QRCodeGenerator"),
	}
}

func (q *QRCodeGenerator) GetPeerQRCode(id uint) (qrcodePath string, err error) {
	var peer model.Peer
	if err = q.db.First(&peer, "id = ?", id).Error; err != nil {
		q.logger.Error("failed to get peer from database", zap.Uint("id", id), zap.Error(err))
		return
	}

	qrcodePath = fmt.Sprintf("./%s/%s.jpeg", peerQrCodesPath, peer.Name)

	return qrcodePath, nil
}

func (q *QRCodeGenerator) BuildPeerQRCode(config string, peerName string) error {
	dirPath := fmt.Sprintf("./%s", peerQrCodesPath)

	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	qrc, err := qrcode.New(config)
	if err != nil {
		fmt.Printf("could not generate QRCode: %v", err)
		return err
	}

	filePath := fmt.Sprintf("%s/%s.jpeg", dirPath, peerName)
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
