/**
 * 表格列元素类型枚举
 */
export enum D2TableElemType {
  /** 普通文本（默认） */
  Text = 'text',
  /** 标签显示 */
  Tag = 'tag',
  /** 时间戳转换 */
  ConvertTime = 'convertTime',
  /** 枚举转描述 */
  EnumToDesc = 'enumToDesc',
  /** 下载链接（带 baseUrl 前缀） */
  DownloadWithSortUrl = 'downloadWithSortUrl',
  /** 复制链接 */
  CopyUrl = 'copyUrl',
  /** 跳转链接 */
  LinkJump = 'linkJump',
  /** 图片（带 baseUrl 前缀） */
  ImageWithSortUrl = 'imageWithSortUrl',
  /** 图片 */
  Image = 'image',
  /** 可编辑输入框 */
  EditInput = 'editInput',
  /** 字节转MB */
  Byte2MB = 'byte2MB',
  /** 下拉选择 */
  Select = 'select'
}

/**
 * 表格列配置
 */
export interface TableColumn {
  /** 字段名 */
  prop: string;
  /** 列标题 */
  label: string;
  /** 列宽度 */
  width?: string | number;
  /** 固定列 */
  fixed?: boolean | 'left' | 'right';
  /** 列类型 */
  type?: D2TableElemType;
  /** 枚举转描述映射（当 type 为 EnumToDesc 时使用） */
  enum2StrMap?: Record<string | number, string>;
  /** 下拉选项（当 type 为 Select 时使用） */
  options?: Array<{label: string; value: string | number}>;
  /** 是否可编辑（当 type 为 Image 时使用） */
  canEdit?: boolean;
}

/**
 * 抽屉列配置（用于详情/编辑抽屉）
 */
export interface DrawerColumn extends TableColumn {
  /** 是否必填 */
  required?: boolean;
}

