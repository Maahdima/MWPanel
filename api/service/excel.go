package service

import (
	"sort"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xuri/excelize/v2"

	"github.com/maahdima/mwp/api/dataservice/model"
)

type ExcelGenerator struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewExcelGenerator(db *gorm.DB) *ExcelGenerator {
	return &ExcelGenerator{
		db:     db,
		logger: zap.L().Named("ExcelGenerator"),
	}
}

func (e *ExcelGenerator) GetTrafficUsageReport() (filePath string, err error) {
	var peers []model.Peer
	sheetName := "traffic-usage"
	filePath = "traffic-report.xlsx"

	if err = e.db.Find(&peers).Error; err != nil {
		e.logger.Error("failed to get peers from database", zap.Error(err))
		return "", err
	}

	excelFile := excelize.NewFile()
	defer func() {
		if closeErr := excelFile.Close(); closeErr != nil {
			e.logger.Warn("failed to close excel file", zap.Error(closeErr))
		}
	}()

	sheetIndex, err := excelFile.NewSheet(sheetName)
	if err != nil {
		e.logger.Error("failed to create new sheet in excel file", zap.Error(err))
		return "", err
	}

	style, err := excelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 20,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		e.logger.Error("failed to create new style", zap.Error(err))
		return "", err
	}

	if err = excelFile.SetCellStyle(sheetName, "A1", "D1000", style); err != nil {
		e.logger.Error("failed to set cell style", zap.Error(err))
		return "", err
	}

	if err = e.setHeaders(excelFile, sheetName); err != nil {
		return "", err
	}

	if err = e.setColumnsData(excelFile, peers, sheetName); err != nil {
		return "", err
	}

	excelFile.SetActiveSheet(sheetIndex)
	if err := excelFile.DeleteSheet("Sheet1"); err != nil {
		e.logger.Warn("failed to delete default sheet", zap.Error(err))
	}

	if err := excelFile.SaveAs(filePath); err != nil {
		e.logger.Error("failed to save excel file", zap.Error(err))
		return "", err
	}

	return filePath, nil
}

func (e *ExcelGenerator) setHeaders(excelFile *excelize.File, sheetName string) error {
	headers := []string{"ID", "Name", "Comment", "Traffic (GB)"}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		// set width for each column
		if err := excelFile.SetColWidth(sheetName, cell[:1], cell[:1], 20); err != nil {
			e.logger.Error("failed to set column width", zap.String("column", cell[:1]), zap.Error(err))
			return err
		}

		if err := excelFile.SetCellValue(sheetName, cell, header); err != nil {
			e.logger.Error("failed to set header cell value", zap.String("cell", cell), zap.Error(err))
			return err
		}
	}

	return nil
}

func (e *ExcelGenerator) setColumnsData(excelFile *excelize.File, peers []model.Peer, sheetName string) error {
	sort.Slice(peers, func(i, j int) bool {
		totalI := peers[i].DownloadUsage + peers[i].UploadUsage
		totalJ := peers[j].DownloadUsage + peers[j].UploadUsage
		return totalI > totalJ
	})

	for idx, peer := range peers {
		rowIndex := idx + 2

		idCell, _ := excelize.CoordinatesToCellName(1, rowIndex)
		nameCell, _ := excelize.CoordinatesToCellName(2, rowIndex)
		commentCell, _ := excelize.CoordinatesToCellName(3, rowIndex)
		usageCell, _ := excelize.CoordinatesToCellName(4, rowIndex)

		if err := excelFile.SetCellValue(sheetName, idCell, peer.ID); err != nil {
			e.logger.Error("failed to set ID cell value", zap.String("cell", idCell), zap.Error(err))
			return err
		}

		if err := excelFile.SetCellValue(sheetName, nameCell, peer.Name); err != nil {
			e.logger.Error("failed to set Name cell value", zap.String("cell", nameCell), zap.Error(err))
			return err
		}

		comment := "-"
		if peer.Comment != nil {
			comment = *peer.Comment
		}
		if err := excelFile.SetCellValue(sheetName, commentCell, comment); err != nil {
			e.logger.Error("failed to set Comment cell value", zap.String("cell", commentCell), zap.Error(err))
			return err
		}

		totalUsageGB := float64(peer.DownloadUsage+peer.UploadUsage) / float64(1024*1024*1024)
		if err := excelFile.SetCellFloat(sheetName, usageCell, totalUsageGB, 2, 64); err != nil {
			e.logger.Error("failed to set traffic usage cell value", zap.String("cell", usageCell), zap.Error(err))
			return err
		}
	}

	return nil
}
