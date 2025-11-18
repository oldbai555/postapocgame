# Godot客户端开发进度文档

## 📋 项目概述

**项目名称**: postapocgame（后启示录游戏）  
**客户端引擎**: Godot  
**开发语言**: C#  
**游戏类型**: 仿2D横版格斗DNF游戏  
**目标平台**: Windows、安卓、iOS  
**服务端架构**: Gateway + GameServer + DungeonServer  
**开发模式**: 个人开发

---

## ✅ 已完成功能

### Phase 1: 基础框架 ✅

#### 1.1 项目初始化 ✅
**当前状态**: 已完成

**已完成内容**:
- [x] 创建Godot项目
- [x] 配置C#开发环境（.NET 8.0）
- [x] 设置项目目录结构（Scripts/Network, Scripts/Protocol, Scripts/Utils, Scripts/Data等）
- [x] 添加Google.Protobuf NuGet包依赖

**实现细节**:
- 项目使用Godot 4.5.1，C# .NET 8.0
- 已创建完整的目录结构：Network（网络层）、Protocol（协议层）、Utils（工具层）、Data（数据层）、Scene（场景层）、GameLogic（游戏逻辑层）、UI（界面层）
- 在`client.csproj`中添加了`Google.Protobuf`包引用

**位置**: `client/`

#### 1.2 Protocol Buffer协议生成 ✅
**当前状态**: 已完成

**已完成内容**:
- [x] 安装Protocol Buffers C#生成工具（使用proto目录下的protoc.exe）
- [x] 配置proto文件路径（proto/csproto/）
- [x] 创建协议生成脚本（genproto_csharp.bat和genproto_csharp.sh）
- [x] 创建协议生成说明文档

**实现细节**:
- 创建了Windows批处理脚本`proto/genproto_csharp.bat`和Linux脚本`proto/genproto_csharp.sh`
- 脚本会将`proto/csproto/*.proto`文件生成C#代码到`client/Scripts/Protocol/`目录
- 添加了`client/Scripts/Protocol/README.md`说明文档，包含生成步骤和协议映射注册示例
- 注意：协议代码需要手动运行脚本生成，生成后需要在`ProtocolHandler.cs`中注册协议映射

**位置**: `proto/`、`client/Scripts/Protocol/`

#### 1.3 网络层框架 ✅
**当前状态**: 已完成

**已完成内容**:
- [x] 网络连接管理（`NetworkManager.cs`）
- [x] 协议消息封装/解析（`ProtocolHandler.cs`）
- [x] 消息发送/接收框架（`MessageSender.cs`、`MessageReceiver.cs`）
- [x] 心跳机制
- [x] 断线重连机制

**实现细节**:
- **NetworkManager.cs**: 
  - 实现TCP连接管理，支持异步连接
  - 实现心跳机制（每5秒发送一次，消息类型0x06）
  - 实现自动断线重连（最多重试5次，间隔3秒）
  - 提供服务器时间同步接口（待后续实现Ping/Pong时间同步）
  - 单例模式，通过Godot Node管理生命周期
- **ProtocolHandler.cs**: 
  - 实现消息序列化/反序列化
  - 支持服务端消息格式：`[4字节长度][1字节类型][消息体]`
  - 客户端消息格式：`[2字节MsgId][数据]`，包装在ForwardMessage中
  - 提供协议映射注册机制（C2S和S2C协议分别注册）
- **MessageSender.cs**: 
  - 提供便捷的消息发送接口
  - 支持SessionId参数（登录前为空字符串）
- **MessageReceiver.cs**: 
  - 在后台线程接收消息，避免阻塞主线程
  - 使用消息队列在主线程处理，保证线程安全
  - 支持协议处理器注册机制
  - 自动处理心跳消息（类型0x06）
- **Logger.cs**: 
  - 提供统一的日志工具类，支持不同日志级别

**位置**: `client/Scripts/Network/`

---

### Phase 2: 登录与角色管理 ✅

#### 2.1 登录界面 ✅
**当前状态**: 已完成

**已完成内容**:
- [x] 登录界面逻辑（`LoginUI.cs`）
- [x] 账号注册功能
- [x] 账号登录功能
- [x] 记住账号功能（使用LocalStorage）
- [x] 错误提示显示
- [x] 自动连接服务器

**实现细节**:
- **LoginUI.cs**: 
  - 实现账号注册和登录功能
  - 支持记住账号和密码（使用LocalStorage加密存储）
  - 自动连接服务器，连接失败时显示错误
  - 登录成功后自动切换到角色选择界面
  - 完整的错误处理和状态提示
- **LocalStorage.cs**: 
  - 提供本地数据存储功能
  - 支持账号信息加密存储（Base64编码，实际项目中应使用更安全的加密）
  - 使用Godot的用户数据目录存储
- **协议注册**: 在`ProtocolHandler`中注册了所有登录相关协议

**相关协议**:
- `C2SRegister` / `S2CRegisterResult` ✅
- `C2SLogin` / `S2CLoginResult` ✅

**位置**: `client/Scripts/UI/LoginUI.cs`、`client/Scripts/Data/LocalStorage.cs`、`client/Scenes/Login.tscn`

**注意**: 登录场景（`Login.tscn`）已创建完成，参考 `client/Scenes/README.md`

#### 2.2 角色选择界面 ✅
**当前状态**: 已完成

**已完成内容**:
- [x] 角色选择界面逻辑（`RoleSelectUI.cs`）
- [x] 角色列表显示（角色名称、职业、等级等）
- [x] 创建角色功能
- [x] 进入游戏功能
- [x] 角色详细信息显示

**实现细节**:
- **RoleSelectUI.cs**: 
  - 实现角色列表查询和显示
  - 支持角色选择（点击角色按钮）
  - 实现创建角色功能（使用AcceptDialog创建角色对话框，支持输入角色名、选择职业和性别）
  - 实现进入游戏功能
  - 显示角色详细信息（名称、职业、性别、等级）
  - 完整的错误处理和状态提示
- **RoleSelect.tscn**: 
  - 已创建角色选择场景文件
  - 使用CenterContainer居中布局
  - 包含角色列表容器、角色信息显示、创建/进入/删除按钮等UI元素
  - 脚本已附加到场景根节点
- **协议处理**: 
  - 正确处理`S2CRoleList`、`S2CCreateRoleResult`、`S2CLoginSuccess`等协议
  - 错误通过`S2CError`协议统一处理

