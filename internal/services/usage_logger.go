package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
)

// UsageLogger 异步使用记录器（性能优化）
type UsageLogger struct {
	db          *database.DB
	logChannel  chan *models.UsageRecord
	batchSize   int
	flushInterval time.Duration
	wg          sync.WaitGroup
	stopChan    chan struct{}
}

// NewUsageLogger 创建异步使用记录器
func NewUsageLogger(db *database.DB, bufferSize, batchSize int, flushInterval time.Duration) *UsageLogger {
	if bufferSize < 100 {
		bufferSize = 1000 // 默认缓冲区大小
	}
	if batchSize < 10 {
		batchSize = 100 // 默认批量大小
	}
	if flushInterval == 0 {
		flushInterval = time.Second // 默认每秒刷新
	}

	ul := &UsageLogger{
		db:           db,
		logChannel:   make(chan *models.UsageRecord, bufferSize),
		batchSize:    batchSize,
		flushInterval: flushInterval,
		stopChan:     make(chan struct{}),
	}

	// 启动异步写入 goroutine
	ul.wg.Add(1)
	go ul.batchWriter()

	return ul
}

// batchWriter 批量写入使用记录
func (ul *UsageLogger) batchWriter() {
	defer ul.wg.Done()

	batch := make([]*models.UsageRecord, 0, ul.batchSize)
	ticker := time.NewTicker(ul.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case record := <-ul.logChannel:
			batch = append(batch, record)

			// 达到批量大小时立即写入
			if len(batch) >= ul.batchSize {
				ul.flushBatch(batch)
				batch = batch[:0] // 重置 batch
			}

		case <-ticker.C:
			// 定期刷新（即使未达到批量大小）
			if len(batch) > 0 {
				ul.flushBatch(batch)
				batch = batch[:0]
			}

		case <-ul.stopChan:
			// 停止信号，刷新剩余记录
			if len(batch) > 0 {
				ul.flushBatch(batch)
			}
			return
		}
	}
}

// flushBatch 批量写入数据库
func (ul *UsageLogger) flushBatch(batch []*models.UsageRecord) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startTime := time.Now()
	if err := ul.db.WithContext(ctx).CreateInBatches(batch, ul.batchSize).Error; err != nil {
		log.Printf("批量写入使用记录失败: %v (记录数=%d)", err, len(batch))
		// TODO: 实现重试机制或写入死信队列
		return
	}

	elapsed := time.Since(startTime)
	log.Printf("批量写入使用记录成功: 数量=%d 耗时=%v", len(batch), elapsed)
}

// RecordUsage 记录使用量（非阻塞）
func (ul *UsageLogger) RecordUsage(record *models.UsageRecord) error {
	select {
	case ul.logChannel <- record:
		return nil
	default:
		// 通道已满，记录警告并同步写入
		log.Printf("警告: 使用记录通道已满，回退到同步写入")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return ul.db.WithContext(ctx).Create(record).Error
	}
}

// RecordUsageBlocking 阻塞式记录使用量（用于重要记录）
func (ul *UsageLogger) RecordUsageBlocking(record *models.UsageRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ul.db.WithContext(ctx).Create(record).Error
}

// GetChannelSize 获取当前通道大小（用于监控）
func (ul *UsageLogger) GetChannelSize() int {
	return len(ul.logChannel)
}

// Close 关闭使用记录器
func (ul *UsageLogger) Close() error {
	close(ul.stopChan)
	ul.wg.Wait()
	return nil
}

// 使用记录器单例
var globalUsageLogger *UsageLogger

// InitUsageLogger 初始化全局使用记录器
func InitUsageLogger(db *database.DB) {
	globalUsageLogger = NewUsageLogger(db, 1000, 100, time.Second)
	log.Println("✅ 异步使用记录器已启动")
}

// GetUsageLogger 获取全局使用记录器
func GetUsageLogger() *UsageLogger {
	return globalUsageLogger
}

// RecordUsage 全局函数：记录使用量
func RecordUsage(record *models.UsageRecord) error {
	if globalUsageLogger != nil {
		return globalUsageLogger.RecordUsage(record)
	}
	// 回退到同步写入
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return globalUsageLogger.db.WithContext(ctx).Create(record).Error
}
