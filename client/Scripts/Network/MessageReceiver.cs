using Godot;
using System;
using System.IO;
using System.Net.Sockets;
using System.Threading;
using System.Threading.Tasks;
using Google.Protobuf;
using System.Collections.Generic;

namespace PostApocGame.Network
{
    /// <summary>
    /// 消息接收器 - 负责接收和分发服务器消息
    /// </summary>
    public partial class MessageReceiver : Node
    {
        private static MessageReceiver _instance;
        public static MessageReceiver Instance => _instance;

        private NetworkStream _stream;
        private CancellationTokenSource _cancellationTokenSource;
        private Task _receiveTask;
        private bool _isReceiving = false;

        // 消息队列（线程安全）
        private Queue<ReceivedMessage> _messageQueue = new Queue<ReceivedMessage>();
        private readonly object _queueLock = new object();

        // 协议处理器字典
        private Dictionary<int, Action<IMessage>> _protocolHandlers = new Dictionary<int, Action<IMessage>>();

        public override void _Ready()
        {
            _instance = this;
        }

        /// <summary>
        /// 设置NetworkManager引用
        /// </summary>
        public void SetNetworkManager(NetworkManager networkManager)
        {
            // 可以在这里设置一些回调
        }

        /// <summary>
        /// 开始接收消息
        /// </summary>
        public void StartReceiving(NetworkStream stream)
        {
            if (_isReceiving)
            {
                return;
            }

            _stream = stream;
            _cancellationTokenSource = new CancellationTokenSource();
            _isReceiving = true;

            _receiveTask = Task.Run(() => ReceiveLoop(_cancellationTokenSource.Token));
            GD.Print("[MessageReceiver] 开始接收消息");
        }

        /// <summary>
        /// 停止接收消息
        /// </summary>
        public void StopReceiving()
        {
            if (!_isReceiving)
            {
                return;
            }

            _isReceiving = false;
            _cancellationTokenSource?.Cancel();

            try
            {
                _receiveTask?.Wait(1000); // 等待最多1秒
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[MessageReceiver] 停止接收时出错: {ex.Message}");
            }

            _cancellationTokenSource?.Dispose();
            _cancellationTokenSource = null;
            _receiveTask = null;
            _stream = null;

            GD.Print("[MessageReceiver] 已停止接收消息");
        }

