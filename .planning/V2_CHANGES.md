# AI Gateway 项目 v2.0 调整总结

## 调整时间
2026-05-08

## 调整原因
根据用户反馈和项目需求，对技术栈和架构进行重要调整，以更好地满足项目目标。

---

## 主要调整内容

### 1. 技术栈调整

#### 后端调整
- **Go 版本**: `1.26.1` → `1.26.2`
  - 使用最新稳定版本
  - 性能和安全性改进

#### 前端调整
- **框架**: `React` → `Vue 3`
- **UI 组件库**: `Radix UI` → `shadcn-vue`
- **状态管理**: `Redux` → `Pinia`
- **构建工具**: `Vite` (保持不变)

**调整原因**:
- Vue 3 生态成熟，性能优秀
- shadcn-vue 提供现代化组件库
- 团队更熟悉 Vue 技术栈
- 开发效率更高

---

### 2. 服务架构调整

#### 原架构 (v1.0)
```
用户端 (React)  ─┐
运维端 (React)  ─┼─→ 业务服务 → Bifrost → AI 提供商
```

#### 新架构 (v2.0)
```
用户端 (Vue)    ─┐
运维端 (Vue)    ─┼─→ API 网关 (独立) → 业务服务 → Bifrost → AI 提供商
```

**关键变化**:
1. **API 网关独立部署**
   - 专门处理高并发 API 调用
   - 支持水平扩展
   - 目标: 10000+ QPS

2. **用户端和运维端完全分离**
   - 独立的前端应用
   - 独立的认证体系
   - 独立的数据访问

3. **业务服务微服务化**
   - 用户服务
   - 模型服务
   - 计费服务
   - 管理服务

---

### 3. 功能扩展

#### 模型接口扩展 (v1.0 → v2.0)

**v1.0**: 主要支持文本模型
- Chat Completions
- Completions

**v2.0**: 支持 8 种模型类型
1. **文本模型**
   - Chat Completions
   - Completions

2. **图片生成**
   - Images Generation
   - Images Edit
   - Images Variation

3. **视频生成**
   - Video Generation
   - 按秒计费

4. **语音合成 (TTS)**
   - Audio Speech
   - 多种声音支持

5. **语音识别 (STT)**
   - Audio Transcription
   - Audio Translation

6. **向量嵌入**
   - Embeddings
   - 不同维度支持

7. **重排序**
   - Reranking
   - 文档重排序

8. **其他功能**
   - Files (文件管理)
   - Batches (批量处理)
   - Moderation (内容审核)

**接口格式**: 全部兼容 OpenAI 格式

---

### 4. 价格体系完善

#### 基于 Bifrost 的价格体系

**v1.0**: 基础价格体系
- 输入/输出 token 价格
- 简单的按次计费

**v2.0**: 完善的价格体系 (参考 Bifrost)

##### 文本模型定价
```go
// 基础定价
InputCostPerToken          float64
OutputCostPerToken         float64

// 层级定价
InputCostPerTokenAbove128kTokens  *float64
InputCostPerTokenAbove200kTokens  *float64
InputCostPerTokenAbove272kTokens  *float64

// 模式定价
InputCostPerTokenBatches    *float64
InputCostPerTokenPriority  *float64

// 缓存定价
CacheCreationInputTokenCost *float64
CacheReadInputTokenCost     *float64
```

##### 图片模型定价
```go
// 按图片计费
OutputCostPerImage *float64

// 按分辨率定价
OutputCostPerImageAbove1024x1024Pixels *float64
OutputCostPerImageAbove2048x2048Pixels *float64

// 按质量定价
OutputCostPerImageLowQuality    *float64
OutputCostPerImageHighQuality   *float64
OutputCostPerImagePremiumImage  *float64
```

##### 视频/音频定价
```go
// 按秒计费
InputCostPerVideoPerSecond  *float64
OutputCostPerVideoPerSecond *float64
InputCostPerAudioPerSecond  *float64
OutputCostPerAudioPerSecond *float64

// 按 Token 计费
InputCostPerAudioToken  *float64
OutputCostPerAudioToken *float64
```

