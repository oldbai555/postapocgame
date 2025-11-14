using Godot;
using System;
using System.Net.Sockets;
using System.Threading.Tasks;
using System.Collections.Generic;

namespace PostApocGame.Network
{
    /// <summary>
    /// 网络管理器 - 负责TCP连接管理、心跳、断线重连
    /// </summary>
    public partial class NetworkManager : Node
    {
        private static NetworkManager _instance;
        public static NetworkManager Instance => _instance;

        private TcpClient _tcpClient;
        private NetworkStream _stream;
        private bool _isConnected = false;
        private bool _isConnecting = false;

        /// <summary>
        /// 是否已连接
        /// </summary>
        public new bool IsConnected => _isConnected;

        // 配置
        [Export] public string ServerHost { get; set; } = "127.0.0.1"; // 客户端连接地址，不能使用0.0.0.0
        [Export] public int ServerPort { get; set; } = 1011;
        [Export] public float HeartbeatInterval { get; set; } = 5.0f; // 心跳间隔（秒）
        [Export] public float ReconnectInterval { get; set; } = 3.0f; // 重连间隔（秒）
        [Export] public int MaxReconnectAttempts { get; set; } = 5; // 最大重连次数

        // 心跳相关
        private float _heartbeatTimer = 0f;
        private float _lastHeartbeatTime = 0f;
        private float _reconnectTimer = 0f;
        private int _reconnectAttempts = 0;

        // 时间同步
        private long _serverTimeOffset = 0; // 服务器时间偏移量（毫秒）
        private bool _timeSynced = false;

        // 事件
        public event Action OnConnected;
        public event Action OnDisconnected;
        public event Action<string> OnError;

        public override void _Ready()
        {
            _instance = this;
            MessageReceiver.Instance?.SetNetworkManager(this);
        }

        public override void _Process(double delta)
        {
            if (_isConnected)
            {
                // 心跳检测
                _heartbeatTimer += (float)delta;
                if (_heartbeatTimer >= HeartbeatInterval)
                {
                    _heartbeatTimer = 0f;
                    SendHeartbeat();
                }

                // 检查连接状态
                if (_tcpClient != null && !_tcpClient.Connected)
                {
                    HandleDisconnection();
                }
            }
            else if (!_isConnecting && _reconnectAttempts < MaxReconnectAttempts)
            {
                // 自动重连
                _reconnectTimer += (float)delta;
                if (_reconnectTimer >= ReconnectInterval)
                {
                    _reconnectTimer = 0f;
                    _reconnectAttempts++;
                    // 不等待连接完成，继续执行
                    _ = ConnectAsync();
                }
            }
        }

        /// <summary>
        /// 连接到服务器
        /// </summary>
        public async Task<bool> ConnectAsync()
        {
            if (_isConnecting || _isConnected)
            {
                return _isConnected;
            } 

            _isConnecting = true;

            try
            {
                // 验证服务器地址
                if (string.IsNullOrEmpty(ServerHost) || ServerHost == "0.0.0.0")
                {
                    string errorMsg = "服务器地址无效：0.0.0.0 是服务器监听地址，客户端应使用 127.0.0.1（本地）或服务器实际IP地址";
                    GD.PrintErr($"[NetworkManager] {errorMsg}");
                    _isConnecting = false;
                    OnError?.Invoke(errorMsg);
                    return false;
                }

                _tcpClient = new TcpClient();
                await _tcpClient.ConnectAsync(ServerHost, ServerPort);
                _stream = _tcpClient.GetStream();
                _isConnected = true;
                _isConnecting = false;
                _reconnectAttempts = 0;

                GD.Print($"[NetworkManager] 已连接到服务器 {ServerHost}:{ServerPort}");

                // 启动接收线程
                MessageReceiver.Instance?.StartReceiving(_stream);

                OnConnected?.Invoke();
                return true;
            }
            catch (Exception ex)
            {
                _isConnecting = false;
                _isConnected = false;
                string errorMsg = $"连接失败: {ex.Message}";
                GD.PrintErr($"[NetworkManager] {errorMsg}");
                OnError?.Invoke(errorMsg);
                return false;
            }
        }

