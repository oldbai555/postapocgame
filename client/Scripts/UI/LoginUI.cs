using Godot;
using System;
using PostApocGame.Network;
using PostApocGame.Data;
using Pb3;

namespace PostApocGame.UI
{
    /// <summary>
    /// 登录界面
    /// </summary>
    public partial class LoginUI : Control
    {
        // UI节点引用
        private LineEdit _usernameInput;
        private LineEdit _passwordInput;
        private CheckBox _rememberPasswordCheck;
        private Button _loginButton;
        private Button _registerButton;
        private Label _errorLabel;
        private Label _statusLabel;

        private bool _isConnecting = false;
        private string _currentToken = "";

        public override void _Ready()
        {
            // 获取UI节点（场景路径：CenterContainer/VBoxContainer/...）
            _usernameInput = GetNode<LineEdit>("CenterContainer/VBoxContainer/UsernameInput");
            _passwordInput = GetNode<LineEdit>("CenterContainer/VBoxContainer/PasswordInput");
            _rememberPasswordCheck = GetNode<CheckBox>("CenterContainer/VBoxContainer/RememberPasswordCheck");
            _loginButton = GetNode<Button>("CenterContainer/VBoxContainer/LoginButton");
            _registerButton = GetNode<Button>("CenterContainer/VBoxContainer/RegisterButton");
            _errorLabel = GetNode<Label>("CenterContainer/VBoxContainer/ErrorLabel");
            _statusLabel = GetNode<Label>("CenterContainer/VBoxContainer/StatusLabel");

            // 连接信号
            _loginButton.Pressed += OnLoginButtonPressed;
            _registerButton.Pressed += OnRegisterButtonPressed;

            // 注册协议处理器
            MessageReceiver.Instance?.RegisterHandler((int)S2CProtocol.S2CregisterResult, OnRegisterResult);
            MessageReceiver.Instance?.RegisterHandler((int)S2CProtocol.S2CloginResult, OnLoginResult);
            MessageReceiver.Instance?.RegisterHandler((int)S2CProtocol.S2Cerror, OnError);

            // 连接网络事件
            if (NetworkManager.Instance != null)
            {
                NetworkManager.Instance.OnConnected += OnNetworkConnected;
                NetworkManager.Instance.OnDisconnected += OnNetworkDisconnected;
            }

            // 加载保存的账号信息
            LoadSavedAccount();

            // 自动连接服务器
            AutoConnect();
        }

        /// <summary>
        /// 自动连接服务器
        /// </summary>
        private async void AutoConnect()
        {
            if (NetworkManager.Instance == null)
            {
                ShowError("网络管理器未初始化");
                return;
            }

            if (!NetworkManager.Instance.IsConnected)
            {
                SetStatus("正在连接服务器...");
                bool connected = await NetworkManager.Instance.ConnectAsync();
                if (!connected)
                {
                    ShowError("连接服务器失败，请检查网络");
                }
            }
        }

        /// <summary>
        /// 加载保存的账号信息
        /// </summary>
        private void LoadSavedAccount()
        {
            var (username, password, rememberPassword) = LocalStorage.LoadAccount();
            if (!string.IsNullOrEmpty(username))
            {
                _usernameInput.Text = username;
                _rememberPasswordCheck.ButtonPressed = rememberPassword;
                if (rememberPassword && !string.IsNullOrEmpty(password))
                {
                    _passwordInput.Text = password;
                }
            }
        }

        /// <summary>
        /// 登录按钮点击
        /// </summary>
        private void OnLoginButtonPressed()
        {
            if (_isConnecting)
            {
                return;
            }

            string username = _usernameInput.Text.Trim();
            string password = _passwordInput.Text;

            if (string.IsNullOrEmpty(username))
            {
                ShowError("请输入账号");
                return;
            }

            if (string.IsNullOrEmpty(password))
            {
                ShowError("请输入密码");
                return;
            }

            // 保存账号信息
            LocalStorage.SaveAccount(username, password, _rememberPasswordCheck.ButtonPressed);

            // 发送登录请求
            SendLoginRequest(username, password);
        }

        /// <summary>
        /// 注册按钮点击
        /// </summary>
        private void OnRegisterButtonPressed()
        {
            if (_isConnecting)
            {
                return;
            }

            string username = _usernameInput.Text.Trim();
            string password = _passwordInput.Text;

            if (string.IsNullOrEmpty(username))
            {
                ShowError("请输入账号");
                return;
            }

            if (string.IsNullOrEmpty(password))
            {
                ShowError("请输入密码");
                return;
            }

            if (password.Length < 6)
            {
                ShowError("密码长度至少6位");
                return;
            }

            // 发送注册请求
            SendRegisterRequest(username, password);
        }

