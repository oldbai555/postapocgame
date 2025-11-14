using Godot;
using System;

namespace PostApocGame.Utils
{
    /// <summary>
    /// 日志工具类
    /// </summary>
    public static class Logger
    {
        public enum LogLevel
        {
            Debug,
            Info,
            Warning,
            Error
        }

        private static LogLevel _minLevel = LogLevel.Debug;

        /// <summary>
        /// 设置最小日志级别
        /// </summary>
        public static void SetMinLevel(LogLevel level)
        {
            _minLevel = level;
        }

        /// <summary>
        /// 输出调试日志
        /// </summary>
        public static void Debug(string message, params object[] args)
        {
            if (_minLevel <= LogLevel.Debug)
            {
                string formatted = args.Length > 0 ? string.Format(message, args) : message;
                GD.Print($"[DEBUG] {formatted}");
            }
        }

        /// <summary>
        /// 输出信息日志
        /// </summary>
        public static void Info(string message, params object[] args)
        {
            if (_minLevel <= LogLevel.Info)
            {
                string formatted = args.Length > 0 ? string.Format(message, args) : message;
                GD.Print($"[INFO] {formatted}");
            }
        }

        /// <summary>
        /// 输出警告日志
        /// </summary>
        public static void Warning(string message, params object[] args)
        {
            if (_minLevel <= LogLevel.Warning)
            {
                string formatted = args.Length > 0 ? string.Format(message, args) : message;
                GD.PrintErr($"[WARNING] {formatted}");
            }
        }

        /// <summary>
        /// 输出错误日志
        /// </summary>
        public static void Error(string message, params object[] args)
        {
            if (_minLevel <= LogLevel.Error)
            {
                string formatted = args.Length > 0 ? string.Format(message, args) : message;
                GD.PrintErr($"[ERROR] {formatted}");
            }
        }
    }
}