**相关协议**:
- `C2SQueryRoles` / `S2CRoleList` ✅
- `C2SCreateRole` / `S2CCreateRoleResult` ✅
- `C2SEnterGame` / `S2CLoginSuccess` ✅

**位置**: `client/Scripts/UI/RoleSelectUI.cs`、`client/Scenes/RoleSelect.tscn`

**待完善**:
- 删除角色功能（需要实现确认对话框）

---

### Phase 3: 游戏场景系统（进行中）⚠️

#### 3.0 场景切换Loading界面 ✅
**当前状态**: 已完成

**已完成内容**:
- [x] 创建通用加载界面 `LoadingScreen.tscn`，包含简易加载图、提示文字、进度条与百分比
- [x] 编写 `LoadingScreen.cs` 控制脚本，实现显示/隐藏与进度更新
- [x] 新增 `SceneManager.cs`，封装异步场景切换与加载进度回调
- [x] 在 `AutoLoad.tscn` 中注册 `SceneManager` 与 `LoadingScreen`，并让 `LoginUI` 切换至角色选择时使用新的加载流程

**实现细节**:
- `SceneManager.SwitchSceneAsync()` 通过 `ResourceLoader.LoadThreadedRequest()` 异步加载场景，实时向 Loading UI 推送进度
- 场景切换期间显示半透明遮罩、图标与百分比提示，加载完成后自动隐藏
- 若 `SceneManager` 未初始化，自动回退到原有的 `ChangeSceneToFile` 逻辑，避免阻塞开发流程

**位置**: `client/Scripts/Scene/SceneManager.cs`、`client/Scenes/LoadingScreen.tscn`、`client/Scripts/UI/LoadingScreen.cs`、`client/Scenes/AutoLoad.tscn`

**注意**:
- 目前仅在登录 → 角色选择流程中启用，后续切场景时请统一通过 `SceneManager` 调用
- Loading UI 预留文案参数，可按不同场景切换需求传入提示文本

#### 3.1 场景管理 ⚠️
**当前状态**: 进行中（基础Loading流程已接入）

**需要完成**:
- [x] 场景管理器基础能力（`SceneManager.cs`，异步加载 + Loading UI）
- [ ] 场景加载/卸载的资源回收策略
- [ ] 场景切换动画 / 过渡特效
- [ ] 场景配置读取
- [ ] 游戏主场景（`Main.tscn`）- 进入游戏后的主场景，包含游戏世界和主界面UI

**相关协议**:
- `S2CEnterScene` - 进入场景
- `C2SChangeScene` / `S2CChangeSceneResult` - 切换场景

**位置**: `client/Scripts/Scene/SceneManager.cs`、`client/Scenes/Main.tscn`

**注意**: 游戏主场景（`Main.tscn`）需要创建，参考 `client/Scenes/README.md`

#### 3.2 实体管理 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 实体管理器（`EntityManager.cs`）
- [ ] 基础实体类（`Entity.cs`）
- [ ] 玩家实体（`PlayerEntity.cs`）
- [ ] 怪物实体（`MonsterEntity.cs`）
- [ ] NPC实体（`NPCEntity.cs`）
- [ ] 掉落物实体（`DropItemEntity.cs`）
- [ ] 实体创建/销毁
- [ ] 实体位置同步
- [ ] AOI视野管理（实体进入/离开视野）

**相关协议**:
- `S2CEntityAppear` - 实体进入视野
- `S2CEntityDisappear` - 实体离开视野
- `S2CEntityMove` - 实体移动
- `S2CEntityStopMove` - 实体停止移动

**位置**: `client/Scripts/Scene/EntityManager.cs`、`client/Scripts/Scene/Entity.cs`

#### 3.3 场景渲染 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 2D横版地图渲染
- [ ] 实体精灵渲染
- [ ] 实体动画播放
- [ ] 血条/名字显示
- [ ] 伤害数字显示
- [ ] 技能特效渲染

**位置**: `client/Scripts/Scene/`

#### 3.x 调试注意事项 ⚠️
**当前状态**: 持续补充中

**新手易错项**:
- 角色 Sprite 必须与 `CollisionShape2D` 对齐，否则看起来像“穿墙”。建议在 Godot 中开启 `Debug -> Visible Collision Shapes`，确保视觉 Sprite 正好覆盖碰撞形状，避免 Sprite 偏移到碰撞之外导致误判。

---

### Phase 4: 游戏逻辑系统（未开始）⚠️

#### 4.1 玩家控制 ⚠️
**当前状态**: 进行中（已完成基础状态机 + 攻击状态）

**已完成内容**:
- [x] 玩家基础状态机（`Player.cs` + `Player_state_machine.cs` + `StateIdle.cs` / `StateWalk.cs`），负责输入、动画切换、Idle/Walk移动
- [x] 攻击状态（`StateAttack.cs`），负责攻击动画播放和状态切换

**实现细节**:
- **StateAttack.cs**: 
  - 实现攻击状态逻辑，播放攻击动画（`attack_up` / `attack_down` / `attack_side`）
  - 支持攻击特效动画播放（`AttackEffectAnimationPlayer`），路径：`Sprite2D/AttackEffectSprite/AttackEffectAnimationPlayer`
  - 攻击特效动画根据当前方向自动选择（`attack_down` / `attack_up` / `attack_side`），并结合 `EffectOffset*` 在左右翻转时镜像位置，保证表现自然
  - 支持攻击音效播放（`Audio/AudioStreamPlayer2D`），可配置音效资源、基准音调、随机音调偏移
  - 攻击期间直接锁定玩家速度为 0，并在动画完成后根据最新输入切回 Idle / Walk，逻辑简单明了
  - 攻击状态进入时会重新采样方向/状态，并同步特效的 FlipH 与偏移，确保左/右攻击动画与特效方向正确
  - 通过 `AnimationPlayer.AnimationFinished` 信号监听动画完成
  - 攻击完成后根据输入方向自动切换到 `Idle` 或 `Walk` 状态
  - 攻击状态下速度设为 0，保持角色静止
  - 在 `StateIdle` 和 `StateWalk` 中通过 `HandleInput` 监听攻击按键（`attack`），可切换到攻击状态
  - 遵循代码风格：早返回、减少嵌套、清晰的变量命名
  - 添加了完善的 null 检查和错误日志，避免节点缺失导致的崩溃
