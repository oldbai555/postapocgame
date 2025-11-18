using Godot;
using System;
using System.Collections.Generic;
using PostApocGame.Network;
using Pb3;

namespace PostApocGame.UI
{
    /// <summary>
    /// 角色选择界面
    /// </summary>
    public partial class RoleSelectUI : Control
    {
        // UI节点引用
        private VBoxContainer _roleListContainer;
        private Button _createRoleButton;
        private Button _enterGameButton;
        private Button _deleteRoleButton;
        private Label _roleInfoLabel;
        private Label _errorLabel;
        private Label _statusLabel;

        private List<PlayerSimpleData> _roleList = new List<PlayerSimpleData>();
        private PlayerSimpleData _selectedRole = null;
        private bool _isLoading = false;

        public override void _Ready()
        {
            // 获取UI节点（场景路径：CenterContainer/VBoxContainer/...）
            _roleListContainer = GetNode<VBoxContainer>("CenterContainer/VBoxContainer/RoleListScrollContainer/RoleListContainer");
            _createRoleButton = GetNode<Button>("CenterContainer/VBoxContainer/CreateRoleButton");
            _enterGameButton = GetNode<Button>("CenterContainer/VBoxContainer/EnterGameButton");
            _deleteRoleButton = GetNode<Button>("CenterContainer/VBoxContainer/DeleteRoleButton");
            _roleInfoLabel = GetNode<Label>("CenterContainer/VBoxContainer/RoleInfoLabel");
            _errorLabel = GetNode<Label>("CenterContainer/VBoxContainer/ErrorLabel");
            _statusLabel = GetNode<Label>("CenterContainer/VBoxContainer/StatusLabel");

            // 连接信号
            _createRoleButton.Pressed += OnCreateRoleButtonPressed;
            _enterGameButton.Pressed += OnEnterGameButtonPressed;
            _deleteRoleButton.Pressed += OnDeleteRoleButtonPressed;

            // 注册协议处理器
            MessageReceiver.Instance?.RegisterHandler((int)S2CProtocol.S2CroleList, OnRoleList);
            MessageReceiver.Instance?.RegisterHandler((int)S2CProtocol.S2CcreateRoleResult, OnCreateRoleResult);
            MessageReceiver.Instance?.RegisterHandler((int)S2CProtocol.S2CloginSuccess, OnLoginSuccess);
            MessageReceiver.Instance?.RegisterHandler((int)S2CProtocol.S2Cerror, OnError);

            // 初始化UI
            UpdateUI();

            // 请求角色列表
            RequestRoleList();
        }

        /// <summary>
        /// 请求角色列表
        /// </summary>
        private void RequestRoleList()
        {
            if (NetworkManager.Instance == null || !NetworkManager.Instance.IsConnected)
            {
                ShowError("未连接到服务器");
                return;
            }

            _isLoading = true;
            SetStatus("正在加载角色列表...");
            ClearError();

            var req = new C2SQueryRolesReq();
            MessageSender.Send(req, (int)C2SProtocol.C2SqueryRoles);
        }

        /// <summary>
        /// 处理角色列表
        /// </summary>
        private void OnRoleList(Google.Protobuf.IMessage message)
        {
            if (message is S2CRoleListReq result)
            {
                _isLoading = false;
                SetStatus("");

                _roleList.Clear();
                foreach (var role in result.RoleList)
                {
                    _roleList.Add(role);
                }

                UpdateRoleListUI();
                UpdateUI();
            }
        }

        /// <summary>
        /// 更新角色列表UI
        /// </summary>
        private void UpdateRoleListUI()
        {
            // 清空现有按钮
            foreach (Node child in _roleListContainer.GetChildren())
            {
                child.QueueFree();
            }

            // 创建角色按钮
            foreach (var role in _roleList)
            {
                Button roleButton = new Button();
                roleButton.Text = $"{role.RoleName} (Lv.{role.Level})";
                roleButton.Pressed += () => OnRoleSelected(role);
                // 设置按钮不扩展高度，只使用内容高度
                roleButton.SizeFlagsVertical = Control.SizeFlags.ShrinkCenter;
                _roleListContainer.AddChild(roleButton);
            }
        }