        /// <summary>
        /// 断开连接
        /// </summary>
        public void Disconnect()
        {
            if (!_isConnected)
            {
                return;
            }

            _isConnected = false;
            _isConnecting = false;

            try
            {
                MessageReceiver.Instance?.StopReceiving();
                _stream?.Close();
                _tcpClient?.Close();
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[NetworkManager] 断开连接时出错: {ex.Message}");
            }

            _stream = null;
            _tcpClient = null;

            GD.Print("[NetworkManager] 已断开连接");
            OnDisconnected?.Invoke();
        }

        /// <summary>
        /// 发送消息（data已经是完整的消息格式：长度+类型+内容）
        /// </summary>
        public bool SendMessage(byte[] data)
        {
            if (!_isConnected || _stream == null)
            {
                GD.PrintErr("[NetworkManager] 未连接，无法发送消息");
                return false;
            }

            try
            {
                _stream.Write(data, 0, data.Length);
                _stream.Flush();
                return true;
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[NetworkManager] 发送消息失败: {ex.Message}");
                HandleDisconnection();
                return false;
            }
        }

        /// <summary>
        /// 发送心跳
        /// </summary>
        private void SendHeartbeat()
        {
            try
            {
                // 服务端消息格式: [4字节长度][1字节类型][1字节flags][消息体]
                // 心跳消息类型: 0x06 (MsgTypeHeartbeat)
                // 注意：服务端使用DecodeMessageWithCompression，期望格式包含flags字节
                byte[] pingBytes = System.Text.Encoding.UTF8.GetBytes("ping");
                int totalLen = 2 + pingBytes.Length; // 1字节类型 + 1字节flags + payload
                // 使用BigEndian编码长度（网络字节序，与服务端保持一致）
                uint totalLenUint = (uint)totalLen;
                byte[] lengthBytes = new byte[4];
                lengthBytes[0] = (byte)((totalLenUint >> 24) & 0xFF); // 最高字节
                lengthBytes[1] = (byte)((totalLenUint >> 16) & 0xFF);
                lengthBytes[2] = (byte)((totalLenUint >> 8) & 0xFF);
                lengthBytes[3] = (byte)(totalLenUint & 0xFF);         // 最低字节

                // 构建完整消息：长度(4) + 类型(1) + flags(1) + 内容
                byte[] message = new byte[4 + 2 + pingBytes.Length];
                Buffer.BlockCopy(lengthBytes, 0, message, 0, 4);
                message[4] = 0x06; // MsgTypeHeartbeat
                message[5] = 0x00; // Flags: 0x00 = FlagNone (无压缩)
                Buffer.BlockCopy(pingBytes, 0, message, 6, pingBytes.Length);

                _stream.Write(message, 0, message.Length);
                _stream.Flush();
                _lastHeartbeatTime = (float)Time.GetUnixTimeFromSystem();
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[NetworkManager] 发送心跳失败: {ex.Message}");
                HandleDisconnection();
            }
        }

        /// <summary>
        /// 处理断线
        /// </summary>
        private void HandleDisconnection()
        {
            if (!_isConnected)
            {
                return;
            }

            GD.Print("[NetworkManager] 检测到断线");
            Disconnect();
        }

        /// <summary>
        /// 获取服务器时间（毫秒）
        /// </summary>
        public long GetServerTime()
        {
            if (!_timeSynced)
            {
                return DateTimeOffset.UtcNow.ToUnixTimeMilliseconds();
            }
            return DateTimeOffset.UtcNow.ToUnixTimeMilliseconds() + _serverTimeOffset;
        }

        /// <summary>
        /// 设置服务器时间偏移
        /// </summary>
        public void SetServerTimeOffset(long offset)
        {
            _serverTimeOffset = offset;
            _timeSynced = true;
        }

        public override void _ExitTree()
        {
            Disconnect();
            base._ExitTree();
        }
    }
}