- **Player.cs**:
  - 新增 `AnimDirection()` 方法，返回当前方向对应的动画方向名称（`down` / `up` / `side`）
  - 该方法用于统一获取方向名称，供状态机和其他系统使用
  - 角色左右朝向通过修改 `Sprite2D.Scale.X` 控制，负值代表面向左，保证 `AttackEffectSprite` 等子节点自动镜像
- **StateAttack.cs / State.cs / Player.cs（2025-11-19）**:
  - `_attackEffectSprite` 会在进入攻击时缓存初始位置，并依据 `Player.IsFacingLeft` 自动镜像偏移（`EffectOffset*`），即便特效动画本身包含位移轨迹也能自然反向
  - 状态基类新增 `RefreshMovementAndAnimation()` 与 `TryEnterAttack()`，Idle/Walk 等状态通过共享工具刷新方向、状态和攻击输入逻辑，减少重复代码
  - 攻击状态保留 `DecelerateSpeed` 可调减速逻辑，攻击开始后速度按系数逐帧衰减，既能保留“按下时有惯性”，又能保证动画结束时自然收脚
- **Player.cs / StateAttack.cs 优化（2025-11-19）**:
  - Player 仅保留输入解析、朝向和动画选择的必要逻辑，`UpdateAnimation` 直接复用 `AnimDirection()`，不再堆额外日志
  - StateAttack 在 `_Ready()` 中一次性缓存所有节点与 AnimationFinished 信号，进入状态时只做“停速→播放动画/特效/音效”的直观流程
  - 攻击结束由动画信号单点决定，完成后根据最新输入切回 Idle / Walk，便于服务端同学快速理解整条“按键→动画→落地状态”链路

**需要完成**:
- [ ] 玩家控制器（`PlayerController.cs`）
- [ ] 输入处理（键盘/触屏）
- [ ] 角色移动控制
- [ ] 技能释放控制
- [ ] 交互控制（拾取、对话等）

**位置**: `client/Player/Scripts/StateAttack.cs`、`client/Scripts/GameLogic/PlayerController.cs`

#### 4.2 移动系统 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 移动系统（`MovementSystem.cs`）
- [ ] 移动输入处理
- [ ] 移动动画播放（基于服务器时间戳对齐移动开始/结束，减少客户端本地时钟偏差）
- [ ] 移动同步到服务端
- [ ] 服务端移动数据接收和同步

**相关协议**:
- `C2SStartMove` - 开始移动
- `C2SUpdateMove` - 移动中更新
- `C2SEndMove` - 结束移动
- `S2CEntityMove` - 实体移动广播
- `S2CEntityStopMove` - 实体停止移动广播

**位置**: `client/Scripts/GameLogic/MovementSystem.cs`

#### 4.3 战斗系统 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 战斗系统（`CombatSystem.cs`）
- [ ] 技能释放逻辑
- [ ] 技能动画播放（支持前摇→判定→后摇的帧定义表现）
- [ ] 伤害数字显示
- [ ] 受击动画播放
- [ ] 战斗状态管理
- [ ] Buff/Debuff显示

**相关协议**:
- `C2SUseSkill` - 使用技能
- `S2CSkillCastResult` - 技能释放结果
- `C2SGetNearestMonster` / `S2CGetNearestMonsterResult` - 获取最近怪物

**位置**: `client/Scripts/GameLogic/CombatSystem.cs`

#### 4.4 技能系统 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 技能系统（`SkillSystem.cs`）
- [ ] 技能列表管理
- [ ] 技能CD显示
- [ ] 技能图标显示
- [ ] 技能详情显示

**相关协议**:
- `C2SLearnSkill` / `S2CLearnSkillResult` - 学习技能
- `C2SUpgradeSkill` / `S2CUpgradeSkillResult` - 升级技能

**位置**: `client/Scripts/GameLogic/SkillSystem.cs`

#### 4.5 自动战斗AI（可选）⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 自动战斗AI（`AutoCombatAI.cs`）
- [ ] 自动选择目标
- [ ] 自动释放技能
- [ ] 自动拾取掉落物
- [ ] 自动使用药水（HP/MP低于阈值）

**注意**: 自动战斗AI在客户端实现，但所有操作都需要发送到服务端进行校验

**位置**: `client/Scripts/GameLogic/AutoCombatAI.cs`

---

### Phase 5: UI系统（未开始）⚠️

#### 5.1 主界面（HUD）⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 游戏主场景文件（`Main.tscn`）- 游戏主场景，包含游戏世界和主界面UI
- [ ] 主界面UI脚本（`MainUI.cs`）
- [ ] 角色信息显示（头像、等级、HP/MP条）
- [ ] 快捷栏（技能、物品）
- [ ] 任务追踪显示
- [ ] 小地图
- [ ] 系统按钮（背包、装备、技能、任务、邮件、设置等）

**位置**: `client/Scenes/Main.tscn`、`client/Scripts/UI/MainUI.cs`

**注意**: 
- 游戏主场景（`Main.tscn`）需要创建，参考 `client/Scenes/README.md`
- 主场景应包含游戏世界节点和主界面UI节点
- 主界面UI可以作为场景的子节点或独立场景

---

### Phase 6: 时间同步与动画同步（未开始）⚠️

#### 6.1 时间同步模块 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 在 Network 层实现 Ping/Pong 时间同步逻辑（定期向 Gateway 请求服务器时间）
- [ ] 计算本地与服务器时间差（偏移量），提供统一的“服务器时间”访问接口
- [ ] 在 Debug HUD 中显示当前本地时间与服务器时间差，便于调试

**位置**: `client/Scripts/Network/NetworkManager.cs`

#### 6.2 动画同步集成 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 在 MovementSystem 中使用服务器时间驱动移动动画（开始/停止时刻以服务器时间为准）
- [ ] 在 CombatSystem 中实现技能帧时间线（前摇/判定/后摇），并以服务器时间戳对齐动画与伤害出现时机
- [ ] 根据服务器广播的战斗/状态协议更新本地实体的动画状态（施法、受击、倒地等）

**位置**: `client/Scripts/GameLogic/MovementSystem.cs`、`client/Scripts/GameLogic/CombatSystem.cs`

