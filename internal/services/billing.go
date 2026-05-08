package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rs/xid"
	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
)

// BillingService 账单服务
type BillingService struct {
	db *database.DB
}

// NewBillingService 创建账单服务
func NewBillingService(db *database.DB) (*BillingService, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	service := &BillingService{db: db}
	log.Println("✅ 账单服务初始化成功")
	return service, nil
}

// GenerateMonthlyBill 生成月度账单 (per D-17: 每月1号自动生成)
func (bs *BillingService) GenerateMonthlyBill(ctx context.Context, userID string, year, month int) (*models.Bill, error) {
	// 计算账期
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	log.Printf("生成月度账单: 用户=%s 账期=%s", userID, startDate.Format("2006-01"))

	// 查询使用记录
	var records []models.UsageRecord
	err := bs.db.WithContext(ctx).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startDate, endDate).
		Order("created_at ASC").
		Find(&records).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch usage records: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no usage records for period %s", startDate.Format("2006-01"))
	}

	// 创建账单
	bill := &models.Bill{
		ID:        xid.New().String(),
		UserID:    userID,
		Period:    startDate.Format("2006-01"),
		StartDate: startDate,
		EndDate:   endDate.Add(-time.Second),
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
	}

	// 按模型分组统计
	modelStats := make(map[string]*models.BillModelDetail)
	for _, record := range records {
		if stat, exists := modelStats[record.ModelID]; exists {
			stat.RequestCount++
			stat.InputTokens += record.InputTokens
			stat.OutputTokens += record.OutputTokens
			stat.TotalTokens += record.TotalTokens
			stat.TotalCost += record.CostPrice
			stat.TotalRevenue += record.SellingPrice
			stat.TotalProfit += record.Profit
		} else {
			modelStats[record.ModelID] = &models.BillModelDetail{
				ModelID:       record.ModelID,
				RequestCount:  1,
				InputTokens:   record.InputTokens,
				OutputTokens:  record.OutputTokens,
				TotalTokens:   record.TotalTokens,
				TotalCost:     record.CostPrice,
				TotalRevenue:  record.SellingPrice,
				TotalProfit:   record.Profit,
			}
		}
	}

	// 计算总计
	for _, stat := range modelStats {
		bill.Items = append(bill.Items, *stat)
		bill.TotalCost += stat.TotalCost
		bill.TotalRevenue += stat.TotalRevenue
		bill.TotalProfit += stat.TotalProfit
		bill.TotalRequests += int64(stat.RequestCount)
	}

	// 保存账单
	if err := bs.db.WithContext(ctx).Create(bill).Error; err != nil {
		return nil, fmt.Errorf("failed to save bill: %w", err)
	}

	log.Printf("✅ 账单生成成功: ID=%s 用户=%s 总额=%.2f",
		bill.ID, userID, bill.TotalRevenue)

	return bill, nil
}

// GetUserBills 获取用户账单列表
func (bs *BillingService) GetUserBills(ctx context.Context, userID string, limit, offset int) ([]models.Bill, error) {
	var bills []models.Bill
	query := bs.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&bills).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bills: %w", err)
	}

	return bills, nil
}

// GetBillDetails 获取账单明细
func (bs *BillingService) GetBillDetails(ctx context.Context, billID, userID string) (*models.Bill, error) {
	var bill models.Bill
	err := bs.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", billID, userID).
		First(&bill).Error

	if err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}

	return &bill, nil
}

// GetBillByPeriod 获取指定时期的账单（优化版，解决 N+1 问题）
func (bs *BillingService) GetBillByPeriod(ctx context.Context, userID, period string) (*models.Bill, error) {
	var bill models.Bill
	if err := bs.db.WithContext(ctx).Where("user_id = ? AND period = ?", userID, period).First(&bill).Error; err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}

	// 一次性查询所有明细，避免 N+1
	var items []models.BillItem
	if err := bs.db.WithContext(ctx).Where("bill_id = ?", bill.ID).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to load bill items: %w", err)
	}

	// 转换为 BillModelDetail
	details := make([]models.BillModelDetail, len(items))
	for i, item := range items {
		details[i] = models.BillModelDetail{
			ModelID:       item.ModelID,
			RequestCount:  item.RequestCount,
			InputTokens:   item.InputTokens,
			OutputTokens:  item.OutputTokens,
			TotalTokens:   item.TotalTokens,
			TotalCost:     item.TotalCost,
			TotalRevenue:  item.TotalRevenue,
			TotalProfit:   item.TotalProfit,
		}
	}
	bill.Items = details

	return &bill, nil
}

