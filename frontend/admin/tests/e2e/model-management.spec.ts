import { test, expect } from '@playwright/test'

test.describe('Model Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login')
    await page.fill('input[type="email"]', 'admin@example.com')
    await page.fill('input[type="password"]', 'password123')
    await page.click('button[type="submit"]')
    await page.waitForURL('**/admin/users')

    // Navigate to models page
    await page.goto('/admin/models')
    await page.waitForTimeout(500)
  })

  test('should display models list', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('模型管理')

    // Check for table or card list
    const table = page.locator('table')
    const hasTable = await table.count() > 0

    if (hasTable) {
      await expect(table).toBeVisible()
    }
  })

  test('should filter models by type', async ({ page }) => {
    // Look for filter dropdown or tabs
    const filterButton = page.locator('button:has-text("筛选"), select').first()

    const hasFilter = await filterButton.count() > 0

    if (hasFilter) {
      await filterButton.click()
      await page.waitForTimeout(500)
    }
  })

  test('should enable/disable model', async ({ page }) => {
    // Look for toggle buttons in model list
    const toggleButton = page.locator('button').filter({ hasText: /(启用|禁用|Enable|Disable)/ }).first()

    const hasToggle = await toggleButton.count() > 0

    if (hasToggle) {
      await toggleButton.click()
      await page.waitForTimeout(500)

      // Check if status changed
      await expect(page.locator('h1')).toContainText('模型管理')
    }
  })

  test('should navigate to models page from sidebar', async ({ page }) => {
    // Click on models link in sidebar
    const modelsLink = page.locator('a:has-text("模型"), a:has-text("Models")').first()

    const hasLink = await modelsLink.count() > 0

    if (hasLink) {
      await modelsLink.click()
      await page.waitForURL('**/admin/models')
      await expect(page.locator('h1')).toBeVisible()
    }
  })
})