#### 5.2 背包界面 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 背包界面UI（`BagUI.tscn`、`BagUI.cs`）
- [ ] 背包格子显示
- [ ] 物品图标/数量显示
- [ ] 物品详情显示（悬停显示）
- [ ] 物品操作（使用、装备、回收）
- [ ] 背包整理功能（客户端排序）

**相关协议**:
- `C2SOpenBag` / `S2CBagData` - 查询背包
- `S2CUpdateBagData` - 更新背包数据
- `C2SUseItem` / `S2CUseItemResult` - 使用物品
- `C2SRecycleItem` / `S2CRecycleItemResult` - 回收物品

**位置**: `client/Scenes/UI/BagUI.tscn`、`client/Scripts/UI/BagUI.cs`

#### 5.3 装备界面 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 装备界面UI（`EquipUI.tscn`、`EquipUI.cs`）
- [ ] 装备槽位显示（武器、防具等）
- [ ] 装备详情显示（属性、强化等级等）
- [ ] 装备穿脱功能
- [ ] 装备强化/精炼/附魔界面（可选）

**相关协议**:
- `C2SEquipItem` / `S2CEquipResult` - 穿戴装备

**位置**: `client/Scenes/UI/EquipUI.tscn`、`client/Scripts/UI/EquipUI.cs`

#### 5.4 任务界面 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 任务界面UI（`QuestUI.tscn`、`QuestUI.cs`）
- [ ] 任务列表显示（进行中、可接取、已完成）
- [ ] 任务详情显示（目标、进度、奖励）
- [ ] 任务追踪功能（自动导航到目标位置）
- [ ] 任务接取/提交功能

**相关协议**:
- `C2STalkToNPC` / `S2CTalkToNPCResult` - 和NPC对话
- `S2CQuestData` - 任务数据

**位置**: `client/Scenes/UI/QuestUI.tscn`、`client/Scripts/UI/QuestUI.cs`

#### 5.5 技能界面 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 技能界面UI（`SkillUI.tscn`、`SkillUI.cs`）
- [ ] 技能列表显示（技能图标、等级、CD等）
- [ ] 技能详情显示（伤害、消耗、描述等）
- [ ] 技能学习/升级功能

**位置**: `client/Scenes/UI/SkillUI.tscn`、`client/Scripts/UI/SkillUI.cs`

#### 5.6 邮件界面 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 邮件界面UI（`MailUI.tscn`、`MailUI.cs`）
- [ ] 邮件列表显示（未读、已读、已领取）
- [ ] 邮件详情显示（标题、内容、附件）
- [ ] 邮件操作（读取、领取附件、删除）

**相关协议**:
- `S2CMailList` - 邮件列表
- `S2CMailDetail` - 邮件详情

**位置**: `client/Scenes/UI/MailUI.tscn`、`client/Scripts/UI/MailUI.cs`

#### 5.7 商城界面 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 商城界面UI（`ShopUI.tscn`、`ShopUI.cs`）
- [ ] 商品列表显示（物品、价格）
- [ ] 购买功能
- [ ] 货币显示

**相关协议**:
- `C2SShopBuy` / `S2CShopBuyResult` - 商城购买
- `C2SOpenMoney` / `S2CMoneyData` - 查询货币

**位置**: `client/Scenes/UI/ShopUI.tscn`、`client/Scripts/UI/ShopUI.cs`

#### 5.8 副本界面 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 副本界面UI（`DungeonUI.tscn`、`DungeonUI.cs`）
- [ ] 副本列表显示（副本信息、难度、奖励预览）
- [ ] 副本推荐（根据玩家等级推荐）
- [ ] 进入副本功能
- [ ] 副本结算界面（显示用时、击杀数、获得奖励等）

**相关协议**:
- `C2SEnterDungeon` / `S2CEnterDungeonResult` - 进入副本
- `C2SChangeScene` / `S2CChangeSceneResult` - 切换场景
- `C2SClaimOfflineReward` / `S2CClaimOfflineRewardResult` - 领取离线收益

**位置**: `client/Scenes/UI/DungeonUI.tscn`、`client/Scripts/UI/DungeonUI.cs`

#### 5.9 设置界面 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 设置界面UI（`SettingsUI.tscn`、`SettingsUI.cs`）
- [ ] 音量设置
- [ ] 画面设置
- [ ] 按键设置
- [ ] 自动战斗设置

**位置**: `client/Scenes/UI/SettingsUI.tscn`、`client/Scripts/UI/SettingsUI.cs`

---

### Phase 6: 数据层（未开始）⚠️

#### 6.1 配置表管理 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 配置表管理器（`ConfigManager.cs`）
- [ ] JSON配置表加载
- [ ] 配置表缓存
- [ ] 配置表热更新（可选）

**位置**: `client/Scripts/Data/ConfigManager.cs`

#### 6.2 本地数据存储 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 本地数据存储（`LocalStorage.cs`）
- [ ] 账号信息存储（加密）
- [ ] 设置数据存储
- [ ] 数据持久化

**位置**: `client/Scripts/Data/LocalStorage.cs`

#### 6.3 玩家数据缓存 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 玩家数据缓存（`PlayerDataCache.cs`）
- [ ] 角色数据缓存
- [ ] 背包数据缓存
- [ ] 任务数据缓存
- [ ] 数据同步机制

**位置**: `client/Scripts/Data/PlayerDataCache.cs`

---

### Phase 7: 工具层（未开始）⚠️

#### 7.1 日志系统 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 日志系统（`Logger.cs`）
- [ ] 日志级别管理
- [ ] 日志文件输出
- [ ] 日志格式化

**位置**: `client/Scripts/Utils/Logger.cs`

#### 7.2 资源管理 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 资源管理器（`ResourceManager.cs`）
- [ ] 资源预加载
- [ ] 资源缓存
- [ ] 资源释放

**位置**: `client/Scripts/Utils/ResourceManager.cs`

#### 7.3 工具函数 ⚠️
**当前状态**: 未开始

**需要完成**:
- [ ] 工具函数（`UtilityFunctions.cs`）
- [ ] 数学工具函数
- [ ] 字符串工具函数
- [ ] 时间工具函数

**位置**: `client/Scripts/Utils/UtilityFunctions.cs`

---

## 🚧 待实现功能

### Phase 8: 高级功能（待实现）

#### 8.1 动画同步协议 ⚠️
**当前状态**: 未实现

**需要实现**:
- 技能释放动画同步
- 受击动画同步
- 移动动画同步
- 状态切换动画同步

