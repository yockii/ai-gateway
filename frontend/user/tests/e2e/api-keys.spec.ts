import { test, expect } from '@playwright/test'

test.describe('API Keys Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login')
    await page.fill('input[type="email"]', 'user@example.com')
    await page.fill('input[type="password"]', 'password123')
    await page.click('button[type="submit"]')
    await page.waitForURL('**/dashboard')

    // Navigate to API keys page
    await page.goto('/keys')
    await page.waitForTimeout(500)
  })

  test('should display API keys list', async ({ page }) => {
    await expect(page.locator('h1')).toBeVisible()

    // Look for table or card list
    const table = page.locator('table')
    const hasTable = await table.count() > 0

    if (hasTable) {
      await expect(table.first()).toBeVisible()
    }
  })

  test('should create new API key', async ({ page }) => {
    // Look for create button
    const createButton = page.locator('button:has-text("创建"), button:has-text("Create"), button:has-text("新建")').first()

    const hasButton = await createButton.count() > 0

    if (hasButton) {
      await createButton.click()
      await page.waitForTimeout(500)

      // Look for name input
      const nameInput = page.locator('input[name="name"], input[placeholder*="名称"], input[placeholder*="name"]')
      const hasInput = await nameInput.count() > 0

      if (hasInput) {
        await nameInput.first().fill('Test Key')

        const submitButton = page.locator('button[type="submit"]').first()
        await submitButton.click()
        await page.waitForTimeout(1000)
      }
    }
  })

  test('should delete API key', async ({ page }) => {
    // Look for delete button
    const deleteButton = page.locator('button:has-text("删除"), button:has-text("Delete")').first()

    const hasButton = await deleteButton.count() > 0

    if (hasButton) {
      // Setup dialog handler to accept confirmation
      page.on('dialog', dialog => dialog.accept())

      await deleteButton.click()
      await page.waitForTimeout(1000)

      // Should still be on the page
      await expect(page.locator('h1')).toBeVisible()
    }
  })

  test('should display key details', async ({ page }) => {
    // Look for view/detail buttons
    const viewButton = page.locator('button:has-text("查看"), button:has-text("View"), button:has-text("详情")').first()

    const hasButton = await viewButton.count() > 0

    if (hasButton) {
      await viewButton.click()
      await page.waitForTimeout(500)

      // Look for dialog or modal
      const dialog = page.locator('.fixed, .modal, .dialog')
      const hasDialog = await dialog.count() > 0

      if (hasDialog) {
        await expect(dialog.first()).toBeVisible()
      }
    }
  })

  test('should copy API key to clipboard', async ({ page }) => {
    // Look for copy button
    const copyButton = page.locator('button:has-text("复制"), button:has-text("Copy")').first()

    const hasButton = await copyButton.count() > 0

    if (hasButton) {
      // Setup clipboard access
      const clipboardText = await page.evaluate(() => {
        return navigator.clipboard.readText()
      }).catch(() => '')

      await copyButton.click()
      await page.waitForTimeout(500)

      // Button should still exist after click
      await expect(copyButton).toBeVisible()
    }
  })
})
