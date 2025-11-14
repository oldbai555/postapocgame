using Godot;
using System;
using System.Collections.Generic;
using System.IO;
using System.Text;

namespace PostApocGame.Data
{
    /// <summary>
    /// 本地数据存储 - 用于存储账号信息、设置等
    /// </summary>
    public static class LocalStorage
    {
        private static string _dataPath;
        private static Dictionary<string, string> _cache = new Dictionary<string, string>();

        static LocalStorage()
        {
            // 使用Godot的用户数据目录
            _dataPath = OS.GetUserDataDir();
            if (!Directory.Exists(_dataPath))
            {
                Directory.CreateDirectory(_dataPath);
            }
        }

        /// <summary>
        /// 保存字符串数据
        /// </summary>
        public static void SaveString(string key, string value)
        {
            _cache[key] = value;
            string filePath = GetFilePath(key);
            try
            {
                File.WriteAllText(filePath, value, Encoding.UTF8);
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[LocalStorage] 保存数据失败: {key}, 错误: {ex.Message}");
            }
        }

        /// <summary>
        /// 读取字符串数据
        /// </summary>
        public static string LoadString(string key, string defaultValue = "")
        {
            // 先从缓存读取
            if (_cache.TryGetValue(key, out string cachedValue))
            {
                return cachedValue;
            }

            string filePath = GetFilePath(key);
            try
            {
                if (File.Exists(filePath))
                {
                    string value = File.ReadAllText(filePath, Encoding.UTF8);
                    _cache[key] = value;
                    return value;
                }
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[LocalStorage] 读取数据失败: {key}, 错误: {ex.Message}");
            }

            return defaultValue;
        }

        /// <summary>
        /// 保存账号信息（加密存储）
        /// </summary>
        public static void SaveAccount(string username, string password, bool rememberPassword)
        {
            SaveString("last_username", username);
            SaveString("remember_password", rememberPassword ? "1" : "0");
            
            if (rememberPassword)
            {
                // 简单加密（实际项目中应使用更安全的加密方式）
                string encrypted = SimpleEncrypt(password);
                SaveString("last_password", encrypted);
            }
            else
            {
                // 不记住密码时删除
                DeleteKey("last_password");
            }
        }

        /// <summary>
        /// 读取账号信息
        /// </summary>
        public static (string username, string password, bool rememberPassword) LoadAccount()
        {
            string username = LoadString("last_username", "");
            string rememberStr = LoadString("remember_password", "0");
            bool rememberPassword = rememberStr == "1";

            string password = "";
            if (rememberPassword)
            {
                string encrypted = LoadString("last_password", "");
                if (!string.IsNullOrEmpty(encrypted))
                {
                    password = SimpleDecrypt(encrypted);
                }
            }

            return (username, password, rememberPassword);
        }

        /// <summary>
        /// 删除键
        /// </summary>
        public static void DeleteKey(string key)
        {
            _cache.Remove(key);
            string filePath = GetFilePath(key);
            try
            {
                if (File.Exists(filePath))
                {
                    File.Delete(filePath);
                }
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[LocalStorage] 删除数据失败: {key}, 错误: {ex.Message}");
            }
        }

        /// <summary>
        /// 获取文件路径
        /// </summary>
        private static string GetFilePath(string key)
        {
            // 使用安全的文件名
            string safeKey = key.Replace("/", "_").Replace("\\", "_");
            return Path.Combine(_dataPath, $"{safeKey}.dat");
        }

        /// <summary>
        /// 简单加密（Base64编码，实际项目中应使用更安全的加密）
        /// </summary>
        private static string SimpleEncrypt(string plainText)
        {
            try
            {
                byte[] bytes = Encoding.UTF8.GetBytes(plainText);
                return Convert.ToBase64String(bytes);
            }
            catch
            {
                return plainText;
            }
        }

        /// <summary>
        /// 简单解密
        /// </summary>
        private static string SimpleDecrypt(string encryptedText)
        {
            try
            {
                byte[] bytes = Convert.FromBase64String(encryptedText);
                return Encoding.UTF8.GetString(bytes);
            }
            catch
            {
                return encryptedText;
            }
        }
    }
}