#### 8.2 战斗统计 ⚠️
**当前状态**: 未实现

**需要实现**:
- 战斗统计显示（总伤害、击杀数、连击数等）
- 战斗结算界面优化
- 战斗回放功能（可选）

#### 8.3 快捷栏系统 ⚠️
**当前状态**: 未实现

**需要实现**:
- 快捷栏UI
- 快捷栏配置（保存到本地或发送到服务端）
- 快捷栏使用

#### 8.4 背包整理系统 ⚠️
**当前状态**: 未实现

**需要实现**:
- 一键整理功能
- 物品排序逻辑（按类型、品质、等级等）
- 自动堆叠相同物品

#### 8.5 任务导航系统 ⚠️
**当前状态**: 未实现

**需要实现**:
- 任务目标自动导航
- 任务目标高亮显示
- 小地图任务标记

---

## 📋 开发优先级总结

### 🔥 紧急（必须尽快完成）

1. **项目初始化** - 创建Godot项目，配置开发环境
2. **Protocol Buffer协议生成** - 生成C#协议代码
3. **网络层框架** - 实现网络连接和消息处理
4. **登录流程** - 登录界面和角色选择界面

### ⭐ 重要（核心功能）

5. **场景系统** - 场景管理和实体管理
6. **移动系统** - 玩家移动控制
7. **战斗系统** - 技能释放和战斗逻辑
8. **主界面（HUD）** - 游戏主界面
9. **背包界面** - 物品管理

### 📌 中优先级（功能完善）

10. **装备界面** - 装备管理
11. **任务界面** - 任务管理
12. **技能界面** - 技能管理
13. **邮件界面** - 邮件管理
14. **商城界面** - 商城购买
15. **副本界面** - 副本选择和管理

### 🎯 低优先级（功能优化）

16. **自动战斗AI** - 自动战斗功能
17. **背包整理** - 一键整理功能
18. **快捷栏** - 快捷栏功能
19. **任务导航** - 任务自动导航
20. **战斗统计** - 战斗数据统计

---

## 📁 关键代码位置

### 网络层
- `client/Scripts/Network/NetworkManager.cs` - 网络管理器（✅ 已完成）
- `client/Scripts/Network/ProtocolHandler.cs` - 协议处理（✅ 已完成）
- `client/Scripts/Network/MessageSender.cs` - 消息发送（✅ 已完成）
- `client/Scripts/Network/MessageReceiver.cs` - 消息接收（✅ 已完成）
- `client/Scripts/Utils/Logger.cs` - 日志工具（✅ 已完成）

### 游戏场景层
- `client/Scripts/Scene/SceneManager.cs` - 场景管理器（✅ 初版完成，负责异步切换与Loading UI）
- `client/Scripts/Scene/EntityManager.cs` - 实体管理器（待创建）
- `client/Scripts/Scene/Entity.cs` - 基础实体（待创建）
- `client/Scripts/Scene/PlayerEntity.cs` - 玩家实体（待创建）
- `client/Scripts/Scene/MonsterEntity.cs` - 怪物实体（待创建）

### 游戏逻辑层
- `client/Player/Scripts/Player.cs` - Player宿主节点（输入采集、动画拼装、状态机路径配置、AnimDirection方法）
- `client/Player/Scripts/Player_state_machine.cs` - 玩家状态机（调度Idle/Walk/Attack等状态）
- `client/Player/Scripts/State.cs` - 状态基类（统一Player注入、`RefreshMovementAndAnimation`、`TryEnterAttack` 等共享工具方法）
- `client/Player/Scripts/StateIdle.cs` - Idle状态（静止、方向采集、进入Walk/Attack）
- `client/Player/Scripts/StateWalk.cs` - Walk状态（移动速度计算、动画播放、进入Attack）
- `client/Player/Scripts/StateAttack.cs` - Attack状态（攻击动画/特效/音效播放、攻击减速、完成后切换回Idle/Walk，`_attackEffectSprite` 通过 `Player.IsFacingLeft` 与主体朝向保持一致）
- `client/Scripts/GameLogic/PlayerController.cs` - 玩家控制器（待创建）
- `client/Scripts/GameLogic/MovementSystem.cs` - 移动系统（待创建）
- `client/Scripts/GameLogic/CombatSystem.cs` - 战斗系统（待创建）
- `client/Scripts/GameLogic/SkillSystem.cs` - 技能系统（待创建）
- `client/Scripts/GameLogic/AutoCombatAI.cs` - 自动战斗AI（待创建）

### UI层
- `client/Scripts/UI/LoginUI.cs` - 登录界面（✅ 已完成）
- `client/Scripts/UI/RoleSelectUI.cs` - 角色选择界面（✅ 已完成）
- `client/Scripts/UI/LoadingScreen.cs` - Loading界面（✅ 已完成，用于场景切换）
- `client/Scenes/LoadingScreen.tscn` - Loading界面场景（✅ 已完成）
- `client/Scripts/UI/MainUI.cs` - 主界面（待创建）
- `client/Scripts/UI/BagUI.cs` - 背包界面（待创建）
- `client/Scripts/UI/EquipUI.cs` - 装备界面（待创建）
- `client/Scripts/UI/QuestUI.cs` - 任务界面（待创建）
- `client/Scripts/UI/SkillUI.cs` - 技能界面（待创建）
- `client/Scripts/UI/MailUI.cs` - 邮件界面（待创建）
- `client/Scripts/UI/ShopUI.cs` - 商城界面（待创建）
- `client/Scripts/UI/DungeonUI.cs` - 副本界面（待创建）
- `client/Scripts/UI/SettingsUI.cs` - 设置界面（待创建）

### 数据层
- `client/Scripts/Data/ConfigManager.cs` - 配置表管理器（待创建）
- `client/Scripts/Data/LocalStorage.cs` - 本地数据存储（✅ 已完成）
- `client/Scripts/Data/PlayerDataCache.cs` - 玩家数据缓存（待创建）

### 协议层
- `client/Scripts/Protocol/` - 协议代码（✅ 已生成）
  - `Cs.cs` - 客户端到服务端协议
  - `Sc.cs` - 服务端到客户端协议
  - `Base.cs` - 基础数据结构
  - `Player.cs` - 玩家相关数据结构

---

## 🔧 开发注意事项与架构决策

### 架构原则

