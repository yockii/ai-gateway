# Phase 2: 核心功能开发 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-08
**Phase:** 2-核心功能开发
**Areas discussed:** Bifrost 集成, 价格体系, 模型与 Key 管理, 会员套餐, 账单系统

---

## Bifrost 集成

### 集成方式
| Option | Description | Selected |
|--------|-------------|----------|
| 作为库集成 | 将 Bifrost 作为 Go 库直接集成，代码在同一进程。性能最优，但耦合度高。 | ✓ |
| 服务化调用 | Bifrost 独立运行服务，通过 HTTP/gRPC 调用。解耦但增加延迟。 | |
| 混合模式 | 核心功能用库集成，治理功能（虚拟密钥、预算）用服务调用。平衡性能和灵活性。 | |

**User's choice:** 作为库集成

### 配置管理
| Option | Description | Selected |
|--------|-------------|----------|
| 复用 Bifrost 配置 | 使用 Bifrost 原生的 provider 和 model 配置系统。升级简单但与现有模型可能有冗余。 | |
| 封装现有模型 | 将现有的 Supplier/ExternalModel 映射到 Bifrost 配置。保持数据模型统一，但需要适配层。 | |
| 混合管理 | Bifrost 管理供应商和路由，外部模型管理对外展示和定价。职责分离。 | ✓ |

**User's choice:** 让 Claude 推荐
**Notes:** 推荐混合管理方案 — Bifrost 专注底层调用，业务层专注定价和利润管理

### 接口兼容范围
| Option | Description | Selected |
|--------|-------------|----------|
| 全接口兼容 | 实现所有 OpenAI 接口（Chat/Completions/Images/Audio/Embeddings 等）。工作量大但完整。 | ✓ |
| 核心接口优先 | 先实现 Chat/Completions/Embeddings，其他接口后续迭代。快速上线核心功能。 | |
| 按需求实现 | 根据 REQUIREMENTS.md 中的 FR-001，实现 8 种模型类型接口。平衡范围和需求。 | |

**User's choice:** 全接口兼容

### 流式响应
| Option | Description | Selected |
|--------|-------------|----------|
| SSE 流式 | 使用 Server-Sent Events (SSE) 实现流式响应。OpenAI 标准，用户体验好。 | |
| 非流式优先 | 先实现非流式，后续迭代添加流式支持。降低初期复杂度。 | |
| 双模式支持 | 同时支持流式和非流式，通过 stream 参数控制。完整实现但工作量大。 | ✓ |

**User's choice:** 双模式支持

### 故障转移策略
| Option | Description | Selected |
|--------|-------------|----------|
| 双重保障 | Bifrost 原生故障转移 + 业务层重试。两层保障但可能过度设计。 | |
| Bifrost 原生 | 使用 Bifrost 的故障转移机制，业务层只处理最终失败。简单有效。 | ✓ |
| 业务层控制 | 业务层完全控制路由和重试，不依赖 Bifrost 故障转移。最大控制权。 | |

**User's choice:** 让 Claude 推荐
**Notes:** 推荐协同方案 — 业务层智能路由（利润排序）+ Bifrost 故障转移

### 性能监控
| Option | Description | Selected |
|--------|-------------|----------|
| 全面监控 | 基础指标（QPS、延迟、错误率）+ 业务指标（利润率、供应商成本）。全面但工作量大。 | ✓ |
| 核心指标 | 先实现基础性能指标，业务指标后续迭代。快速上线核心功能。 | |
| 混合监控 | 使用 Bifrost 原生监控 + 自定义业务指标。平衡工作量和需求。 | |

**User's choice:** 全面监控

---

## 价格体系

### Bifrost 价格集成
| Option | Description | Selected |
|--------|-------------|----------|
| 直接使用 | 直接使用 Bifrost 的定价结构和模型。简单但可能不完全匹配业务需求。 | |
| 重新实现 | 参考 Bifrost 结构，用现有模型重新实现。灵活但工作量大。 | |
| 混合集成 | 核心定价逻辑用 Bifrost，业务层做包装和扩展。平衡复用和定制。 | ✓ |

**User's choice:** 让 Claude 推荐
**Notes:** 推荐参考 Bifrost 结构 + 业务层利润管理

### 计费触发时机
| Option | Description | Selected |
|--------|-------------|----------|
| 异步记录 | API 响应后异步记录。性能最优，但可能有少量数据丢失。 | |
| 同步记录 | API 响应前同步记录。数据一致性好，但增加响应延迟。 | |
| 混合模式 | 流式响应在首字节返回时记录，非流式在响应后记录。平衡方案。 | |

