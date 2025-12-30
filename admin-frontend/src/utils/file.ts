/**
 * 文件 URL 工具函数
 * 用于统一处理文件 URL 的拼接逻辑
 */

/**
 * 拼接文件完整 URL
 * @param baseUrl 基础 URL（如 http://localhost:8888）
 * @param path 文件访问路径（相对路径，如 /uploads/xxx）
 * @param fallbackUrl 兼容字段：如果 baseUrl 和 path 都不存在，使用此 URL
 * @returns 完整 URL
 */
export function buildFileUrl(
  baseUrl?: string | null,
  path?: string | null,
  fallbackUrl?: string | null
): string {
  // 如果提供了完整 URL（兼容字段），且是完整 URL，直接返回
  if (fallbackUrl) {
    if (fallbackUrl.startsWith('http://') || fallbackUrl.startsWith('https://')) {
      return fallbackUrl;
    }
  }

  // 优先使用 baseUrl + path 拼接
  if (baseUrl && path) {
    const base = baseUrl.endsWith('/') ? baseUrl.slice(0, -1) : baseUrl;
    const filePath = path.startsWith('/') ? path : `/${path}`;
    return `${base}${filePath}`;
  }

  // 如果只有 path，使用前端环境变量中的 baseUrl 拼接
  if (path) {
    const frontendBaseUrl = import.meta.env.VITE_API_BASE_URL || '';
    if (frontendBaseUrl) {
      const base = frontendBaseUrl.endsWith('/') ? frontendBaseUrl.slice(0, -1) : frontendBaseUrl;
      const filePath = path.startsWith('/') ? path : `/${path}`;
      return `${base}${filePath}`;
    }
    // 如果没有前端 baseUrl，直接返回 path（相对路径）
    return path.startsWith('/') ? path : `/${path}`;
  }

  // 如果只有 fallbackUrl（相对路径），使用前端 baseUrl 拼接
  if (fallbackUrl) {
    const frontendBaseUrl = import.meta.env.VITE_API_BASE_URL || '';
    if (frontendBaseUrl) {
      const base = frontendBaseUrl.endsWith('/') ? frontendBaseUrl.slice(0, -1) : frontendBaseUrl;
      const filePath = fallbackUrl.startsWith('/') ? fallbackUrl : `/${fallbackUrl}`;
      return `${base}${filePath}`;
    }
    return fallbackUrl.startsWith('/') ? fallbackUrl : `/${fallbackUrl}`;
  }

  return '';
}

/**
 * 从 FileUploadResp 对象构建完整 URL
 * @param file 文件上传响应对象
 * @returns 完整 URL
 */
export function buildFileUrlFromResponse(file: {
  baseUrl?: string | null;
  path?: string | null;
  url?: string | null;
}): string {
  return buildFileUrl(file.baseUrl, file.path, file.url);
}

