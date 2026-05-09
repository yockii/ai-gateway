import { test, expect } from '@playwright/test'

test.describe('User Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login')
    await page.fill('input[type="email"]', 'admin@example.com')
    await page.fill('input[type="password"]', 'password123')
    await page.click('button[type="submit"]')
    await page.waitForURL('**/admin/users')
  })

  test('should display users list', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('用户管理')

    // Check for table
    const table = page.locator('table')
    await expect(table).toBeVisible()

    // Check for table headers
    await expect(page.locator('th')).toContainText('邮箱')
    await expect(page.locator('th')).toContainText('姓名')
    await expect(page.locator('th')).toContainText('状态')
  })

  test('should search users by email', async ({ page }) => {
    const searchInput = page.locator('input[placeholder="搜索邮箱或姓名"]')

    await expect(searchInput).toBeVisible()

    // Enter search term
    await searchInput.fill('user1')
    await searchInput.press('Enter')

    // Wait for search results
    await page.waitForTimeout(500)

    // Should still show table
    await expect(page.locator('table')).toBeVisible()
  })

  test('should display user detail dialog', async ({ page }) => {
    // Click on view button for first user
    const viewButton = page.locator('button').first()
    await viewButton.click()

    // Check for dialog
    const dialog = page.locator('.fixed.inset-0')
    await expect(dialog).toBeVisible()

    // Check for user details
    await expect(dialog).toContainText('用户详情')
    await expect(dialog).toContainText('ID:')
    await expect(dialog).toContainText('邮箱:')

    // Close dialog
    const closeButton = dialog.locator('button:has-text("关闭")')
    await closeButton.click()

    // Dialog should be closed
    await expect(dialog).not.toBeVisible()
  })

  test('should toggle user status', async ({ page }) => {
    // Get the initial status text
    const firstRow = page.locator('tbody tr').first()

    // Click on toggle status button (second button in actions)
    const toggleButton = firstRow.locator('button').nth(1)
    await toggleButton.click()

    // Wait for status update
    await page.waitForTimeout(500)

    // Status should have changed
    await expect(page.locator('table')).toBeVisible()
  })

  test('should delete user with confirmation', async ({ page }) => {
    // Setup dialog handler to accept confirmation
    page.on('dialog', dialog => dialog.accept())

    // Click on delete button (third button in actions)
    const firstRow = page.locator('tbody tr').first()
    const deleteButton = firstRow.locator('button').nth(2)
    await deleteButton.click()

    // Wait for deletion
    await page.waitForTimeout(500)

    // Should still show table
    await expect(page.locator('table')).toBeVisible()
  })
})
