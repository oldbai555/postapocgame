using Godot;
using System;
using System.Collections.Generic;
using Google.Protobuf;
using Pb3;

namespace PostApocGame.Network
{
    /// <summary>
    /// 协议处理器 - 负责协议的序列化和反序列化
    /// </summary>
    public static class ProtocolHandler
    {
        // 协议ID到消息类型的映射
        private static Dictionary<int, Type> _c2sProtocolMap = new Dictionary<int, Type>();
        private static Dictionary<int, Type> _s2cProtocolMap = new Dictionary<int, Type>();

        static ProtocolHandler()
        {
            InitializeProtocolMaps();
        }

        /// <summary>
        /// 初始化协议映射表
        /// </summary>
        private static void InitializeProtocolMaps()
        {
            // C2S协议注册
            RegisterC2SProtocol((int)C2SProtocol.C2Sregister, typeof(C2SRegisterReq));
            RegisterC2SProtocol((int)C2SProtocol.C2Slogin, typeof(C2SLoginReq));
            RegisterC2SProtocol((int)C2SProtocol.C2SqueryRoles, typeof(C2SQueryRolesReq));
            RegisterC2SProtocol((int)C2SProtocol.C2ScreateRole, typeof(C2SCreateRoleReq));
            RegisterC2SProtocol((int)C2SProtocol.C2SenterGame, typeof(C2SEnterGameReq));
            RegisterC2SProtocol((int)C2SProtocol.C2Sreconnect, typeof(C2SReconnectReq));

            // S2C协议注册
            RegisterS2CProtocol((int)S2CProtocol.S2CregisterResult, typeof(S2CRegisterResultReq));
            RegisterS2CProtocol((int)S2CProtocol.S2CloginResult, typeof(S2CLoginResultReq));
            RegisterS2CProtocol((int)S2CProtocol.S2CroleList, typeof(S2CRoleListReq));
            RegisterS2CProtocol((int)S2CProtocol.S2CcreateRoleResult, typeof(S2CCreateRoleResultReq));
            RegisterS2CProtocol((int)S2CProtocol.S2CloginSuccess, typeof(S2CLoginSuccessReq));
            RegisterS2CProtocol((int)S2CProtocol.S2CreconnectSuccess, typeof(S2CReconnectSuccessReq));
            RegisterS2CProtocol((int)S2CProtocol.S2Cerror, typeof(ErrorData));
        }

        /// <summary>
        /// 注册C2S协议
        /// </summary>
        public static void RegisterC2SProtocol(int protocolId, Type messageType)
        {
            _c2sProtocolMap[protocolId] = messageType;
        }

        /// <summary>
        /// 注册S2C协议
        /// </summary>
        public static void RegisterS2CProtocol(int protocolId, Type messageType)
        {
            _s2cProtocolMap[protocolId] = messageType;
        }

        /// <summary>
        /// 序列化消息（C2S）
        /// 参考 example 客户端：直接编码 ClientMessage [2字节MsgId][protobuf数据]
        /// 然后包装在通用Message中: [4字节长度][1字节类型][1字节flags][ClientMessage数据]
        /// </summary>
        public static byte[] SerializeMessage(IMessage message, int protocolId, string sessionId = "")
        {
            try
            {
                // 1. 序列化protobuf消息体
                byte[] protoData = message.ToByteArray();

                // 2. 编码ClientMessage: [2字节MsgId][数据]（参考 example/game_client.go:126）
                // 使用BigEndian（网络字节序），与服务端保持一致
                byte[] clientMsgData = new byte[2 + protoData.Length];
                ushort protocolIdUshort = (ushort)protocolId;
                // BigEndian: 高字节在前
                clientMsgData[0] = (byte)((protocolIdUshort >> 8) & 0xFF); // 高字节
                clientMsgData[1] = (byte)(protocolIdUshort & 0xFF);        // 低字节
                Buffer.BlockCopy(protoData, 0, clientMsgData, 2, protoData.Length);

                // 3. 编码通用消息: [4字节长度][1字节类型(0x02=MsgTypeClient)][1字节flags][ClientMessage数据]
                // 注意：服务端使用DecodeMessageWithCompression，期望格式包含flags字节
                int totalLen = 2 + clientMsgData.Length; // 1字节类型 + 1字节flags + payload
                byte[] result = new byte[4 + totalLen];
                // 长度使用BigEndian（网络字节序）
                uint totalLenUint = (uint)totalLen;
                result[0] = (byte)((totalLenUint >> 24) & 0xFF); // 最高字节
                result[1] = (byte)((totalLenUint >> 16) & 0xFF);
                result[2] = (byte)((totalLenUint >> 8) & 0xFF);
                result[3] = (byte)(totalLenUint & 0xFF);         // 最低字节
                result[4] = 0x02; // MsgTypeClient
                result[5] = 0x00; // Flags: 0x00 = FlagNone (无压缩)
                Buffer.BlockCopy(clientMsgData, 0, result, 6, clientMsgData.Length);

                return result;
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[ProtocolHandler] 序列化消息失败: {ex.Message}");
                return null;
            }
        }