        /// <summary>
        /// 接收循环（在后台线程运行）
        /// 服务端消息格式: [4字节长度][1字节类型][消息体]
        /// </summary>
        private void ReceiveLoop(CancellationToken cancellationToken)
        {
            byte[] lengthBuffer = new byte[4];

            while (!cancellationToken.IsCancellationRequested && _stream != null)
            {
                try
                {
                    // 读取消息长度
                    int bytesRead = 0;
                    while (bytesRead < 4)
                    {
                        int read = _stream.Read(lengthBuffer, bytesRead, 4 - bytesRead);
                        if (read == 0)
                        {
                            // 连接已关闭
                            CallDeferred(nameof(HandleDisconnection));
                            return;
                        }
                        bytesRead += read;
                    }

                    // 使用BigEndian解码长度（网络字节序，与服务端保持一致）
                    int messageLength = (lengthBuffer[0] << 24) | (lengthBuffer[1] << 16) | (lengthBuffer[2] << 8) | lengthBuffer[3];
                    if (messageLength <= 0 || messageLength > 1024 * 1024) // 最大1MB
                    {
                        GD.PrintErr($"[MessageReceiver] 无效的消息长度: {messageLength}");
                        break;
                    }

                    // 读取消息内容（类型+payload）
                    byte[] messageBuffer = new byte[messageLength];
                    bytesRead = 0;
                    while (bytesRead < messageLength)
                    {
                        int read = _stream.Read(messageBuffer, bytesRead, messageLength - bytesRead);
                        if (read == 0)
                        {
                            CallDeferred(nameof(HandleDisconnection));
                            return;
                        }
                        bytesRead += read;
                    }

                    // 检查消息类型和flags
                    // 服务端消息格式: [4字节长度][1字节类型][1字节flags][payload]
                    byte messageType = messageBuffer[0];
                    byte flags = messageBuffer[1];
                    byte[] payload = new byte[messageLength - 2]; // 减去类型和flags
                    Buffer.BlockCopy(messageBuffer, 2, payload, 0, payload.Length);

                    // 处理心跳消息
                    if (messageType == 0x06) // MsgTypeHeartbeat
                    {
                        // 心跳消息，忽略
                        continue;
                    }

                    // 处理客户端消息 (MsgTypeClient = 0x02)
                    if (messageType == 0x02)
                    {
                        // 直接解码ClientMessage（参考 example/client_handler.go:202）
                        // payload 就是 ClientMessage 数据: [2字节MsgId][protobuf数据]
                        IMessage message = ProtocolHandler.DeserializeMessage(payload, out int protocolId);
                        if (message != null)
                        {
                            // 获取协议名称和消息摘要
                            string protocolName = ProtocolHandler.GetProtocolName(protocolId, false);
                            string messageSummary = ProtocolHandler.GetMessageSummary(message);
                            GD.Print($"[MessageReceiver] 📥 收到消息: {protocolName} (ID={protocolId}) | {messageSummary}");

                            // 将消息加入队列，在主线程处理
                            lock (_queueLock)
                            {
                                _messageQueue.Enqueue(new ReceivedMessage { ProtocolId = protocolId, Message = message });
                            }
                        }
                    }
                }
                catch (IOException ex)
                {
                    GD.PrintErr($"[MessageReceiver] 接收消息时IO错误: {ex.Message}");
                    CallDeferred(nameof(HandleDisconnection));
                    break;
                }
                catch (Exception ex)
                {
                    GD.PrintErr($"[MessageReceiver] 接收消息时出错: {ex.Message}");
                    // 继续接收，不中断
                }
            }
        }

        /// <summary>
        /// 在主线程处理消息队列
        /// </summary>
        public override void _Process(double delta)
        {
            // 处理消息队列
            while (true)
            {
                ReceivedMessage receivedMsg = null;
                lock (_queueLock)
                {
                    if (_messageQueue.Count > 0)
                    {
                        receivedMsg = _messageQueue.Dequeue();
                    }
                }

                if (receivedMsg == null)
                {
                    break;
                }

                HandleMessage(receivedMsg.ProtocolId, receivedMsg.Message);
            }
        }

        /// <summary>
        /// 处理接收到的消息
        /// </summary>
        private void HandleMessage(int protocolId, IMessage message)
        {
            if (_protocolHandlers.TryGetValue(protocolId, out Action<IMessage> handler))
            {
                try
                {
                    handler(message);
                }
                catch (Exception ex)
                {
                    GD.PrintErr($"[MessageReceiver] 处理协议 {protocolId} 时出错: {ex.Message}");
                }
            }
            else
            {
                GD.Print($"[MessageReceiver] 未注册的协议处理器: {protocolId}");
            }
        }

        /// <summary>
        /// 注册协议处理器
        /// </summary>
        public void RegisterHandler(int protocolId, Action<IMessage> handler)
        {
            _protocolHandlers[protocolId] = handler;
        }

        /// <summary>
        /// 取消注册协议处理器
        /// </summary>
        public void UnregisterHandler(int protocolId)
        {
            _protocolHandlers.Remove(protocolId);
        }

        /// <summary>
        /// 处理断线
        /// </summary>
        private void HandleDisconnection()
        {
            NetworkManager.Instance?.Disconnect();
        }

        public override void _ExitTree()
        {
            StopReceiving();
            base._ExitTree();
        }

        /// <summary>
        /// 接收到的消息结构
        /// </summary>
        private class ReceivedMessage
        {
            public int ProtocolId { get; set; }
            public IMessage Message { get; set; }
        }
    }
}