---

### 5. 用户和运维管理分离

#### v1.0
- 单一前端应用
- 用户和管理员混在一起

#### v2.0
- **用户端应用**: 面向普通用户
- **运维端应用**: 面向运维管理员
- 完全独立的认证体系
- 完全独立的界面设计

---

## 项目文件变化

### 更新的文档
1. **PROJECT.md** - 更新技术栈和架构
2. **REQUIREMENTS.md** - 扩展功能需求
3. **ROADMAP.md** - 调整开发计划
4. **STATE.md** - 更新项目状态
5. **QUICKSTART.md** - 更新启动指南
6. **config.json** - 更新配置信息

### 新增的文档
- **V2_CHANGES.md** - 本文档，记录 v2.0 变化

---

## 开发影响

### 时间影响
- **v1.0**: 10-12 周
- **v2.0**: 12-16 周
- **增加**: 2-4 周

**原因**:
- 服务分离架构更复杂
- 前端从 React 换成 Vue 需要学习
- 模型接口扩展需要更多开发时间

### 资源影响
- **v1.0**: 2-3 人
- **v2.0**: 3-4 人

**原因**:
- 双端前端开发工作量增加
- API 网关独立需要专门的优化

### 技术风险
- **新增风险**:
  - Vue 3 学习曲线
  - 服务分离架构复杂度
  - API 网关性能要求高

- **风险缓解**:
  - 团队培训和技术分享
  - 充分的架构设计评审
  - 早期性能测试

---

## 优势分析

### v2.0 相比 v1.0 的优势

#### 1. 性能提升
- API 网关独立部署，支持水平扩展
- 目标从 1000+ QPS 提升到 10000+ QPS

#### 2. 功能完善
- 从 2 种模型类型扩展到 8 种
- 完善的价格体系
- 更好的用户体验

#### 3. 可维护性
- 服务分离，职责清晰
- 用户端和运维端独立
- 便于团队协作

#### 4. 可扩展性
- API 网关可独立扩展
- 业务服务可独立部署
- 前端应用可独立维护

---

## 下一步行动

### 立即行动
1. ✅ 更新开发环境 (Go 1.26.2, Vue 3)
2. ⏳ 团队培训 (Vue 3, shadcn-vue)
3. ⏳ 技术预研 (服务分离架构)
4. ⏳ 性能测试准备

### 近期目标
1. ⏳ 完成 Bifrost 集成
2. ⏳ 设计服务分离架构
3. ⏳ 搭建 Vue 3 项目
4. ⏳ 实现 API 网关原型

---

## 版本对比表

| 特性 | v1.0 | v2.0 |
|------|------|------|
| Go 版本 | 1.26.1 | 1.26.2 |
| 前端框架 | React | Vue 3 |
| UI 组件 | Radix UI | shadcn-vue |
| 服务架构 | 单体应用 | 服务分离 |
| API 网关 | 集成 | 独立部署 |
| 模型类型 | 2 种 | 8 种 |
| 价格体系 | 基础 | 完善 (Bifrost) |
| 用户管理 | 混合 | 分离 |
| 开发周期 | 10-12 周 | 12-16 周 |
| 团队规模 | 2-3 人 | 3-4 人 |
| 目标 QPS | 1000+ | 10000+ |

---

## 总结

v2.0 版本在 v1.0 的基础上进行了重要调整，主要变化包括：

1. **技术栈升级**: Go 1.26.2, Vue 3, shadcn-vue
2. **架构重构**: 服务分离，API 网关独立
3. **功能扩展**: 8 种模型类型支持
4. **价格完善**: 基于 Bifrost 的完善价格体系
5. **管理分离**: 用户端和运维端独立

虽然开发周期增加了 2-4 周，但这些调整显著提升了项目的性能、可维护性和可扩展性，为长期发展奠定了更好的基础。

---

**调整完成时间**: 2026-05-08
**当前版本**: v2.0
**下一步**: 开始 Phase 1 开发