        /// <summary>
        /// 角色被选中
        /// </summary>
        private void OnRoleSelected(PlayerSimpleData role)
        {
            _selectedRole = role;
            UpdateRoleInfo();
            UpdateUI();
        }

        /// <summary>
        /// 更新角色信息显示
        /// </summary>
        private void UpdateRoleInfo()
        {
            if (_selectedRole == null)
            {
                _roleInfoLabel.Text = "请选择角色";
                return;
            }

            string jobName = GetJobName(_selectedRole.Job);
            string sexName = _selectedRole.Sex == 1 ? "男" : "女";
            _roleInfoLabel.Text = $"角色名: {_selectedRole.RoleName}\n" +
                                  $"职业: {jobName}\n" +
                                  $"性别: {sexName}\n" +
                                  $"等级: {_selectedRole.Level}";
        }

        /// <summary>
        /// 获取职业名称
        /// </summary>
        private string GetJobName(uint job)
        {
            switch (job)
            {
                case 1: return "战士";
                case 2: return "法师";
                case 3: return "刺客";
                default: return "未知";
            }
        }

        /// <summary>
        /// 创建角色按钮点击
        /// </summary>
        private void OnCreateRoleButtonPressed()
        {
            if (_isLoading)
            {
                return;
            }

            if (_roleList.Count >= 3)
            {
                ShowError("最多只能创建3个角色");
                return;
            }

            // 打开创建角色对话框
            ShowCreateRoleDialog();
        }

        /// <summary>
        /// 显示创建角色对话框
        /// </summary>
        private void ShowCreateRoleDialog()
        {
            // 创建对话框窗口
            AcceptDialog dialog = new AcceptDialog();
            dialog.Title = "创建角色";
            dialog.DialogText = "请输入角色信息";
            
            // 创建对话框内容容器
            VBoxContainer dialogContent = new VBoxContainer();
            dialogContent.CustomMinimumSize = new Vector2(300, 200);
            
            // 角色名输入
            Label nameLabel = new Label();
            nameLabel.Text = "角色名（2-12个字符）：";
            LineEdit nameInput = new LineEdit();
            nameInput.PlaceholderText = "请输入角色名";
            nameInput.MaxLength = 12;
            
            // 职业选择
            Label jobLabel = new Label();
            jobLabel.Text = "职业：";
            OptionButton jobSelect = new OptionButton();
            jobSelect.AddItem("战士");
            jobSelect.AddItem("法师");
            jobSelect.AddItem("刺客");
            jobSelect.Selected = 0;
            
            // 性别选择
            Label sexLabel = new Label();
            sexLabel.Text = "性别：";
            OptionButton sexSelect = new OptionButton();
            sexSelect.AddItem("男");
            sexSelect.AddItem("女");
            sexSelect.Selected = 0;
            
            // 添加到容器
            dialogContent.AddChild(nameLabel);
            dialogContent.AddChild(nameInput);
            dialogContent.AddChild(jobLabel);
            dialogContent.AddChild(jobSelect);
            dialogContent.AddChild(sexLabel);
            dialogContent.AddChild(sexSelect);
            
            // 将内容添加到对话框
            dialog.AddChild(dialogContent);
            
            // 确认按钮处理
            dialog.Confirmed += () =>
            {
                string roleName = nameInput.Text.Trim();
                uint job = (uint)(jobSelect.Selected + 1); // 1=战士, 2=法师, 3=刺客
                uint sex = (uint)(sexSelect.Selected + 1); // 1=男, 2=女
                
                if (string.IsNullOrEmpty(roleName) || roleName.Length < 2 || roleName.Length > 12)
                {
                    ShowError("角色名长度必须在2-12个字符之间");
                    dialog.QueueFree();
                    return;
                }
                
                CreateRole(roleName, job, sex);
                dialog.QueueFree();
            };
            
            // 取消按钮处理
            dialog.Canceled += () =>
            {
                dialog.QueueFree();
            };
            
            // 显示对话框
            AddChild(dialog);
            dialog.PopupCentered();
            
            // 聚焦到输入框
            nameInput.GrabFocus();
        }

