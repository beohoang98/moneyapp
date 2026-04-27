package services

import (
	"context"
	"fmt"
	"time"

	"github.com/beohoang98/moneyapp/internal/models"
	"gorm.io/gorm"
)

var ErrDateRangeInvalid = fmt.Errorf("date_from must be before or equal to date_to")

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

func (s *DashboardService) defaultDateRange() (string, string) {
	now := time.Now()
	dateFrom := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	dateTo := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	return dateFrom, dateTo
}

func (s *DashboardService) GetMonthlyTrend(ctx context.Context, dateFrom, dateTo string) ([]models.MonthlyTrendItem, error) {
	if dateFrom == "" || dateTo == "" {
		dateFrom, dateTo = s.defaultDateRange()
	}

	type monthAmount struct {
		Month string
		Total int64
	}

	var expenseRows []monthAmount
	if err := s.db.WithContext(ctx).
		Model(&models.Expense{}).
		Select("strftime('%Y-%m', date) AS month, COALESCE(SUM(amount), 0) AS total").
		Where("date BETWEEN ? AND ?", dateFrom, dateTo).
		Group("month").
		Order("month ASC").
		Scan(&expenseRows).Error; err != nil {
		return nil, fmt.Errorf("monthly expenses: %w", err)
	}

	var incomeRows []monthAmount
	if err := s.db.WithContext(ctx).
		Model(&models.Income{}).
		Select("strftime('%Y-%m', date) AS month, COALESCE(SUM(amount), 0) AS total").
		Where("date BETWEEN ? AND ?", dateFrom, dateTo).
		Group("month").
		Order("month ASC").
		Scan(&incomeRows).Error; err != nil {
		return nil, fmt.Errorf("monthly incomes: %w", err)
	}

	expenseMap := make(map[string]int64, len(expenseRows))
	for _, r := range expenseRows {
		expenseMap[r.Month] = r.Total
	}
	incomeMap := make(map[string]int64, len(incomeRows))
	for _, r := range incomeRows {
		incomeMap[r.Month] = r.Total
	}

	allMonths := contiguousMonths(dateFrom, dateTo)

	result := make([]models.MonthlyTrendItem, len(allMonths))
	for i, m := range allMonths {
		result[i] = models.MonthlyTrendItem{
			Month:         m,
			TotalIncome:   incomeMap[m],
			TotalExpenses: expenseMap[m],
		}
	}
	return result, nil
}

func (s *DashboardService) GetExpenseByCategory(ctx context.Context, dateFrom, dateTo string) ([]models.CategoryBreakdownItem, error) {
	if dateFrom == "" || dateTo == "" {
		dateFrom, dateTo = s.defaultDateRange()
	}

	var rows []models.CategoryBreakdownItem
	if err := s.db.WithContext(ctx).
		Table("expenses").
		Select("COALESCE(categories.name, 'Uncategorized') AS category_name, COALESCE(SUM(expenses.amount), 0) AS total").
		Joins("LEFT JOIN categories ON expenses.category_id = categories.id").
		Where("expenses.date BETWEEN ? AND ?", dateFrom, dateTo).
		Group("categories.id").
		Order("total DESC").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("expense by category: %w", err)
	}
	return rows, nil
}

func (s *DashboardService) ValidateDateRange(dateFrom, dateTo string) error {
	if dateFrom == "" || dateTo == "" {
		return nil
	}
	if dateFrom > dateTo {
		return ErrDateRangeInvalid
	}
	return nil
}

// contiguousMonths returns every YYYY-MM between dateFrom and dateTo inclusive.
func contiguousMonths(dateFrom, dateTo string) []string {
	start, err := time.Parse("2006-01-02", dateFrom)
	if err != nil {
		return nil
	}
	end, err := time.Parse("2006-01-02", dateTo)
	if err != nil {
		return nil
	}

	cur := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)

	var months []string
	for !cur.After(endMonth) {
		months = append(months, cur.Format("2006-01"))
		cur = cur.AddDate(0, 1, 0)
	}
	return months
}
