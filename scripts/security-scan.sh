#!/bin/bash
set -e

echo "======================================"
echo "  AI Gateway Security Scan"
echo "======================================"
echo ""

# 创建结果目录
mkdir -p .security-reports
cd .security-reports

# GoSec 扫描
echo "1. Running GoSec..."
if command -v gosec &> /dev/null; then
    gosec -no-fail -fmt json -out gosec-report.json ../../... 2>&1 | tee gosec-output.txt || true
    gosec -severity medium -confidence medium -fmt text -out gosec-report.txt ../../... 2>&1 || true

    # 统计结果
    if [ -f gosec-report.json ]; then
        ISSUES=$(jq '.Issues | length' gosec-report.json 2>/dev/null || echo "N/A")
        echo "   GoSec found $ISSUES issues"
    fi
else
    echo "   GoSec not found, skipping..."
fi

echo ""

# GoVulnCheck 检查已知漏洞
echo "2. Running GoVulnCheck..."
if command -v govulncheck &> /dev/null; then
    govulncheck -json ../../... 2>&1 | tee vulncheck-output.json || true
    govulncheck ../../... 2>&1 | tee vulncheck-report.txt || true

    # 检查是否有漏洞
    if grep -q "Vulnerabilities found" vulncheck-report.txt 2>/dev/null; then
        echo "   ⚠️  Vulnerabilities detected!"
    else
        echo "   No known vulnerabilities found"
    fi
else
    echo "   govulncheck not found, skipping..."
fi

echo ""

# Golangci-lint
echo "3. Running golangci-lint..."
if command -v golangci-lint &> /dev/null; then
    golangci-lint run --timeout 5m --out-format json ../../... 2>&1 | tee golangci-lint-report.json || true
    golangci-lint run --timeout 5m ../../... 2>&1 | tee golangci-lint-report.txt || true
else
    echo "   golangci-lint not found, skipping..."
fi

echo ""

# 检查依赖
echo "4. Checking dependencies..."
if [ -f ../../go.mod ]; then
    echo "   Checking for outdated dependencies..."
    cd ../..
    go list -u -m all 2>&1 | head -20
    cd .security-reports
fi

echo ""

# 检查敏感信息
echo "5. Checking for sensitive information..."
cd ../..
if command -v trufflehog &> /dev/null; then
    trufflehog filesystem --directory . --output .security-reports/trufflehog-report.json || true
    echo "   TruffleHog scan complete"
else
    echo "   trufflehog not found, skipping secret detection..."
fi

# 简单的敏感信息检查
echo "   Checking for common secrets..."
GREP_RESULTS=0
if grep -r -i -E "password\s*=\s*['\"][^'\"]+['\"]" --include="*.go" --include="*.yaml" --include="*.yml" --include="*.json" . 2>/dev/null | grep -v "test" | grep -v "example" > .security-reports/password-check.txt; then
    if [ -s .security-reports/password-check.txt ]; then
        echo "   ⚠️  Possible hardcoded passwords found!"
        GREP_RESULTS=1
    fi
fi

if [ $GREP_RESULTS -eq 0 ]; then
    echo "   No obvious hardcoded secrets found"
fi

echo ""

# 生成摘要报告
echo "6. Generating summary report..."
cd .security-reports
cat > SECURITY_SUMMARY.md << 'EOF'
# Security Scan Summary

Generated: $(date)

## Scan Results

### GoSec
- Report: `gosec-report.json`
- Output: `gosec-output.txt`

### GoVulnCheck
- Report: `vulncheck-report.txt`
- JSON: `vulncheck-output.json`

### GolangCI-Lint
- Report: `golangci-lint-report.txt`
- JSON: `golangci-lint-report.json`

### Secret Detection
- Password Check: `password-check.txt`
- TruffleHog: `trufflehog-report.json` (if available)

## Recommendations

1. Review all high-severity issues found by GoSec
2. Address any known vulnerabilities identified by GoVulnCheck
3. Fix linting issues that may indicate security problems
4. Rotate any hardcoded credentials found
5. Keep dependencies up to date

## Next Steps

- Review individual reports for detailed findings
- Create issues for critical security problems
- Schedule regular security scans
- Consider integrating with GitHub Security Alerts
EOF

echo ""
echo "======================================"
echo "  Security Scan Complete"
echo "======================================"
echo ""
echo "Reports saved to: .security-reports/"
echo "  - SECURITY_SUMMARY.md"
echo "  - gosec-report.json"
echo "  - vulncheck-report.txt"
echo "  - golangci-lint-report.txt"
echo ""
echo "View summary: cat .security-reports/SECURITY_SUMMARY.md"
