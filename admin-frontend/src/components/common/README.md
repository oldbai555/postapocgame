# D2Table 表格组件

## 简介

D2Table 是一个功能丰富的表格组件，基于 Element Plus 的 `el-table` 封装，提供了统一的表格展示、分页、详情/编辑抽屉、新增抽屉等功能。

## 特性

- ✅ 统一的表格展示和分页
- ✅ 支持多种列类型（时间戳、标签、枚举、图片、链接等）
- ✅ 内置详情/编辑抽屉
- ✅ 内置新增抽屉
- ✅ 支持自定义列渲染（通过插槽）
- ✅ TypeScript 类型支持
- ✅ 国际化支持

## 基础用法

```vue
<template>
  <D2Table
    :columns="columns"
    :data="list"
    :total="total"
    :page-size="query.pageSize"
    :current-page="query.page"
    :drawer-columns="drawerColumns"
    :drawer-add-columns="drawerAddColumns"
    @size-change="handleSizeChange"
    @current-change="handlePageChange"
    @onclick-delete="handleDelete"
    @onclick-update-row="handleUpdate"
    @onclick-add-row="handleAdd"
  />
</template>

<script setup lang="ts">
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';

const columns: TableColumn[] = [
  {prop: 'id', label: 'ID', width: 80},
  {prop: 'name', label: '名称'},
  {prop: 'status', label: '状态', width: 100}
];

const drawerColumns: DrawerColumn[] = [
  {prop: 'id', label: 'ID', type: D2TableElemType.Tag},
  {prop: 'name', label: '名称', type: D2TableElemType.EditInput, required: true},
  {prop: 'status', label: '状态', type: D2TableElemType.Tag}
];

const drawerAddColumns: DrawerColumn[] = [
  {prop: 'name', label: '名称', required: true},
  {prop: 'status', label: '状态', type: D2TableElemType.Select, options: [
    {label: '启用', value: 1},
    {label: '禁用', value: 0}
  ]}
];
</script>
```

## Props

| 参数 | 说明 | 类型 | 默认值 | 必填 |
|------|------|------|--------|------|
| columns | 表格列配置 | `TableColumn[]` | - | 是 |
| data | 表格数据 | `any[]` | - | 是 |
| total | 总条数 | `number` | - | 是 |
| pageSize | 每页显示条数 | `number` | `10` | 否 |
| currentPage | 当前页码 | `number` | `1` | 否 |
| baseUrl | 基础URL（用于文件下载、图片显示等） | `string` | `''` | 否 |
| haveEdit | 是否显示编辑按钮 | `boolean` | `true` | 否 |
| haveDetail | 是否显示查看按钮 | `boolean` | `true` | 否 |
| drawerColumns | 详情/编辑抽屉列配置 | `DrawerColumn[]` | - | 是 |
| drawerAddColumns | 新增抽屉列配置 | `DrawerColumn[]` | - | 否 |
| havCustomBtn | 是否显示自定义按钮 | `boolean` | `false` | 否 |
| havCustomStr | 自定义按钮文本 | `string` | `'自定义按钮'` | 否 |
| maxHeight | 表格最大高度 | `string \| number` | `600` | 否 |
| actionColumnWidth | 操作列宽度 | `number` | `220` | 否 |
| drawerWidth | 抽屉宽度 | `string \| number` | `'50%'` | 否 |
| pageSizes | 分页每页条数选项 | `number[]` | `[10, 20, 50, 100]` | 否 |
| paginationLayout | 分页布局 | `string` | `'total, sizes, prev, pager, next, jumper'` | 否 |

## Events

| 事件名 | 说明 | 参数 |
|--------|------|------|
| size-change | 每页条数改变时触发 | `(size: number)` |
| current-change | 当前页改变时触发 | `(page: number)` |
| onclick-delete | 点击删除按钮时触发 | `(index: number, row: any)` |
| onclick-update-row | 点击保存（编辑）时触发 | `(row: any)` |
| onclick-add-row | 点击新增时触发 | `(row: any)` |
| onclick-btn-custom | 点击自定义按钮时触发 | `(index: number, row: any)` |

## Slots

| 插槽名 | 说明 | 作用域参数 |
|--------|------|-----------|
| cell | 自定义单元格内容 | `{row, column, index}` |

## 列类型（D2TableElemType）

| 类型 | 说明 | 适用场景 |
|------|------|----------|
| `Text` | 普通文本（默认） | 普通文本显示 |
| `Tag` | 标签显示 | 使用 el-tag 显示 |
| `ConvertTime` | 时间戳转换 | 将秒级时间戳转换为日期时间字符串 |
| `EnumToDesc` | 枚举转描述 | 通过 enum2StrMap 映射显示描述 |
| `DownloadWithSortUrl` | 下载链接 | 带 baseUrl 前缀的下载链接 |
| `CopyUrl` | 复制链接 | 点击复制链接到剪贴板 |
| `LinkJump` | 跳转链接 | 跳转到指定URL |
| `ImageWithSortUrl` | 图片（带 baseUrl） | 带 baseUrl 前缀的图片显示 |
| `Image` | 图片 | 直接显示图片 |
| `EditInput` | 可编辑输入框 | 在抽屉中可编辑的输入框 |
| `Byte2MB` | 字节转MB | 将字节数转换为MB显示 |
| `Select` | 下拉选择 | 在新增抽屉中使用下拉选择 |

## 自定义列渲染示例

```vue
<template>
  <D2Table
    :columns="columns"
    :data="list"
    :total="total"
    @onclick-delete="handleDelete"
  >
    <!-- 自定义状态列 -->
    <template #cell="{row, column}">
      <el-tag v-if="column.prop === 'status'" :type="row.status === 1 ? 'success' : 'info'">
        {{ row.status === 1 ? '启用' : '禁用' }}
      </el-tag>
    </template>
  </D2Table>
</template>
```

## 注意事项

1. **权限控制**：组件内部的操作按钮不包含权限控制，如需权限控制，可以通过隐藏操作列（设置 `haveEdit`、`haveDetail` 为 `false`），然后使用自定义列来实现。

2. **时间戳格式**：`ConvertTime` 类型会将秒级时间戳转换为 `YYYY-MM-DD HH:mm:ss` 格式。

3. **图片上传**：图片上传功能需要配置 `baseUrl`，并且后端需要支持文件上传接口 `/upload`。

4. **枚举映射**：使用 `EnumToDesc` 类型时，需要在列配置中提供 `enum2StrMap` 对象。

5. **抽屉表单验证**：组件不包含表单验证逻辑，需要在事件处理函数中自行实现。

## 完整示例

参考 `src/views/system/RoleList.vue` 和 `src/views/system/PermissionList.vue`。

