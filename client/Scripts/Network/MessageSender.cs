using Godot;
using System;
using Google.Protobuf;

namespace PostApocGame.Network
{
    /// <summary>
    /// 消息发送器 - 负责发送消息到服务器
    /// </summary>
    public static class MessageSender
    {
        /// <summary>
        /// 发送消息
        /// </summary>
        public static bool Send(IMessage message, int protocolId, string sessionId = "")
        {
            if (NetworkManager.Instance == null)
            {
                GD.PrintErr("[MessageSender] NetworkManager未初始化");
                return false;
            }

            byte[] data = ProtocolHandler.SerializeMessage(message, protocolId, sessionId);
            if (data == null)
            {
                GD.PrintErr($"[MessageSender] 消息序列化失败，协议ID: {protocolId}");
                return false;
            }

            // 获取协议名称和消息摘要
            string protocolName = ProtocolHandler.GetProtocolName(protocolId, true);
            string messageSummary = ProtocolHandler.GetMessageSummary(message);

            bool success = NetworkManager.Instance.SendMessage(data);
            if (success)
            {
                GD.Print($"[MessageSender] 📤 发送消息: {protocolName} (ID={protocolId}) | {messageSummary}");
            }

            return success;
        }

        /// <summary>
        /// 发送C2S协议消息（便捷方法）
        /// </summary>
        public static bool SendC2S(IMessage message, int c2sProtocolId)
        {
            return Send(message, c2sProtocolId);
        }
    }
}