1. **模块化设计**: 各个模块独立，通过接口通信
2. **事件驱动**: 使用Godot的信号系统解耦模块间通信
3. **单例模式**: 网络管理器、场景管理器等使用单例模式
4. **资源管理**: 合理管理游戏资源，避免内存泄漏
5. **统一场景切换流程**: 场景切换统一走 `SceneManager`，确保Loading界面与异步加载逻辑一致

### 网络通信

1. **异步处理**: 所有网络操作使用异步方式，避免阻塞主线程
2. **消息队列**: 使用消息队列管理网络消息（MessageReceiver在主线程处理消息队列）
3. **断线重连**: 实现完善的断线重连机制（最多重试5次，间隔3秒）
4. **心跳机制**: 定期发送心跳包（每5秒），保持连接活跃，消息类型0x06
5. **线程安全**: 消息接收在后台线程，处理在主线程，通过消息队列保证线程安全
6. **SessionId管理**: 客户端发送消息时SessionId可为空（登录前），Gateway会自动创建Session

### 性能优化

1. **对象池**: 对于频繁创建/销毁的对象（如伤害数字、特效），使用对象池
2. **资源预加载**: 预加载常用资源，减少运行时加载
3. **批量渲染**: 合并相同类型的渲染，减少DrawCall
4. **LOD系统**: 根据距离使用不同精度的模型/贴图

### 协议处理

1. **协议验证**: 接收协议后先验证数据合法性
2. **错误处理**: 完善的错误处理和提示
3. **协议缓存**: 对于频繁的协议，可以考虑客户端缓存

### UI开发

1. **响应式设计**: UI适配不同分辨率
2. **UI复用**: 相同类型的UI组件复用
3. **动画优化**: UI动画使用Godot的Tween系统，性能更好

### 玩家状态机

1. **节点要求**: `Player` 需要包含 `Player_state_machine` 子节点，或在 `Player.StateMachinePath` 中显式配置路径，否则输入/状态逻辑不会执行
2. **状态扩展**: 任何继承自 `State` 的节点只要挂在状态机下，就会被自动注入 `Player` 并可在 `Process/Physics/HandleInput` 中返回下一个状态
3. **动画前缀**: 状态内调用 `Player.UpdateAnimation("idle" | "walk" | "attack" | …)`，由状态自身决定动画前缀，`Player` 负责根据主方向拼接 `*_up/_down/_side`
4. **信号连接**: 状态中如需监听 `AnimationPlayer` 信号（如 `AnimationFinished`），应在 `Enter()` 中连接，在 `Exit()` 中断开，避免内存泄漏
5. **节点查找**: 状态节点位于 `StateMachine` 下，如需访问 `Player` 的直接子节点（如 `AnimationPlayer`），需要通过 `GetParent().GetParent()` 向上查找
6. **攻击特效**: 攻击状态支持播放攻击特效动画，通过 `AttackEffectAnimationPlayer` 节点控制，路径为 `Sprite2D/AttackEffectSprite/AttackEffectAnimationPlayer`，特效动画命名格式为 `attack_{direction}`（direction 为 `down` / `up` / `side`）
7. **攻击音效**: 攻击状态通过 `Audio/AudioStreamPlayer2D` 播放音效，可在 `StateAttack` 中配置音效资源与音调范围（`AttackSound`、`AttackPitchBase`、`AttackPitchRandomRange`），确保音效节点命名一致
8. **方向获取**: 使用 `Player.AnimDirection()` 方法统一获取当前方向对应的动画方向名称，避免在多个地方重复实现方向判断逻辑
9. **日志辅助**: `Player` 和状态机在初始化失败时会输出错误日志，运行期可通过Godot输出快速排查
10. **特效与主体同源判定**: `_attackEffectSprite` 的左右翻转务必复用 `Player.IsFacingLeft`（内部读取 `Sprite2D.Scale.X` 是否为负），保持特效与主体动画完全同步
11. **Scale镜像注意事项**: 通过 `Sprite2D.Scale.X` 调整朝向会连带碰撞、特效等子节点一起镜像，若碰撞形状出现反向问题需在 Godot 中确认其父级层次；必要时可将碰撞与视觉拆分
12. **重复逻辑统一封装**: `client/Player/Scripts` 下的共性逻辑（节点查找、特效播放、动画方向判断等）优先提取到 `Player` 或共享辅助方法，避免各状态脚本重复编写
13. **攻击动画不可循环**: `player.tscn` 中主 `AnimationPlayer` 的攻击动作为了依赖 `AnimationFinished` 信号完成状态切换，`loop_mode` 必须保持为 `0 (Loop None)`；若设置为循环（例如 `loop_mode = 2`），信号不会触发，`StateAttack` 会一直停留在攻击状态。

### 客户端脚本维护原则（2025-11-19）

1. **禁止误删表现逻辑**: 优化脚本时务必保留方向采样、`_cardinalDirection` 更新及 `_sprite.Scale` 镜像规则，防止再次出现“按 ↑/↓ 却只播放向右动画”的回归。
2. **攻击减速为可调惯性**: `StateAttack` 的 `DecelerateSpeed` 驱动逐帧衰减，模拟“边走边出刀→慢慢收脚”。如需调整，优先改 Inspector 数值，不要直接把速度清零。
3. **特效镜像统一交给 Sprite**: `AttackEffectSprite` 的左右翻转由父级 `Sprite2D.Scale` 负责，状态脚本不重复设置 `FlipH`，避免和其他节点打架。
4. **注释面向 Go 同学**: 解释保持“类似服务端 XX 逻辑”的直白叙述，方便服务端同学阅读/维护；新增行为时写清“改了什么、为什么”。
5. **修改前后必做冒烟**: 每次重构玩家脚本，至少自测“八方向移动 + 左右攻击特效”两条路径，确认朝向、特效镜像、减速都正常再提交。

### Protocol Buffer集成

1. **协议生成**: 从`proto/csproto/`生成C#代码，使用`proto/genproto_csharp.bat`脚本
2. **版本管理**: 确保客户端和服务端使用相同版本的协议
3. **协议更新**: 协议更新后需要重新生成代码，并在`ProtocolHandler.InitializeProtocolMaps()`中注册新协议
4. **消息格式**: 
   - 服务端消息格式：`[4字节长度][1字节类型][1字节flags][消息体]`
   - 客户端消息格式：`[2字节MsgId][protobuf数据]`，包装在通用Message中（包含SessionId）
   - 心跳消息类型：0x06 (MsgTypeHeartbeat)，内容为"ping"