// ExportBillAsPDF 导出账单为 PDF (per D-18: 总结性 PDF)
func (bs *BillingService) ExportBillAsPDF(ctx context.Context, billID string) ([]byte, error) {
	bill, err := bs.getBillByID(ctx, billID)
	if err != nil {
		return nil, err
	}

	// TODO: 实现 PDF 生成
	// 使用库如 github.com/jung-kurt/gofpdf 或 github.com/signintech/gopdf
	//
	// PDF 内容应包含:
	// - 账单基本信息 (ID, 账期, 总额)
	// - 按模型汇总的使用量和费用
	// - 总计: 请求次数、Token 数、成本、收入、利润
	//
	// 示例结构:
	// ==================
	//   AI GATEWAY 账单
	// ==================
	// 账单号: BILL-XXX
	// 账期: 2024-05
	// 状态: pending/paid
	//
	// 模型使用汇总:
	// GPT-4: 1000 请求, 1M tokens, ¥100.00
	// GPT-3.5: 5000 请求, 5M tokens, ¥50.00
	//
	// 总计:
	// 总请求: 6000
	// 总 Token: 6M
	// 总成本: ¥30.00
	// 总收入: ¥150.00
	// 总利润: ¥120.00

	log.Printf("导出账单 PDF: ID=%s 用户=%s 账期=%s", billID, bill.UserID, bill.Period)

	return nil, fmt.Errorf("PDF export not yet implemented")
}

// ExportBillAsCSV 导出账单明细为 CSV (per D-18: 明细 CSV)
func (bs *BillingService) ExportBillAsCSV(ctx context.Context, billID string) ([]byte, error) {
	bill, err := bs.getBillByID(ctx, billID)
	if err != nil {
		return nil, err
	}

	// 获取详细的使用记录
	var records []models.UsageRecord
	err = bs.db.WithContext(ctx).
		Where("user_id = ? AND created_at >= ? AND created_at < ?",
			bill.UserID, bill.StartDate, bill.EndDate).
		Order("created_at ASC").
		Find(&records).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch usage records: %w", err)
	}

	// TODO: 实现 CSV 生成
	// CSV 格式 (使用 encoding/csv):
	// RequestID,Timestamp,Model,InputTokens,OutputTokens,TotalTokens,Cost,Revenue,Profit
	// req-001,2024-05-01T10:00:00Z,gpt-4,100,50,150,0.01,0.05,0.04
	// req-002,2024-05-01T11:00:00Z,gpt-3.5-turbo,200,100,300,0.002,0.01,0.008

	log.Printf("导出账单 CSV: ID=%s 记录数=%d", billID, len(records))

	return nil, fmt.Errorf("CSV export not yet implemented")
}

// AutoGenerateMonthlyBills 自动生成月度账单 (定时任务)
func (bs *BillingService) AutoGenerateMonthlyBills(ctx context.Context) error {
	now := time.Now().UTC()

	// 只在每月1号执行
	if now.Day() != 1 {
		return nil
	}

	// 计算上个月的年月
	lastMonth := now.AddDate(0, -1, 0)
	year := lastMonth.Year()
	month := int(lastMonth.Month())

	log.Printf("开始自动生成月度账单: %d-%d", year, month)

	// 获取所有有使用记录的用户
	var userIDs []string
	err := bs.db.WithContext(ctx).
		Model(&models.UsageRecord{}).
		Select("DISTINCT user_id").
		Where("created_at >= ? AND created_at < ?",
			time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC),
			time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)).
		Pluck("user_id", &userIDs).Error

	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	// 为每个用户生成账单
	successCount := 0
	for _, userID := range userIDs {
		_, err := bs.GenerateMonthlyBill(ctx, userID, year, month)
		if err != nil {
			log.Printf("警告: 为用户 %s 生成账单失败: %v", userID, err)
			continue
		}
		successCount++
	}

	log.Printf("✅ 自动生成账单完成: 成功=%d/%d", successCount, len(userIDs))
	return nil
}

// getBillByID 根据 ID 获取账单
func (bs *BillingService) getBillByID(ctx context.Context, billID string) (*models.Bill, error) {
	var bill models.Bill
	err := bs.db.WithContext(ctx).
		Where("id = ?", billID).
		First(&bill).Error

	if err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}

	return &bill, nil
}
