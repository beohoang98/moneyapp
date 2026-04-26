package services

import (
	"context"
	"fmt"
	"time"

	"github.com/beohoang98/moneyapp/internal/models"
	"gorm.io/gorm"
)

type DashboardService struct {
	db             *gorm.DB
	invoiceService *InvoiceService
}

func NewDashboardService(db *gorm.DB, invoiceService *InvoiceService) *DashboardService {
	return &DashboardService{db: db, invoiceService: invoiceService}
}

func (s *DashboardService) GetSummary(ctx context.Context, dateFrom, dateTo string) (*models.DashboardSummary, error) {
	if dateFrom == "" || dateTo == "" {
		now := time.Now()
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		lastDay := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location())
		dateTo = lastDay.Format("2006-01-02")
	}

	var totalExpenses int64
	if err := s.db.WithContext(ctx).Model(&models.Expense{}).
		Where("date BETWEEN ? AND ?", dateFrom, dateTo).
		Select("COALESCE(SUM(amount), 0)").
		Row().Scan(&totalExpenses); err != nil {
		return nil, fmt.Errorf("sum expenses: %w", err)
	}

	var totalIncome int64
	if err := s.db.WithContext(ctx).Model(&models.Income{}).
		Where("date BETWEEN ? AND ?", dateFrom, dateTo).
		Select("COALESCE(SUM(amount), 0)").
		Row().Scan(&totalIncome); err != nil {
		return nil, fmt.Errorf("sum incomes: %w", err)
	}

	stats, err := s.invoiceService.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get invoice stats: %w", err)
	}

	return &models.DashboardSummary{
		TotalIncome:      totalIncome,
		TotalExpenses:    totalExpenses,
		NetBalance:       totalIncome - totalExpenses,
		DateFrom:         dateFrom,
		DateTo:           dateTo,
		UnpaidCount:      stats.UnpaidCount,
		UnpaidAmount:     stats.UnpaidAmount,
		OverdueCount:     stats.OverdueCount,
		OverdueAmount:    stats.OverdueAmount,
		TotalOutstanding: stats.TotalOutstanding,
	}, nil
}