### 代码风格（客户端 C#）

1. **优先使用早返回**: 条件不满足时尽早 `return`，真正的更新逻辑写在函数下半部分，避免多层嵌套  
   - 例如 `UpdateAnimation()` 中：`if (!needPlay) return; _animationPlayer.Play(animName);`
2. **减少简单的 if/else**: 对于只是根据布尔值选择一个结果的情况，优先使用三元运算符  
   - 例如 `state = isMoving ? "move" : "idle";`
3. **拆分长条件表达式**: 复杂条件先提取为中间布尔变量，再逐步判断，避免一行里堆太多逻辑  
   - 例如 `bool hasInput`、`bool isHorizontal`、`bool needPlay` 等
4. **避免在没有 else 的 if 中直接写更新逻辑**: 若某个条件不满足就不需要更新时，优先使用“条件不满足直接 return”的写法  
   - 便于阅读和后续插入新逻辑
5. **命名清晰表达意图**: 中间变量和状态名（如 `state`、`directionChanged`、`stateChanged`）要能直接看出用途，减少额外注释负担

---

## 📖 客户端脚本开发流程

### 1. Protocol Buffer协议生成流程

#### 步骤1: 生成协议代码
```bash
# Windows
cd proto
.\genproto_csharp.bat

# 生成后的文件位于: client/Scripts/Protocol/
```

#### 步骤2: 在ProtocolHandler中注册协议映射
**位置**: `client/Scripts/Network/ProtocolHandler.cs` 的 `InitializeProtocolMaps()` 方法

**C2S协议注册**（客户端发送到服务端）:
```csharp
RegisterC2SProtocol((int)C2SProtocol.C2Sregister, typeof(C2SRegisterReq));
RegisterC2SProtocol((int)C2SProtocol.C2Slogin, typeof(C2SLoginReq));
// ... 其他C2S协议
```

**S2C协议注册**（服务端发送到客户端）:
```csharp
RegisterS2CProtocol((int)S2CProtocol.S2CregisterResult, typeof(S2CRegisterResultReq));
RegisterS2CProtocol((int)S2CProtocol.S2CloginResult, typeof(S2CLoginResultReq));
// ... 其他S2C协议
```

**注意**: 
- 协议ID使用枚举 `C2SProtocol` 和 `S2CProtocol`（定义在生成的 `Cs.cs` 和 `Sc.cs` 中）
- 消息类型使用生成的protobuf类（如 `C2SRegisterReq`、`S2CRegisterResultReq`）

#### 步骤3: 在UI脚本中注册协议处理器
**位置**: UI脚本的 `_Ready()` 方法中

**示例**（LoginUI.cs）:
```csharp
public override void _Ready()
{
    // 注册协议处理器
    MessageReceiver.Instance?.RegisterHandler((int)S2CProtocol.S2CregisterResult, OnRegisterResult);
    MessageReceiver.Instance?.RegisterHandler((int)S2CProtocol.S2CloginResult, OnLoginResult);
    MessageReceiver.Instance?.RegisterHandler((int)S2CProtocol.S2Cerror, OnError);
}

// 处理协议消息
private void OnLoginResult(Google.Protobuf.IMessage message)
{
    if (message is S2CLoginResultReq result)
    {
        // 处理登录结果
    }
}
```

**重要**: 在 `_ExitTree()` 中取消注册，避免内存泄漏：
```csharp
public override void _ExitTree()
{
    MessageReceiver.Instance?.UnregisterHandler((int)S2CProtocol.S2CloginResult);
    base._ExitTree();
}
```

#### 步骤4: 发送协议消息
**位置**: UI脚本或游戏逻辑脚本中

**使用MessageSender发送消息**:
```csharp
// 创建请求消息
var req = new C2SLoginReq
{
    Username = username,
    Password = password
};

// 发送消息（协议ID使用枚举）
MessageSender.Send(req, (int)C2SProtocol.C2Slogin);
```

**完整流程示例**:
1. 用户点击登录按钮 → `OnLoginButtonPressed()`
2. 创建 `C2SLoginReq` 消息
3. 调用 `MessageSender.Send()` 发送
4. 服务端处理并返回 `S2CLoginResultReq`
5. `MessageReceiver` 自动分发到注册的处理器 `OnLoginResult()`
6. 在处理器中处理结果并更新UI

### 2. 场景切换流程

#### Login场景切换到RoleSelect场景

**位置**: `client/Scripts/UI/LoginUI.cs` 的 `SwitchToRoleSelect()` 方法

**实现**:
```csharp
private void SwitchToRoleSelect()
{
    var manager = SceneManager.Instance;
    if (manager != null)
    {
        await manager.SwitchSceneAsync("res://Scenes/RoleSelect.tscn", "正在进入角色选择...");
    }
    else
    {
        GetTree().ChangeSceneToFile("res://Scenes/RoleSelect.tscn");
    }
}
```

**调用时机**: 登录成功后，在 `OnLoginResult()` 中调用

**场景切换机制**:
1. 推荐通过 `SceneManager.SwitchSceneAsync()` 触发场景切换，内部使用 `ResourceLoader` 异步加载并显示Loading界面
2. 在 `SceneManager` 不可用的极端情况下，可回退到 Godot 的 `GetTree().ChangeSceneToFile()`：
   - 卸载当前场景（Login.tscn）
   - 加载新场景（RoleSelect.tscn）
   - 调用旧场景的 `_ExitTree()`（此时应取消注册协议处理器）
   - 调用新场景的 `_Ready()`（此时应注册协议处理器）

3. **AutoLoad场景保持**: `AutoLoad.tscn` 中的 `NetworkManager`、`MessageReceiver`、`SceneManager`、`LoadingScreen` 是自动加载的，不会因为场景切换而销毁

4. **协议处理器管理**:
   - 每个场景在 `_Ready()` 中注册自己需要的协议处理器
   - 在 `_ExitTree()` 中取消注册，避免内存泄漏和重复处理

### 3. 玩家状态机接入流程

