# web - 前端应用

[根目录](../CLAUDE.md) > **web**

## 模块快照

**职责**：数据可视化展示、AG Grid 表格、站点配置管理

**技术栈**：TypeScript + Vite + AG Grid

**入口文件**：`src/main.ts`

## 项目结构

```
web/
├── src/
│   ├── main.ts              # 应用入口
│   ├── types.d.ts           # 类型定义
│   ├── gridConfig.ts        # 表格配置
│   ├── utils/
│   │   ├── index.ts         # 工具函数
│   │   ├── config.ts        # 站点配置加载
│   │   └── operationRender.ts  # 操作列渲染
├── index.html               # HTML 模板
├── vite.config.ts           # Vite 配置
├── package.json             # 依赖管理
└── eslint.config.js         # ESLint 配置
```

## 核心功能

1. **数据表格**（AG Grid）
   - 中文本地化
   - 客户端排序、筛选
   - 自定义列定义
   - 性能优化配置

2. **站点配置预加载**
   - 启动时加载站点配置
   - 失败时显示错误提示

3. **API 集成**
   - `/top1000.json` - Top1000 数据
   - `/sites.json` - 站点列表

## 开发命令

```bash
pnpm install    # 安装依赖
pnpm dev        # 开发服务器（代理后端 API）
pnpm build      # 生产构建
pnpm lint       # 代码检查
```

## Vite 配置

- **代码分割**：使用 `@xiaowaibuzheng/rolldown-vite-split-chunks`
- **输出结构**：`js/`, `css/`, `[ext]/` 分离
- **开发代理**：`/top1000.json` 和 `/sites.json` 代理到 `127.0.0.1:7066`
- **HTML 压缩**：`vite-plugin-html`

## 类型定义

```typescript
interface DataType {
    siteName: string      // 站点名称
    siteid: string        // 站点 ID
    duplication: string   // 重复度
    size: string          // 文件大小
    id: number            // 序号
}

interface ResDataType {
    items: DataType[]     // 站点列表
    time: string          // 更新时间
}
```

## 依赖版本

```json
{
  "node": ">=24.3.0",
  "pnpm": ">=10.12.4",
  "ag-grid-community": "^35.0.0",
  "typescript": "^5.9.3",
  "vite": "8.0.0-beta.5"
}
```

## 测试

无测试文件。

**建议**：添加工具函数单元测试、E2E 测试（Playwright/Cypress）。

---

*文档生成时间：2026-01-28 13:08:52*