**User's choice:** 用户强调不能丢失数据
**Notes:** 用户提出消息队列方案，经过讨论确定双重记录保障

### 消息队列方案
| Option | Description | Selected |
|--------|-------------|----------|
| Redis Stream | 使用 Redis Stream。轻量级，如果项目已有 Redis 可复用，支持持久化和消费者组。 | ✓ |
| RabbitMQ | 使用 RabbitMQ。成熟可靠，企业级消息队列，支持消息确认和持久化。 | |
| NATS | 使用 NATS/JetStream。高性能轻量级，支持消息持久化和重试。 | |

**User's choice:** 提出更优方案
**Notes:** Go 协程异步直接写 DB + 同步推送 Redis Stream + 消费者协程处理 Stream，RequestID 幂等去重

---

## 模型与 Key 管理

### 模型映射方式
| Option | Description | Selected |
|--------|-------------|----------|
| 一对多 | 一个对外模型映射到多个供应商的实际模型。支持成本优先路由。 | ✓ |
| 一对一 | 一个对外模型对应一个供应商模型。简单但不够灵活。 | |
| 多对多 | 复杂的映射关系，支持条件路由。灵活但复杂度高。 | |

**User's choice:** 一对多

### API Key 生成策略
| Option | Description | Selected |
|--------|-------------|----------|
| UUID 标准 | 使用 UUID v4 生成 Key，前缀 'sk-' 开头。OpenAI 标准，易于识别。 | ✓ |
| 安全哈希 | 使用哈希算法生成随机 Key，更安全但不可逆。 | |
| 分段格式 | 使用分段格式（如 'sk-live-xxxx'），便于管理和识别。 | |

**User's choice:** UUID 标准

### API Key 权限模型
| Option | Description | Selected |
|--------|-------------|----------|
| 完整模型 | 完整权限模型：支持模型白名单、IP 白名单、额度限制、过期时间。功能全面但复杂。 | |
| 基础模型 | 基础权限：仅支持额度限制和过期时间。简单够用。 | |
| 标签系统 | 灵活标签：通过标签系统定义权限（如 models=gpt-4,gpt-3.5）。灵活可扩展。 | |

**User's choice:** 详细需求
**Notes:** 需要支持额度限制、并发限制（按模型配置，带默认值）、过期时间

---

## 会员套餐

### 套餐体系设计
| Option | Description | Selected |
|--------|-------------|----------|
| 等级制 | 普通/高级/VIP 等级，每个等级固定权益。简单清晰。 | ✓ |
| 套餐制 | 多个套餐可选，每个套餐独立配置。灵活但复杂。 | |
| 混合制 | 基础等级 + 可选增值套餐。平衡灵活性和复杂度。 | |

**User's choice:** 等级制

### 定价折扣设计
| Option | Description | Selected |
|--------|-------------|----------|
| 固定折扣 | 每个等级有固定折扣（如 VIP 8 折）。简单直接。 | |
| 阶梯定价 | 使用量越大折扣越高。鼓励使用但计算复杂。 | |
| 差异化 | 不同模型不同折扣（如基础模型大折扣，高级模型小折扣）。精细管理。 | ✓ |

**User's choice:** 差异化

### 升级降级处理
| Option | Description | Selected |
|--------|-------------|----------|
| 即时生效 | 升级/降级立即应用。用户体验好但可能产生计费复杂性。 | ✓ |
| 下期生效 | 升级/下个计费周期生效。简单但用户体验差。 | |
| 混合模式 | 升级即时生效，降级下期生效。平衡用户体验和合理性。 | |

**User's choice:** 即时生效

---

## 账单系统

### 账单生成时机
| Option | Description | Selected |
|--------|-------------|----------|
| 自动生成 | 每月 1 号自动生成上月账单。自动化程度高。 | ✓ |
| 按需生成 | 用户手动触发账单生成。灵活但可能遗漏。 | |
| 混合模式 | 自动生成 + 支持手动重新生成。兼顾自动化和灵活性。 | |

**User's choice:** 自动生成

### 导出格式
| Option | Description | Selected |
|--------|-------------|----------|
| CSV | CSV 格式：简单通用，Excel 可打开。基础需求。 | |
| Excel | Excel 格式：支持多 Sheet、格式化。功能丰富但实现复杂。 | |
| 多格式 | CSV + Excel + JSON。满足不同用户需求。 | |

**User's choice:** 创新方案
**Notes:** 总结性账单 PDF（正式、可打印）+ 明细表 CSV（便于数据分析）

---

## Claude's Discretion

无 — 用户在所有决策中都提供了明确的选择或要求推荐。

## Deferred Ideas

None — discussion stayed within phase scope.