1. **场景结构**  
   ```
   Player
     ├─ PlayerStateMachine (Node，脚本 `Player_state_machine.cs`)
     │    ├─ IdleState (Node，脚本 `StateIdle.cs`)
     │    └─ WalkState (Node，脚本 `StateWalk.cs`)
     ├─ AnimationPlayer
     └─ Sprite2D
   ```
   - 如需自定义路径，可在 `Player` 的 `StateMachinePath` 属性中指定状态机 NodePath

2. **状态扩展**  
   - 新状态继承 `State`，并实现所需的 `Enter/Exit/Process/Physics/HandleInput`
   - 状态切换通过返回其他 `State` 实例完成，返回 `null` 表示继续停留在当前状态
   - 动画播放统一调用 `Player.UpdateAnimation("idle" | "walk" | "attack" | …)`，由状态控制前缀
   - 如需监听 `AnimationPlayer` 信号，在 `Enter()` 中连接，在 `Exit()` 中断开
   - 状态节点位于 `StateMachine` 下，访问 `Player` 的直接子节点需通过 `GetParent().GetParent()` 向上查找

3. **调试方式**  
   - `Player` 在未找到状态机时会输出 `[Player] 未找到 Player_state_machine...`，状态机在 `_Process/_PhysicsProcess/_UnhandledInput` 无状态时也会有日志提醒

### 4. 客户端框架完善清单

#### ✅ 已完成的基础框架
- [x] 网络层（NetworkManager、ProtocolHandler、MessageSender、MessageReceiver）
- [x] 协议生成工具和协议映射注册机制
- [x] 登录流程（LoginUI、RoleSelectUI）
- [x] 本地数据存储（LocalStorage）
- [x] 日志工具（Logger）

#### ⚠️ 待完善的核心框架

**3.1 场景管理系统**（高优先级）
- [ ] `SceneManager.cs` - 场景管理器
  - 场景加载/卸载
  - 场景切换动画
  - 场景配置读取
  - 场景状态管理
- [ ] `Main.tscn` - 游戏主场景文件
  - 进入游戏后的主场景
  - 包含游戏世界节点和主界面UI节点
  - 参考 `client/Scenes/README.md`

**3.2 实体管理系统**（高优先级）
- [ ] `EntityManager.cs` - 实体管理器
- [ ] `Entity.cs` - 基础实体类
- [ ] `PlayerEntity.cs` - 玩家实体
- [ ] `MonsterEntity.cs` - 怪物实体
- [ ] `NPCEntity.cs` - NPC实体
- [ ] `DropItemEntity.cs` - 掉落物实体
- [ ] AOI视野管理

**3.3 数据层框架**（中优先级）
- [ ] `ConfigManager.cs` - 配置表管理器
  - JSON配置表加载
  - 配置表缓存
  - 配置表热更新
- [ ] `PlayerDataCache.cs` - 玩家数据缓存
  - 角色数据缓存
  - 背包数据缓存
  - 任务数据缓存
  - 数据同步机制

**3.4 工具层框架**（中优先级）
- [ ] `ResourceManager.cs` - 资源管理器
  - 资源预加载
  - 资源缓存
  - 资源释放
- [ ] `UtilityFunctions.cs` - 工具函数
  - 数学工具函数
  - 字符串工具函数
  - 时间工具函数

**3.5 时间同步模块**（高优先级）
- [ ] 在 `NetworkManager.cs` 中实现 Ping/Pong 时间同步
  - 定期向Gateway请求服务器时间
  - 计算本地与服务器时间差（偏移量）
  - 提供统一的"服务器时间"访问接口
- [ ] Debug HUD显示时间差（可选）

**3.6 游戏逻辑框架**（高优先级）
- [ ] `PlayerController.cs` - 玩家控制器
- [ ] `MovementSystem.cs` - 移动系统（集成服务器时间同步）
- [ ] `CombatSystem.cs` - 战斗系统（集成服务器时间同步和帧定义）
- [ ] `SkillSystem.cs` - 技能系统

### 4. 开发工作流建议

#### 开发新功能的标准流程

1. **协议定义**（如需要）
   - 在 `proto/csproto/` 中定义或修改 `.proto` 文件
   - 运行 `proto/genproto_csharp.bat` 生成C#代码

2. **协议注册**
   - 在 `ProtocolHandler.InitializeProtocolMaps()` 中注册新协议

3. **UI开发**（如需要）
   - 在Godot编辑器中创建场景文件（`.tscn`）
   - 创建对应的C#脚本（如 `XXXUI.cs`）
   - 在脚本的 `_Ready()` 中：
     - 获取UI节点引用
     - 连接信号
     - 注册协议处理器
   - 在脚本的 `_ExitTree()` 中：
     - 取消注册协议处理器

4. **业务逻辑实现**
   - 实现发送消息逻辑（使用 `MessageSender.Send()`）
   - 实现接收消息处理（在注册的处理器中）
   - 实现UI更新逻辑

5. **测试和调试**
   - 测试协议收发
   - 测试UI交互
   - 检查内存泄漏（确保协议处理器正确取消注册）

6. **文档更新**
   - 更新 `docs/godot客户端开发进度文档.md`
   - 记录实现细节和关键代码位置

---

## 💡 开发建议

### 开发顺序建议

1. **Phase 1**: 先完成基础框架（项目初始化、协议生成、网络层）
2. **Phase 2**: 再实现登录流程（登录界面、角色选择界面）
3. **Phase 3**: 然后实现游戏场景系统（场景管理、实体管理）
4. **Phase 4**: 接着实现游戏逻辑系统（移动、战斗）
5. **Phase 5**: 最后实现UI系统（各个功能界面）

### 开发注意事项

- 每次完成新功能后，更新文档的"已完成功能"部分
- 遇到重要架构决策时，在"开发注意事项"中记录
- 发现新的关键代码位置时，补充到"关键代码位置"部分
- 代码中的TODO注释要及时处理，不要积累
- 参考服务端的实现方式，保持协议处理的一致性

---

## 🔗 相关文档

- `docs/godot客户端规划.md` - Godot客户端规划文档
- `docs/服务端开发进度文档.md` - 服务端开发进度文档
- `docs/服务器规划.md` - 服务器规划文档
- `docs/游戏流程优化建议.md` - 游戏流程优化建议

---

## 💡 提示

下次继续开发时，只需要查看这份文档，开发者就能快速了解：
- 项目当前状态
- 已完成功能
- 待实现功能
- 关键代码位置
- 开发注意事项

**建议**: 每次完成新功能后，更新此文档的"已完成功能"部分。