        /// <summary>
        /// 创建角色
        /// </summary>
        private void CreateRole(string roleName, uint job, uint sex)
        {
            if (NetworkManager.Instance == null || !NetworkManager.Instance.IsConnected)
            {
                ShowError("未连接到服务器");
                return;
            }

            if (string.IsNullOrEmpty(roleName) || roleName.Length < 2 || roleName.Length > 12)
            {
                ShowError("角色名长度必须在2-12个字符之间");
                return;
            }

            _isLoading = true;
            SetStatus("正在创建角色...");
            ClearError();

            var roleData = new PlayerSimpleData
            {
                Job = job,
                Sex = sex,
                RoleName = roleName,
                Level = 1
            };

            var req = new C2SCreateRoleReq
            {
                RoleData = roleData
            };

            MessageSender.Send(req, (int)C2SProtocol.C2ScreateRole);
        }

        /// <summary>
        /// 处理创建角色结果
        /// </summary>
        private void OnCreateRoleResult(Google.Protobuf.IMessage message)
        {
            if (message is S2CCreateRoleResultReq result)
            {
                _isLoading = false;
                SetStatus("");

                // S2CCreateRoleResultReq包含创建成功的角色信息
                ShowError($"角色创建成功！{result.RoleName}", false);
                // 重新请求角色列表
                RequestRoleList();
            }
        }

        /// <summary>
        /// 进入游戏按钮点击
        /// </summary>
        private void OnEnterGameButtonPressed()
        {
            if (_isLoading)
            {
                return;
            }

            if (_selectedRole == null)
            {
                ShowError("请先选择角色");
                return;
            }

            EnterGame();
        }

        /// <summary>
        /// 进入游戏
        /// </summary>
        private void EnterGame()
        {
            if (NetworkManager.Instance == null || !NetworkManager.Instance.IsConnected)
            {
                ShowError("未连接到服务器");
                return;
            }

            _isLoading = true;
            SetStatus("正在进入游戏...");
            ClearError();

            var req = new C2SEnterGameReq
            {
                RoleId = _selectedRole.RoleId
            };

            MessageSender.Send(req, (int)C2SProtocol.C2SenterGame);
        }

        /// <summary>
        /// 处理登录成功（进入游戏成功）
        /// </summary>
        private void OnLoginSuccess(Google.Protobuf.IMessage message)
        {
            if (message is S2CLoginSuccessReq result)
            {
                _isLoading = false;
                SetStatus("");

                GD.Print($"[RoleSelectUI] 进入游戏成功，角色ID: {result.RoleData.RoleId}");
                
                // 切换到游戏主场景
                // GetTree().ChangeSceneToFile("res://Scenes/Main.tscn");
                GD.Print("[RoleSelectUI] 进入游戏成功，等待实现游戏主场景");
            }
        }

        /// <summary>
        /// 删除角色按钮点击
        /// </summary>
        private void OnDeleteRoleButtonPressed()
        {
            if (_isLoading)
            {
                return;
            }

            if (_selectedRole == null)
            {
                ShowError("请先选择要删除的角色");
                return;
            }

            // 显示确认对话框
            // 实际项目中应该创建一个确认对话框
            GD.Print($"[RoleSelectUI] 删除角色确认（需要实现确认对话框）: {_selectedRole.RoleName}");
            ShowError("删除角色功能待实现");
        }

        /// <summary>
        /// 处理错误消息
        /// </summary>
        private void OnError(Google.Protobuf.IMessage message)
        {
            if (message is ErrorData error)
            {
                _isLoading = false;
                SetStatus("");
                ShowError($"错误 {error.Code}: {error.Msg}");
            }
        }

        /// <summary>
        /// 更新UI状态
        /// </summary>
        private void UpdateUI()
        {
            _createRoleButton.Disabled = _isLoading || _roleList.Count >= 3;
            _enterGameButton.Disabled = _isLoading || _selectedRole == null;
            _deleteRoleButton.Disabled = _isLoading || _selectedRole == null;
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
            MessageReceiver.Instance?.UnregisterHandler((int)S2CProtocol.S2CroleList);
            MessageReceiver.Instance?.UnregisterHandler((int)S2CProtocol.S2CcreateRoleResult);
            MessageReceiver.Instance?.UnregisterHandler((int)S2CProtocol.S2CloginSuccess);
            MessageReceiver.Instance?.UnregisterHandler((int)S2CProtocol.S2Cerror);

            base._ExitTree();
        }
    }
}

