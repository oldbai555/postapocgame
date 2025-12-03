# Godot快速入门指南

> 本文档面向Godot客户端开发初学者，帮助快速上手Godot引擎，特别是搭建页面、场景等基础操作。

---

## 🚀 快速开始

### 1. 安装Godot

1. **下载Godot**
   - 官网: https://godotengine.org/download
   - 推荐版本: **Godot 4.5** (已安装，支持C#)
   - 下载时选择: **.NET版本** (Mono/C#支持)

2. **安装Godot**
   - Windows: 下载exe，解压即可运行（无需安装）
   - 建议: 将Godot.exe放在固定位置，创建桌面快捷方式
   - **注意**: 你已安装Godot 4.5，可以直接使用

3. **安装.NET SDK** (C#开发必需)
   - 下载: https://dotnet.microsoft.com/download
   - 安装: **.NET SDK 10.0.100** (已安装)
   - 验证: 打开命令行（PowerShell或CMD），输入以下命令检查版本：
     ```bash
     dotnet --version
     ```
   - 应该显示: `10.0.100` 或类似版本号
   - **注意**: .NET 10.0 完全兼容Godot 4.5，是推荐的版本

---

## 📚 推荐学习资源

### 基础教程（必看）

1. **Godot官方文档** (中文)
   - 地址: https://docs.godotengine.org/zh_CN/stable/
   - 推荐: 2D游戏开发教程
   - 适合: 完全新手

2. **Godot官方示例**
   - GitHub: https://github.com/godotengine/godot-demo-projects
   - 推荐: `2d` 目录下的示例项目
   - 适合: 学习实际代码

3. **B站视频教程** (中文)
   - 搜索: "Godot 4 教程"、"Godot C# 教程"
   - 推荐UP主:
     - 极致游戏学院
     - 展露编程
     - 暗黑烈焰使
   - 适合: 视觉学习

4. **YouTube教程** (英文)
   - 搜索: "Godot 4 C# tutorial"
   - 推荐频道:
     - GDQuest
     - HeartBeast
     - KidsCanCode
   - 适合: 深入理解

### 针对你的项目

由于你的项目是**2D横版格斗游戏**，建议重点学习：

1. **2D场景管理**
   - 场景树（Scene Tree）
   - 节点系统（Node System）
   - 场景切换

2. **2D渲染**
   - Sprite2D（精灵）
   - AnimatedSprite2D（动画精灵）
   - TileMap（瓦片地图）
   - Camera2D（相机）

3. **UI系统**
   - Control节点
   - UI布局（MarginContainer, VBoxContainer, HBoxContainer等）
   - 按钮、标签、输入框等控件

4. **输入系统**
   - Input类
   - 键盘输入
   - 触屏输入

5. **C#脚本**
   - GDScript vs C#
   - C#脚本创建和编辑
   - 信号系统（Signals）

---

## 🎨 免费UI资源推荐

### 1. UI资源网站

#### 优先推荐（完全免费，可商用）

1. **Kenney.nl** ⭐⭐⭐⭐⭐
   - 地址: https://kenney.nl/assets
   - 特点: 完全免费，可商用，无需署名
   - 类型: UI资源包、图标、按钮等
   - 推荐: "UI Pack" 系列
   - 格式: PNG（带透明通道）

2. **OpenGameArt.org** ⭐⭐⭐⭐
   - 地址: https://opengameart.org/
   - 特点: 免费，注意许可证（CC0、CC BY等）
   - 类型: 各种游戏资源（UI、角色、背景等）
   - 推荐: 搜索 "UI"、"HUD"、"button" 等关键词
   - 格式: 多种格式

3. **Itch.io** ⭐⭐⭐⭐
   - 地址: https://itch.io/game-assets/free
   - 特点: 免费资源，注意许可证
   - 类型: 完整的UI资源包
   - 推荐: 搜索 "free ui"、"free hud"
   - 格式: 多种格式

4. **Game Dev Market** ⭐⭐⭐
   - 地址: https://www.gamedevmarket.net/category/free-assets/
   - 特点: 有免费资源，也有付费资源
   - 类型: 高质量UI资源包
   - 推荐: 筛选 "Free" 标签
   - 格式: 多种格式

5. **Pixel Game Art** ⭐⭐⭐
   - 地址: https://www.pinterest.com/pixelgameart/
   - 特点: 像素风格UI资源
   - 类型: 像素风格UI、图标
   - 适合: 像素风格游戏

#### 图标资源

1. **Icons8** ⭐⭐⭐⭐⭐
   - 地址: https://icons8.com/
   - 特点: 大量免费图标，可商用（需署名或购买）
   - 类型: UI图标、游戏图标
   - 推荐: 游戏类图标

2. **Flaticon** ⭐⭐⭐⭐
   - 地址: https://www.flaticon.com/
   - 特点: 免费图标库，注意许可证
   - 类型: 各种图标
   - 推荐: 搜索 "game"、"ui" 等

3. **Font Awesome** ⭐⭐⭐⭐
   - 地址: https://fontawesome.com/
   - 特点: 图标字体，免费版可用
   - 类型: 图标字体
   - 适合: 现代风格UI

### 2. 字体资源

1. **Google Fonts**
   - 地址: https://fonts.google.com/
   - 特点: 完全免费，可商用
   - 推荐: 游戏字体
     - Orbitron（科幻风格）
     - Bangers（卡通风格）
     - Press Start 2P（像素风格）

2. **DaFont**
   - 地址: https://www.dafont.com/
   - 特点: 免费字体，注意许可证
   - 推荐: "Video Game" 分类

3. **1001 Fonts**
   - 地址: https://www.1001fonts.com/
   - 特点: 免费字体库
   - 推荐: 游戏风格字体

### 3. 音效资源（可选）

1. **Freesound**
   - 地址: https://freesound.org/
   - 特点: 免费音效，注意许可证
   - 类型: 按钮点击音效、UI音效等

2. **Zapsplat**
   - 地址: https://www.zapsplat.com/
   - 特点: 免费音效库（需注册）
   - 类型: 游戏音效

---

## 🏗️ 项目结构建议

### 推荐的资源目录结构

```
client/
├── Assets/
│   ├── UI/                      # UI资源
│   │   ├── Buttons/             # 按钮
│   │   ├── Icons/               # 图标
│   │   ├── Panels/              # 面板
│   │   ├── Backgrounds/         # 背景
│   │   └── Fonts/               # 字体
│   ├── Sprites/                 # 精灵图（角色、怪物等）
│   ├── Animations/              # 动画
│   ├── Audio/                   # 音频
│   │   ├── Music/               # 音乐
│   │   └── SFX/                 # 音效
│   └── Maps/                    # 地图资源
├── Scenes/                      # 场景文件
├── Scripts/                     # C#脚本
└── ...
```

---

## 📖 Godot基础操作指南

### 1. 创建新项目

1. 打开Godot
2. 点击 "New Project"
3. 选择项目路径: `C:\bgg\postapocgame\client`
4. 项目名称: `postapocgame-client`
5. **渲染器**: 选择 **Forward Plus** (推荐) 或 **Mobile** (移动平台)
6. **.NET支持**: 确保勾选 ".NET" 选项（你已安装 .NET SDK 10.0.100）
7. 点击 "Create & Edit"

**注意事项**:
- Godot 4.5 会自动检测已安装的 .NET SDK
- 如果检测不到 .NET SDK 10.0.100，检查环境变量 `PATH` 是否包含 .NET SDK 路径
- 可以在 `Editor Settings` -> `Network` -> `Languages` -> `C#` 中检查C#配置

### 2. 配置C#项目

1. 打开项目后，点击菜单: `Project` -> `Project Settings`
2. 左侧找到: `Application` -> `Config`
3. 确保 "Main Scene" 已设置（后续会创建）
4. 关闭设置窗口

5. **创建C#解决方案**:
   - 点击菜单: `Project` -> `Tools` -> `C#` -> `Create C# Solution`
   - 或者: `Project` -> `Tools` -> `C#` -> `Create .NET Solution`
   - 等待生成（会自动生成 `.csproj` 文件）
   - **注意**: Godot 4.5 默认使用 .NET 8，但 .NET 10.0 完全兼容

6. **验证C#配置**:
   - 打开命令行，在项目目录执行:
     ```bash
     cd client
     dotnet build
     ```
   - 如果编译成功，说明C#环境配置正确

7. **代码编辑器**:
   - 使用 **Visual Studio Code** (推荐，轻量级)
     - 安装扩展: `C#` (Microsoft)
     - 安装扩展: `C# Tools for Godot` (可选，提供Godot支持)
   - 或使用 **Visual Studio 2022** (功能更强大)
   - 或使用 **JetBrains Rider** (专业IDE，需付费)

8. **重要提示**:
   - .NET SDK 10.0.100 与 Godot 4.5 完全兼容
   - 项目默认使用 .NET 8.0，但 .NET 10.0 可以正常运行
   - 如需修改目标框架，编辑 `.csproj` 文件中的 `<TargetFramework>` 标签

### 3. 创建第一个场景（登录界面示例）

#### 步骤1: 创建场景文件

1. 在文件系统中，右键 `Scenes` 文件夹 -> `New` -> `Scene`
2. 命名为: `Login.tscn`
3. 双击打开场景

#### 步骤2: 设置场景根节点

1. 在场景树中，点击 "Add Node"
2. 选择: `Control` (这是UI的根节点)
3. 重命名为: `LoginUI`
4. 在右侧检查器中:
   - `Layout` -> `Full Rect` (全屏)
   - `Anchors Preset` -> `Full Rect` (全屏锚点)

#### 步骤3: 添加UI元素

1. **背景**
   - 右键 `LoginUI` -> `Add Child Node`
   - 选择: `ColorRect`
   - 重命名: `Background`
   - 在检查器中:
     - `Layout` -> `Full Rect`
     - `Color` -> 选择深色（如 `#1a1a2e`）

2. **标题标签**
   - 右键 `LoginUI` -> `Add Child Node`
   - 选择: `Label`
   - 重命名: `TitleLabel`
   - 在检查器中:
     - `Text` -> "登录游戏"
     - `Layout` -> `Top Wide` (顶部居中)
     - `Horizontal Alignment` -> `Center`
     - `Vertical Alignment` -> `Center`
     - `Theme Overrides` -> `Font Sizes` -> `Font Size` -> 48 (大字体)

3. **账号输入框容器**
   - 右键 `LoginUI` -> `Add Child Node`
   - 选择: `VBoxContainer` (垂直容器)
   - 重命名: `LoginContainer`
   - 在检查器中:
     - `Layout` -> `Center Wide` (居中)

4. **账号标签和输入框**
   - 右键 `LoginContainer` -> `Add Child Node`
   - 选择: `Label`
   - 重命名: `UsernameLabel`
   - `Text` -> "账号:"
   
   - 右键 `LoginContainer` -> `Add Child Node`
   - 选择: `LineEdit` (单行输入框)
   - 重命名: `UsernameInput`
   - `Placeholder Text` -> "请输入账号"

5. **密码标签和输入框**
   - 右键 `LoginContainer` -> `Add Child Node`
   - 选择: `Label`
   - 重命名: `PasswordLabel`
   - `Text` -> "密码:"
   
   - 右键 `LoginContainer` -> `Add Child Node`
   - 选择: `LineEdit`
   - 重命名: `PasswordInput`
   - `Placeholder Text` -> "请输入密码"
   - `Secret` -> 勾选 (隐藏密码)

6. **按钮容器**
   - 右键 `LoginContainer` -> `Add Child Node`
   - 选择: `HBoxContainer` (水平容器)
   - 重命名: `ButtonContainer`

7. **登录按钮**
   - 右键 `ButtonContainer` -> `Add Child Node`
   - 选择: `Button`
   - 重命名: `LoginButton`
   - `Text` -> "登录"

8. **注册按钮**
   - 右键 `ButtonContainer` -> `Add Child Node`
   - 选择: `Button`
   - 重命名: `RegisterButton`
   - `Text` -> "注册"

#### 步骤4: 创建C#脚本

1. 在场景树中，选中 `LoginUI` 根节点
2. 点击右侧检查器中的 "Script" 标签
3. 点击 "New Script"
4. 选择:
   - `Language`: C#
   - `Path`: `Scripts/UI/LoginUI.cs`
5. 点击 "Create"
6. 使用代码编辑器打开 `LoginUI.cs`

#### 步骤5: 编写C#脚本

```csharp
using Godot;

public partial class LoginUI : Control
{
    // UI元素引用
    private LineEdit _usernameInput;
    private LineEdit _passwordInput;
    private Button _loginButton;
    private Button _registerButton;
    
    public override void _Ready()
    {
        // 获取UI元素引用
        _usernameInput = GetNode<LineEdit>("LoginContainer/UsernameInput");
        _passwordInput = GetNode<LineEdit>("LoginContainer/PasswordInput");
        _loginButton = GetNode<Button>("LoginContainer/ButtonContainer/LoginButton");
        _registerButton = GetNode<Button>("LoginContainer/ButtonContainer/RegisterButton");
        
        // 连接按钮信号
        _loginButton.Pressed += OnLoginButtonPressed;
        _registerButton.Pressed += OnRegisterButtonPressed;
    }
    
    private void OnLoginButtonPressed()
    {
        string username = _usernameInput.Text;
        string password = _passwordInput.Text;
        
        GD.Print($"登录: {username}, {password}");
        
        // TODO: 调用网络层发送登录请求
        // NetworkManager.Instance.SendLogin(username, password);
    }
    
    private void OnRegisterButtonPressed()
    {
        string username = _usernameInput.Text;
        string password = _passwordInput.Text;
        
        GD.Print($"注册: {username}, {password}");
        
        // TODO: 调用网络层发送注册请求
        // NetworkManager.Instance.SendRegister(username, password);
    }
}
```

#### 步骤6: 保存并运行

1. 保存场景: `Ctrl + S`
2. 设置为主场景:
   - `Project` -> `Project Settings` -> `Application` -> `Run` -> `Main Scene`
   - 选择: `Scenes/Login.tscn`
3. 运行: 点击顶部播放按钮或按 `F5`

### 4. 常用UI控件说明

| 控件 | 用途 | 常用属性 |
|------|------|---------|
| `Control` | UI根节点 | `Anchors`, `Layout` |
| `Label` | 文本显示 | `Text`, `Horizontal Alignment` |
| `LineEdit` | 单行输入 | `Text`, `Placeholder Text`, `Secret` |
| `TextEdit` | 多行输入 | `Text` |
| `Button` | 按钮 | `Text`, `Pressed` 信号 |
| `CheckBox` | 复选框 | `Button Pressed`, `Checked` |
| `ColorRect` | 颜色矩形 | `Color` |
| `TextureRect` | 图片显示 | `Texture` |
| `VBoxContainer` | 垂直容器 | 自动排列子节点 |
| `HBoxContainer` | 水平容器 | 自动排列子节点 |
| `MarginContainer` | 边距容器 | 设置边距 |
| `Panel` | 面板背景 | 提供背景样式 |
| `ScrollContainer` | 滚动容器 | 内容超出时可滚动 |

### 5. UI布局技巧

#### 使用容器自动布局

1. **VBoxContainer** (垂直布局)
   ```
   VBoxContainer
   ├── Label (第1个，自动在上)
   ├── LineEdit (第2个，自动在中间)
   └── Button (第3个，自动在下)
   ```

2. **HBoxContainer** (水平布局)
   ```
   HBoxContainer
   ├── Button (第1个，自动在左)
   ├── Button (第2个，自动在中间)
   └── Button (第3个，自动在右)
   ```

3. **MarginContainer** (边距)
   ```
   MarginContainer
   └── VBoxContainer (自动添加边距)
   ```

#### 使用锚点系统

- `Anchors Preset`: 预设锚点（如 `Full Rect`、`Center Wide` 等）
- `Anchors`: 自定义锚点（0.0 - 1.0）
- `Layout`: 布局模式（自动调整大小和位置）

#### 响应式布局

- 使用 `Anchors` 而不是固定位置
- 使用 `Size Flags` 控制子节点大小
- 使用 `Theme` 统一UI样式

---

## 🎯 针对你的项目的建议

### 1. 游戏风格选择

你的项目是**仿2D横版格斗DNF游戏**，建议：

1. **像素风格** (推荐)
   - 复古感强
   - 资源易找
   - 性能好
   - 推荐资源: Kenney.nl 的像素风格UI

2. **现代风格**
   - 更精美
   - 资源更多
   - 适合商业化
   - 推荐资源: Itch.io 的现代UI包

### 2. UI设计建议

1. **主界面（HUD）**
   - 左上角: 角色信息（头像、等级、HP/MP条）
   - 底部: 快捷栏（技能、物品）
   - 右上角: 小地图
   - 左下角: 任务追踪

2. **背包界面**
   - 网格布局（10x10格子）
   - 拖拽功能
   - 物品详情悬浮显示

3. **战斗界面**
   - 伤害数字浮动显示
   - 技能冷却时间显示
   - 目标锁定显示

### 3. 开发顺序建议

1. **先搭建UI框架**
   - 创建基础UI场景
   - 学习UI布局
   - 添加基础交互

2. **再实现功能逻辑**
   - 连接网络层
   - 实现数据同步
   - 添加游戏逻辑

3. **最后优化和美化**
   - 添加动画
   - 优化UI样式
   - 添加音效

---

## 💡 实用技巧

### 1. 快速查找节点

- 在场景树中，选中节点后按 `F` 键可快速聚焦
- 使用 `GetNode<T>("路径")` 获取节点引用
- 使用 `%` 前缀可以查找场景中唯一命名的节点: `GetNode<T>("%UniqueName")`

### 2. 调试技巧

- 使用 `GD.Print()` 输出调试信息
- 使用 `GD.PrintErr()` 输出错误信息
- 使用断点调试（在VS Code或Visual Studio中）

### 3. 性能优化

- UI节点不要太多（合并相同功能的节点）
- 使用 `Visible = false` 而不是删除节点（需要时再显示）
- 使用对象池管理频繁创建/销毁的对象

### 4. 资源管理

- 使用 `ResourceLoader.Load()` 加载资源
- 使用 `preload()` 预加载常用资源
- 及时释放不用的资源

---

## 📝 学习路径总结

### 第1周: 基础入门

1. ✅ 安装Godot 4.5和.NET SDK 10.0.100 (已完成)
2. 完成官方2D游戏教程
3. 创建第一个简单场景
4. 学习UI控件和布局

### 第2周: UI开发

1. 下载免费UI资源
2. 创建登录界面
3. 创建主界面（HUD）
4. 学习C#脚本基础

### 第3周: 游戏场景

1. 创建游戏场景
2. 学习2D渲染（Sprite2D、AnimatedSprite2D）
3. 学习相机控制（Camera2D）
4. 实现简单的玩家控制

### 第4周: 功能实现

1. 连接网络层
2. 实现登录流程
3. 实现角色选择
4. 实现基础游戏逻辑

---

## 🔗 推荐资源链接

### 官方资源
- Godot官网: https://godotengine.org/
- Godot文档: https://docs.godotengine.org/zh_CN/stable/
- Godot示例: https://github.com/godotengine/godot-demo-projects
- Godot论坛: https://forum.godotengine.org/

### 学习资源
- GDQuest: https://www.gdquest.com/
- HeartBeast: https://www.youtube.com/c/uheartbeast
- KidsCanCode: https://kidscancode.org/

### 资源网站
- Kenney.nl: https://kenney.nl/assets
- OpenGameArt: https://opengameart.org/
- Itch.io: https://itch.io/game-assets/free
- Icons8: https://icons8.com/

---

## 🔧 Godot 4.5 + .NET 10.0 特定说明

### 版本兼容性

- **Godot 4.5**: 最新稳定版，完全支持C#
- **.NET SDK 10.0.100**: 最新长期支持版本，完全兼容Godot 4.5
- **建议**: 使用 .NET 8.0 作为目标框架（Godot 4.5默认），但可以用 .NET 10.0 SDK 编译

### C#项目配置

编辑 `.csproj` 文件（在项目根目录），确保配置正确：

```xml
<Project Sdk="Godot.NET.Sdk/4.5.0">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <EnableDynamicLoading>true</EnableDynamicLoading>
  </PropertyGroup>
</Project>
```

**注意**: 
- `TargetFramework` 可以是 `net8.0` 或 `net9.0`（.NET 10.0 SDK支持编译这些目标框架）
- 如果需要使用 .NET 10.0 的特性，可以改为 `net10.0`（但需要确认Godot支持）

### 常见问题排查

1. **C#脚本无法编译**
   - 检查 `.csproj` 文件是否存在
   - 在命令行运行 `dotnet build` 查看详细错误
   - 确保 .NET SDK 10.0.100 已正确安装

2. **Godot找不到C#支持**
   - 重启Godot编辑器
   - 检查 `Editor Settings` -> `Network` -> `Languages` -> `C#` 配置
   - 确保勾选了 ".NET" 选项

3. **C#智能提示不工作**
   - VS Code: 安装 "C#" 扩展，重启编辑器
   - Visual Studio: 确保安装了 ".NET 桌面开发" 工作负载

---

## 💬 常见问题

### Q1: GDScript vs C#，我应该用哪个？

**A**: 对于你的项目（C#开发），建议：
- **核心逻辑用C#**: 网络层、游戏逻辑、数据层
- **简单脚本用GDScript**: UI信号处理、场景切换等简单逻辑（可选）

但既然你已经选择了C#，就全部用C#也可以，性能差异不大。

**注意**: 使用 .NET SDK 10.0.100 和 Godot 4.5，C#性能很好，可以全部使用C#开发。

### Q2: 2D游戏用什么渲染方式？

**A**: Godot 4.5中，2D游戏默认使用Canvas渲染，性能很好，无需特殊配置。

**渲染器选择**:
- **Forward Plus**: 推荐用于大部分项目（2D/3D通用）
- **Mobile**: 移动平台优化，2D游戏也适用
- **Compatibility**: 兼容模式，不推荐

### Q3: UI资源怎么导入到Godot？

**A**:
1. 将资源文件（PNG等）拖到Godot的文件系统面板
2. 或者复制到 `client/Assets/UI/` 目录
3. 在编辑器中自动识别并导入
4. 在检查器中可以直接拖拽使用

### Q4: 如何实现UI动画？

**A**: 使用Godot的Tween系统:
```csharp
var tween = CreateTween();
tween.TweenProperty(this, "position", targetPosition, 0.5f);
```

### Q5: 如何适配不同分辨率？

**A**:
1. 使用 `Anchors` 和 `Layout` 而不是固定坐标
2. 设置 `Project Settings` -> `Display` -> `Window` -> `Stretch` -> `Mode` 为 `viewport`
3. 使用 `Control` 的 `Size Flags` 控制自适应

---

## 🎉 开始你的第一个场景

建议你先按照上面的步骤，创建一个简单的登录界面，熟悉Godot的基本操作。遇到问题可以在：
- Godot官方论坛提问
- B站搜索相关教程
- 查看官方文档

**祝你开发顺利！** 🚀