        /// <summary>
        /// 反序列化消息（S2C）
        /// 参考 example 客户端：直接解码 ClientMessage [2字节MsgId][protobuf数据]
        /// </summary>
        public static IMessage DeserializeMessage(byte[] clientMessagePayload, out int protocolId)
        {
            protocolId = 0;

            if (clientMessagePayload == null || clientMessagePayload.Length < 2)
            {
                GD.PrintErr("[ProtocolHandler] 消息数据不完整");
                return null;
            }

            try
            {
                // 直接解码ClientMessage: [2字节MsgId][数据]（参考 example/client_handler.go:202）
                protocolId = (clientMessagePayload[0] << 8) | clientMessagePayload[1]; // BigEndian
                byte[] messageData = new byte[clientMessagePayload.Length - 2];
                Buffer.BlockCopy(clientMessagePayload, 2, messageData, 0, messageData.Length);

                // 根据协议ID查找消息类型
                if (!_s2cProtocolMap.TryGetValue(protocolId, out Type messageType))
                {
                    GD.PrintErr($"[ProtocolHandler] 未知的协议ID: {protocolId}");
                    return null;
                }

                // 创建消息实例并反序列化
                IMessage message = (IMessage)Activator.CreateInstance(messageType);
                message.MergeFrom(messageData);

                return message;
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[ProtocolHandler] 反序列化消息失败: {ex.Message}");
                return null;
            }
        }

        /// <summary>
        /// 获取协议ID对应的消息类型（C2S）
        /// </summary>
        public static Type GetC2SMessageType(int protocolId)
        {
            _c2sProtocolMap.TryGetValue(protocolId, out Type type);
            return type;
        }

        /// <summary>
        /// 获取协议ID对应的消息类型（S2C）
        /// </summary>
        public static Type GetS2CMessageType(int protocolId)
        {
            _s2cProtocolMap.TryGetValue(protocolId, out Type type);
            return type;
        }

        /// <summary>
        /// 获取协议名称（用于日志显示）
        /// </summary>
        public static string GetProtocolName(int protocolId, bool isC2S)
        {
            if (isC2S)
            {
                // 尝试从C2SProtocol枚举获取名称
                if (System.Enum.IsDefined(typeof(C2SProtocol), protocolId))
                {
                    return ((C2SProtocol)protocolId).ToString();
                }
            }
            else
            {
                // 尝试从S2CProtocol枚举获取名称
                if (System.Enum.IsDefined(typeof(S2CProtocol), protocolId))
                {
                    return ((S2CProtocol)protocolId).ToString();
                }
            }
            return $"Unknown({protocolId})";
        }

        /// <summary>
        /// 获取消息内容摘要（用于日志显示）
        /// </summary>
        public static string GetMessageSummary(IMessage message)
        {
            if (message == null)
            {
                return "null";
            }

            try
            {
                // 根据消息类型生成摘要
                string typeName = message.GetType().Name;
                
                // 尝试提取关键字段
                var props = message.GetType().GetProperties();
                var summaryParts = new System.Collections.Generic.List<string>();
                
                foreach (var prop in props)
                {
                    // 只显示常见的关键字段
                    string propName = prop.Name.ToLower();
                    if (propName.Contains("username") || propName.Contains("password") || 
                        propName.Contains("rolename") || propName.Contains("roleid") ||
                        propName.Contains("success") || propName.Contains("message") ||
                        propName.Contains("token") || propName.Contains("rolelist"))
                    {
                        try
                        {
                            var value = prop.GetValue(message);
                            if (value != null)
                            {
                                // 对密码字段进行脱敏
                                if (propName.Contains("password"))
                                {
                                    summaryParts.Add($"{prop.Name}=***");
                                }
                                else
                                {
                                    string valueStr = value.ToString();
                                    // 限制字符串长度
                                    if (valueStr.Length > 50)
                                    {
                                        valueStr = valueStr.Substring(0, 50) + "...";
                                    }
                                    summaryParts.Add($"{prop.Name}={valueStr}");
                                }
                            }
                        }
                        catch
                        {
                            // 忽略属性访问错误
                        }
                    }
                }

                if (summaryParts.Count > 0)
                {
                    return $"{typeName} {{{string.Join(", ", summaryParts)}}}";
                }
                else
                {
                    return typeName;
                }
            }
            catch
            {
                return message.GetType().Name;
            }
        }
    }
}