        /// <summary>
        /// 发送登录请求
        /// </summary>
        private void SendLoginRequest(string username, string password)
        {
            if (NetworkManager.Instance == null || !NetworkManager.Instance.IsConnected)
            {
                ShowError("未连接到服务器");
                AutoConnect();
                return;
            }

            _isConnecting = true;
            SetStatus("正在登录...");
            ClearError();

            var req = new C2SLoginReq
            {
                Username = username,
                Password = password
            };

            MessageSender.Send(req, (int)C2SProtocol.C2Slogin);
        }

        /// <summary>
        /// 发送注册请求
        /// </summary>
        private void SendRegisterRequest(string username, string password)
        {
            if (NetworkManager.Instance == null || !NetworkManager.Instance.IsConnected)
            {
                ShowError("未连接到服务器");
                AutoConnect();
                return;
            }

            _isConnecting = true;
            SetStatus("正在注册...");
            ClearError();

            var req = new C2SRegisterReq
            {
                Username = username,
                Password = password
            };

            int protocolId = (int)C2SProtocol.C2Sregister;
            GD.Print($"[LoginUI] 发送注册请求，协议ID: {protocolId} (C2Sregister = {(int)C2SProtocol.C2Sregister})");
            MessageSender.Send(req, protocolId);
        }

        /// <summary>
        /// 处理注册结果
        /// </summary>
        private void OnRegisterResult(Google.Protobuf.IMessage message)
        {
            if (message is S2CRegisterResultReq result)
            {
                _isConnecting = false;
                SetStatus("");

                if (result.Success)
                {
                    _currentToken = result.Token;
                    ShowError("注册成功！请登录", false);
                    // 注册成功后自动登录
                    SendLoginRequest(_usernameInput.Text.Trim(), _passwordInput.Text);
                }
                else
                {
                    ShowError(result.Message);
                }
            }
        }

        /// <summary>
        /// 处理登录结果
        /// </summary>
        private void OnLoginResult(Google.Protobuf.IMessage message)
        {
            if (message is S2CLoginResultReq result)
            {
                _isConnecting = false;
                SetStatus("");

                if (result.Success)
                {
                    _currentToken = result.Token;
                    // 保存Token（用于重连）
                    LocalStorage.SaveString("last_token", _currentToken);
                    
                    // 登录成功，切换到角色选择界面
                    GD.Print($"[LoginUI] 登录成功，Token: {_currentToken}");
                    SwitchToRoleSelect();
                }
                else
                {
                    ShowError(result.Message);
                }
            }
        }

        /// <summary>
        /// 处理错误消息
        /// </summary>
        private void OnError(Google.Protobuf.IMessage message)
        {
            if (message is ErrorData error)
            {
                _isConnecting = false;
                SetStatus("");
                ShowError($"错误 {error.Code}: {error.Msg}");
            }
        }

        /// <summary>
        /// 网络连接成功
        /// </summary>
        private void OnNetworkConnected()
        {
            SetStatus("已连接到服务器");
            ClearError();
        }

        /// <summary>
        /// 网络断开
        /// </summary>
        private void OnNetworkDisconnected()
        {
            _isConnecting = false;
            SetStatus("连接已断开");
            ShowError("与服务器断开连接");
        }

        /// <summary>
        /// 切换到角色选择界面
        /// </summary>
        private void SwitchToRoleSelect()
        {
            // 切换到角色选择场景
            // 注意：需要先创建RoleSelect.tscn场景文件
            GetTree().ChangeSceneToFile("res://Scenes/RoleSelect.tscn");
        }

        /// <summary>
        /// 显示错误信息
        /// </summary>
        private void ShowError(string message, bool isError = true)
        {
            _errorLabel.Text = message;
            _errorLabel.Modulate = isError ? Colors.Red : Colors.Green;
            _errorLabel.Visible = true;
        }

        /// <summary>
        /// 清除错误信息
        /// </summary>
        private void ClearError()
        {
            _errorLabel.Text = "";
            _errorLabel.Visible = false;
        }

        /// <summary>
        /// 设置状态信息
        /// </summary>
        private void SetStatus(string status)
        {
            _statusLabel.Text = status;
            _statusLabel.Visible = !string.IsNullOrEmpty(status);
        }

        public override void _ExitTree()
        {
            // 取消注册协议处理器
            MessageReceiver.Instance?.UnregisterHandler((int)S2CProtocol.S2CregisterResult);
            MessageReceiver.Instance?.UnregisterHandler((int)S2CProtocol.S2CloginResult);
            MessageReceiver.Instance?.UnregisterHandler((int)S2CProtocol.S2Cerror);

            base._ExitTree();
        }
    }
}

