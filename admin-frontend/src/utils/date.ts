/**
 * 日期时间格式化工具
 */

/**
 * 格式化时间戳为日期时间字符串
 * @param timestamp 时间戳（毫秒）
 * @param format 格式化模板，默认 'YYYY-MM-DD HH:mm:ss'
 */
export function formatDateTime(timestamp: number, format = 'YYYY-MM-DD HH:mm:ss'): string {
  if (!timestamp) return '';
  
  const date = new Date(timestamp);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  const seconds = String(date.getSeconds()).padStart(2, '0');
  
  return format
    .replace('YYYY', String(year))
    .replace('MM', month)
    .replace('DD', day)
    .replace('HH', hours)
    .replace('mm', minutes)
    .replace('ss', seconds);
}

/**
 * 格式化时间戳（秒级）为日期时间字符串
 */
export function formatUnixTime(timestamp: number, format = 'YYYY-MM-DD HH:mm:ss'): string {
  return formatDateTime(timestamp * 1000, format);
}

