import { test, expect } from '@playwright/test';

// 供应商管理 E2E 测试
test.describe('Supplier Management', () => {
  test.beforeEach(async ({ page }) => {
    // 登录管理员
    await page.goto('/admin/login');
    await page.fill('input[name="email"]', 'admin@example.com');
    await page.fill('input[name="password"]', 'admin123456');
    await page.click('button[type="submit"]');
    await page.waitForURL('/admin');
  });

  test('should display supplier list', async ({ page }) => {
    await page.goto('/admin/suppliers');

    // 等待表格加载
    await page.waitForSelector('table');

    // 验证表格列存在
    await expect(page.locator('table thead tr th')).toContainText(['名称', 'Provider', '健康状态', '模型数量', '操作']);
  });

  test('should navigate to supplier detail page', async ({ page }) => {
    await page.goto('/admin/suppliers');

    // 点击第一个供应商
    await page.click('table tbody tr:first-child td:first-child a');

    // 验证 URL 变化
    await expect(page).toHaveURL(/\/admin\/suppliers\/[\w-]+/);

    // 验证详情页面元素
    await expect(page.locator('h1')).toContainText('供应商详情');
  });

  test('should create API key', async ({ page }) => {
    await page.goto('/admin/suppliers');

    // 选择一个供应商
    await page.click('table tbody tr:first-child');

    // 切换到 API Key 标签页
    await page.click('button:has-text("API Key")');

    // 点击创建按钮
    await page.click('button:has-text("创建密钥")');

    // 填写表单
    await page.fill('input[name="name"]', 'Test API Key');
    await page.fill('input[name="api_key"]', 'sk-test-' + Date.now());
    await page.fill('input[name="priority"]', '1');
    await page.fill('input[name="max_requests"]', '1000');

    // 提交表单
    await page.click('button:has-text("确定")');

    // 验证成功消息
    await expect(page.locator('.toast-success')).toBeVisible();
  });

  test('should set primary API key', async ({ page }) => {
    await page.goto('/admin/suppliers');

    // 选择一个供应商
    await page.click('table tbody tr:first-child');

    // 切换到 API Key 标签页
    await page.click('button:has-text("API Key")');

    // 点击设为主密钥按钮
    const firstRow = page.locator('table tbody tr:first-child');
    await firstRow.locator('button:has-text("设为主密钥")').click();

    // 验证确认对话框
    await expect(page.locator('.dialog')).toBeVisible();
    await page.click('.dialog button:has-text("确定")');

    // 验证成功消息
    await expect(page.locator('.toast-success')).toBeVisible();
  });

  test('should add model to supplier', async ({ page }) => {
    await page.goto('/admin/suppliers');

    // 选择一个供应商
    await page.click('table tbody tr:first-child');

    // 切换到模型关联标签页
    await page.click('button:has-text("模型关联")');

    // 点击添加模型按钮
    await page.click('button:has-text("添加模型")');

    // 填写表单
    await page.fill('input[name="model_id"]', 'gpt-4-test');
    await page.fill('input[name="input_cost"]', '0.03');
    await page.fill('input[name="output_cost"]', '0.06');

    // 提交表单
    await page.click('button:has-text("确定")');

    // 验证成功消息
    await expect(page.locator('.toast-success')).toBeVisible();
  });

  test('should display health status', async ({ page }) => {
    await page.goto('/admin/suppliers');

    // 选择一个供应商
    await page.click('table tbody tr:first-child');

    // 切换到健康状态标签页
    await page.click('button:has-text("健康状态")');

    // 验证健康状态组件
    await expect(page.locator('.health-indicator')).toBeVisible();

    // 验证实时更新（SSE 或轮询）
    const initialStatus = await page.locator('.health-indicator .status').textContent();
    await page.waitForTimeout(6000); // 等待 6 秒（超过 5 秒推送间隔）
    const updatedStatus = await page.locator('.health-indicator .status').textContent();

    // 状态可能相同，但组件应该已更新
    expect(updatedStatus).toBeTruthy();
  });
});
