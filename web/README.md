# Top1000 前端应用

> 使用TypeScript + Vite + AG Grid开发的表格界面

---

## 系统简介

前端页面，主要功能：**展示Top1000的PT资源数据**

技术栈：
- **Vite** - 构建工具
- **TypeScript** - 类型安全
- **AG Grid企业版** - 表格组件，功能完整
- **Vite插件** - HTML模板处理

---

## 启动指南

### 安装依赖

```bash
pnpm install
```

推荐使用pnpm，npm和yarn也可以使用。

### 启动开发服务器

```bash
pnpm dev
```

打开浏览器访问控制台显示的地址（通常是`http://localhost:5173`）。

### 构建生产版本

```bash
pnpm build
```

构建产物会输出到`../web-dist/`目录。

### 预览生产版本

```bash
pnpm preview
```

---

## 目录结构

```
web/
├── src/
│   ├── main.ts              # 入口文件（AG Grid配置）
│   ├── types.d.ts           # TypeScript类型定义
│   ├── utils/
│   │   ├── index.ts         # 工具函数（fetch、大小转换）
│   │   ├── operationRender.ts  # 操作列渲染器
│   │   ├── config.ts        # 站点URL配置
│   │   └── iyuuSites.ts     # 站点元数据（118个）
│   └── index.css            # 全局样式
├── index.html               # HTML模板
├── package.json             # npm依赖
├── pnpm-lock.yaml           # pnpm锁定文件
├── vite.config.ts           # Vite构建配置
└── eslint.config.js         # ESLint配置
```

---

## 核心功能

### 1. AG Grid表格

- **列过滤**：支持名字列过滤
- **列排序**：重复度、文件大小可排序
- **操作列**：点击"详情"或"下载种子"跳转
- **中文界面**：本地化配置

### 2. 数据获取

```typescript
const response = await fetch('/top1000.json')
const json: ResDataType = await response.json()
```

后端API：`http://localhost:7066/top1000.json`

### 3. 大小排序

自定义比较器，支持KB/MB/GB/TB单位：

```typescript
comparator: (valueA, valueB) => {
  const kbA = convertSizeToKb(valueA)
  const kbB = convertSizeToKb(valueB)
  return kbA - kbB
}
```

---

## 构建配置

### Vite插件

```typescript
export default defineConfig({
  plugins: [
    splitChunks(),                    // 代码分割
    createHtmlPlugin({ minify: true }), // HTML压缩
  ],
  build: {
    rollupOptions: {
      output: {
        chunkFileNames: 'js/[name]-[hash].js',
        entryFileNames: 'js/[name]-[hash].js',
        assetFileNames: '[ext]/[name]-[hash].[ext]',
      },
    },
    outDir: resolve(__dirname, '../web-dist'),
  },
})
```

### 输出结构

```
../web-dist/
├── index.html        # 压缩后的HTML
├── js/               # JS文件（带hash）
└── css/              # CSS文件（带hash）
```

---

## 环境要求

- **Node.js**: >=24.3.0
- **pnpm**: >=10.12.4

---

## 常见问题

### Q: 如何修改数据源地址？

**A**: 修改`src/utils/index.ts`：
```typescript
const response = await fetch('http://localhost:7066/top1000.json')
// 修改为实际地址
const response = await fetch('https://your-domain.com/top1000.json')
```

### Q: 如何添加新列？

**A**: 修改`src/main.ts`的`columnDefs`：
```typescript
columnDefs: [
  {
    headerName: '名字',
    field: 'siteName',
    filter: true,
  },
  {
    headerName: '新列',      // 新增列
    field: 'newField',
    sortable: true,
  },
]
```

### Q: 如何导出Excel？

**A**: AG Grid企业版内置支持：
```typescript
gridApi.exportDataAsExcel({
  fileName: 'top1000.xlsx',
})
```

---

**总结**：前端就一个页面，使用AG Grid企业版，功能完整。未编写测试。

**更新**: 2026-01-10
**代码质量**: B级（没测试）
