import { test, expect } from '@playwright/test';

// 定价管理 E2E 测试
test.describe('Pricing Management', () => {
  test.beforeEach(async ({ page }) => {
    // 登录管理员
    await page.goto('/admin/login');
    await page.fill('input[name="email"]', 'admin@example.com');
    await page.fill('input[name="password"]', 'admin123456');
    await page.click('button[type="submit"]');
    await page.waitForURL('/admin');
  });

  test('should display enterprise pricing list', async ({ page }) => {
    await page.goto('/admin/pricing');

    // 等待标签页加载
    await page.waitForSelector('button:has-text("大客户定价")');

    // 点击大客户定价标签
    await page.click('button:has-text("大客户定价")');

    // 等待表格加载
    await page.waitForSelector('table');

    // 验证表格列存在
    await expect(page.locator('table thead tr th')).toContainText(['客户名称', '模型 ID', '输入价格', '输出价格', '利润率', '状态', '操作']);
  });

  test('should create enterprise pricing', async ({ page }) => {
    await page.goto('/admin/pricing');

    // 点击大客户定价标签
    await page.click('button:has-text("大客户定价")');

    // 点击创建按钮
    await page.click('button:has-text("创建定价")');

    // 填写表单
    await page.fill('input[name="customer_id"]', 'enterprise-e2e-' + Date.now());
    await page.fill('input[name="customer_name"]', 'E2E Test Enterprise');
    await page.fill('input[name="model_id"]', 'gpt-4');

    // 使用小数值确保通过验证
    await page.fill('input[name="input_price"]', '0.05');
    await page.fill('input[name="output_price"]', '0.10');
    await page.fill('input[name="min_profit_margin"]', '10');
    await page.fill('input[name="max_cost_price"]', '0.04');

    // 设置生效日期
    const today = new Date().toISOString().split('T')[0];
    await page.fill('input[name="effective_date"]', today);

    // 提交表单
    await page.click('button:has-text("确定")');

    // 验证成功消息
    await expect(page.locator('.toast-success')).toBeVisible({ timeout: 5000 });
  });

  test('should validate form inputs', async ({ page }) => {
    await page.goto('/admin/pricing');

    // 点击大客户定价标签
    await page.click('button:has-text("大客户定价")');

    // 点击创建按钮
    await page.click('button:has-text("创建定价")');

    // 提交空表单（应该触发验证错误）
    await page.click('button:has-text("确定")');

    // 验证错误提示
    await expect(page.locator('.error-message:has-text("客户 ID")')).toBeVisible();
  });

  test('should calculate profit margin in real-time', async ({ page }) => {
    await page.goto('/admin/pricing');

    // 点击大客户定价标签
    await page.click('button:has-text("大客户定价")');

    // 点击创建按钮
    await page.click('button:has-text("创建定价")');

    // 填写必填字段
    await page.fill('input[name="customer_id"]', 'test-profit-margin');
    await page.fill('input[name="customer_name"]', 'Test Enterprise');
    await page.fill('input[name="model_id"]', 'gpt-4');

    // 输入价格
    await page.fill('input[name="input_price"]', '0.05');
    await page.fill('input[name="output_price"]', '0.10');

    // 验证利润率计算显示
    const profitDisplay = page.locator('.profit-margin-display');
    await expect(profitDisplay).toBeVisible();
  });

  test('should update enterprise pricing', async ({ page }) => {
    await page.goto('/admin/pricing');

    // 点击大客户定价标签
    await page.click('button:has-text("大客户定价")');

    // 等待表格加载
    await page.waitForSelector('table');

    // 点击第一个定价记录的编辑按钮
    await page.click('table tbody tr:first-child button:has-text("编辑")');

    // 修改价格
    await page.fill('input[name="input_price"]', '0.055');
    await page.fill('input[name="output_price"]', '0.11');

    // 提交表单
    await page.click('button:has-text("确定")');

    // 验证成功消息
    await expect(page.locator('.toast-success')).toBeVisible();
  });

  test('should delete enterprise pricing with confirmation', async ({ page }) => {
    await page.goto('/admin/pricing');

    // 点击大客户定价标签
    await page.click('button:has-text("大客户定价")');

    // 等待表格加载
    await page.waitForSelector('table');

    // 获取表格行数（删除前）
    const rowsBefore = await page.locator('table tbody tr').count();

    // 点击第一个定价记录的删除按钮
    await page.click('table tbody tr:first-child button:has-text("删除")');

    // 验证确认对话框
    await expect(page.locator('.dialog')).toBeVisible();
    await expect(page.locator('.dialog')).toContainText('确认删除');

    // 确认删除
    await page.click('.dialog button:has-text("确定")');

    // 验证成功消息
    await expect(page.locator('.toast-success')).toBeVisible();

    // 验证表格行数减少
    await page.waitForSelector('table');
    const rowsAfter = await page.locator('table tbody tr').count();
    expect(rowsAfter).toBeLessThan(rowsBefore);
  });

  test('should filter pricing by customer name', async ({ page }) => {
    await page.goto('/admin/pricing');

    // 点击大客户定价标签
    await page.click('button:has-text("大客户定价")');

    // 等待表格加载
    await page.waitForSelector('table');

    // 输入搜索关键词
    await page.fill('input[placeholder*="搜索"]', 'Acme');

    // 验证表格过滤
    const rows = page.locator('table tbody tr');
    await rows.first().waitFor({ state: 'visible' });

    // 验证至少有一行（假设有匹配的数据）
    const count = await rows.count();
    if (count > 0) {
      const firstRowText = await rows.first().textContent();
      // 验证搜索结果包含关键词
      expect(firstRowText?.toLowerCase()).toContain('acme');
    }
  });
});
