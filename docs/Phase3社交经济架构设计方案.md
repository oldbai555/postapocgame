# Phase 3 社交经济架构设计方案

## 概述

Phase 3 社交经济包含四个核心功能：
1. **聊天系统**（世界/私聊）
2. **好友/公会框架**
3. **拍卖行**
4. **排行榜**

所有功能均采用 **PublicActor + PlayerActor 协作** 的架构模式，完全符合 Actor 模型和无锁设计原则。

---

## 架构原则

### PublicActor 职责
- **全局数据管理**：排行榜、公会、拍卖行等全局数据
- **跨玩家消息路由**：聊天、好友消息转发等
- **在线状态管理**：维护玩家在线状态映射（roleId → SessionId）
- **快照数据缓存**：角色展示数据快照（用于排行榜、好友列表等）

### PlayerActor 职责
- **玩家数据管理**：好友关系、玩家公会信息（所属公会ID、职位等）存储在 `PlayerRoleBinaryData` 中
- **数据变更通知**：属性变化、好友关系变化时主动通知 PublicActor
- **消息接收与发送**：接收客户端请求，转发到 PublicActor，接收回调后发送给客户端

### 数据存储原则

**重要说明**：
- **公会成员列表**：存储在 PublicActor 的公会数据中（权威数据源）
- **玩家公会信息**：存储在 `PlayerRoleBinaryData.GuildData` 中（玩家所属公会ID、职位等）
- **原因**：
  - 公会数据是全局的，需要统一管理，查询成员列表时直接从 PublicActor 获取
  - 玩家数据是分散的，每个玩家只存储自己的公会归属信息
  - 避免数据冗余和不一致：公会成员列表是公会的属性，应该由公会来维护

### 协作模式
```
客户端请求
    ↓
PlayerActor（接收请求，验证权限）
    ↓
发送消息到 PublicActor
    ↓
PublicActor（处理全局逻辑，可能需要查询其他 PlayerActor）
    ↓
回调消息到 PlayerActor
    ↓
PlayerActor（发送响应给客户端）
```

---

## 1. 聊天系统

### 功能需求
- **世界聊天**：所有在线玩家可见
- **私聊**：点对点聊天
- **聊天频道**：可扩展支持队伍、公会等频道

### 架构设计

#### 1.1 数据结构

**Proto 定义位置**：需要在 `proto/csproto/` 中定义以下结构：

**在 `proto/csproto/social_def.proto` 中定义**：
```protobuf
// 聊天消息类型枚举
enum ChatType {
    ChatTypeNil = 0;
    ChatTypeWorld = 1;   // 世界聊天
    ChatTypePrivate = 2; // 私聊
    ChatTypeGuild = 3;   // 公会聊天
    ChatTypeTeam = 4;    // 队伍聊天
}

// 聊天消息
message ChatMessage {
    ChatType chat_type = 1;
    uint64 sender_id = 2;
    string sender_name = 3;
    uint64 target_id = 4;  // 私聊时使用
    string content = 5;
    int64 timestamp = 6;
}
```

**注意**：`OnlinePlayerMap` 是 PublicActor 内部实现细节，不需要在 Proto 中定义，可在 `server/internal/argsdef/` 中定义（如果需要）。

#### 1.2 实现方案

**✅ 适合在 PublicActor 中实现**

**理由**：
- 世界聊天需要广播给所有在线玩家，PublicActor 维护在线玩家列表
- 私聊需要根据 roleId 查找目标玩家的 SessionId，PublicActor 维护映射关系
- 完全无锁：所有操作都在 Actor 串行上下文中执行

**流程设计**：

1. **世界聊天流程**：
   ```
   客户端发送世界聊天消息
        ↓
   PlayerActor 接收（验证频率限制、内容过滤）
        ↓
   发送 ChatWorldMsg 到 PublicActor
        ↓
   PublicActor：
     - 验证发送者在线状态
     - 广播消息给所有在线玩家（通过 SessionId 列表）
     - 记录聊天日志（可选）
        ↓
   PublicActor 向所有在线 PlayerActor 发送 ChatBroadcastMsg
        ↓
   各 PlayerActor 收到后发送给对应客户端
   ```

2. **私聊流程**：
   ```
   客户端发送私聊消息
        ↓
   PlayerActor 接收（验证目标是否存在）
        ↓
   发送 ChatPrivateMsg 到 PublicActor
        ↓
   PublicActor：
     - 查找目标玩家的 SessionId
     - 如果目标在线，转发消息到目标 PlayerActor
     - 如果目标离线，存储离线消息（可选）
        ↓
   PublicActor 发送 ChatPrivateRespMsg 给发送者 PlayerActor（确认发送成功）
   PublicActor 发送 ChatPrivateMsg 给目标 PlayerActor
        ↓
   两个 PlayerActor 分别发送消息给对应客户端
   ```

#### 1.3 关键实现点

- **频率限制**：在 PlayerActor 中实现，防止刷屏
- **内容过滤**：在 PlayerActor 中实现，敏感词过滤
- **在线状态同步**：角色进入游戏时注册到 PublicActor，离线时清理

---

## 2. 好友/公会框架

### 功能需求

#### 2.1 好友系统
- 添加好友、删除好友
- 好友列表查询
- 好友在线状态
- 好友申请/同意/拒绝

#### 2.2 公会系统
- 创建公会、解散公会
- 加入公会、退出公会
- 公会成员管理（权限、职位）
- 公会信息查询

### 架构设计

#### 2.1 好友系统

**✅ 适合在 PublicActor + PlayerActor 协作实现**

**数据存储**：
- 好友关系数据存储在 `PlayerRoleBinaryData.FriendData` 中（每个玩家的好友列表）
  - 需要在 `proto/csproto/system.proto` 中定义 `SiFriendData` 结构
  - 需要在 `PlayerRoleBinaryData` 中添加 `friend_data` 字段
  - 需要创建 `FriendSys`（EntitySystem），在 `OnInit` 中从 `BinaryData.FriendData` 初始化
- PublicActor 维护在线好友状态映射（roleId → isOnline）

**流程设计**：

1. **添加好友流程**：
   ```
   客户端发送添加好友请求
        ↓
   PlayerActor 接收（验证是否已是好友、是否达到上限）
        ↓
   发送 AddFriendReqMsg 到 PublicActor
        ↓
   PublicActor：
     - 验证目标玩家是否存在
     - 检查目标玩家是否在线
     - 如果在线，发送 FriendRequestMsg 到目标 PlayerActor
     - 如果离线，存储好友申请（可选）
        ↓
   目标 PlayerActor 收到后：
     - 更新 BinaryData 中的好友申请列表
     - 发送通知给客户端
        ↓
   目标玩家同意/拒绝后：
     - 发送 FriendResponseMsg 到 PublicActor
     - PublicActor 转发给申请者 PlayerActor
     - 双方更新 BinaryData 中的好友列表
   ```

2. **好友列表查询流程**：
   ```
   客户端请求好友列表
        ↓
   PlayerActor 从 BinaryData 读取好友 ID 列表
        ↓
   发送 QueryFriendListMsg 到 PublicActor（携带好友 ID 列表）
        ↓
   PublicActor：
     - 从快照缓存获取好友的展示数据
     - 查询好友在线状态
     - 组装好友列表数据
        ↓
   发送 FriendListRespMsg 回 PlayerActor
        ↓
   PlayerActor 发送给客户端
   ```

#### 2.2 公会系统

**✅ 完全适合在 PublicActor 中实现**

**数据存储**：
- **公会核心数据存储在 PublicActor 中**：
  - 公会信息（GuildId、名称、创建时间、等级等）
  - **公会成员列表**（`[]GuildMember`，包含 roleId、职位、权限等）
  - 公会设置、公告等
  - **这是权威数据源**：查询公会成员时直接从 PublicActor 获取
  - **持久化**：需要将 PublicActor 中的公会数据持久化到数据库（可创建独立的 Guild 表，或使用 JSON 存储）
- **玩家公会信息存储在 `PlayerRoleBinaryData.GuildData` 中**：
  - 所属公会ID（`GuildId`）
  - 在公会中的职位（`Position`）
  - 加入公会时间等
  - **这是玩家自己的数据**：查询玩家所属公会时从 BinaryData 获取
  - 需要在 `proto/csproto/system.proto` 中定义 `SiGuildData` 结构
  - 需要在 `PlayerRoleBinaryData` 中添加 `guild_data` 字段
  - 需要创建 `GuildSys`（EntitySystem，仅处理玩家公会信息），在 `OnInit` 中从 `BinaryData.GuildData` 初始化

**为什么这样设计？**
- **数据所有权清晰**：公会成员列表是公会的属性，应该由公会（PublicActor）来维护
- **避免数据冗余**：如果每个玩家都存储完整的成员列表，会导致数据冗余和不一致
- **查询效率高**：查询公会成员时直接从 PublicActor 获取，无需遍历所有玩家
- **数据一致性**：成员列表的增删改都在 PublicActor 中统一处理，保证数据一致性

**流程设计**：

1. **创建公会流程**：
   ```
   客户端发送创建公会请求
        ↓
   PlayerActor 接收（验证是否已有公会、是否满足条件）
        ↓
   发送 CreateGuildMsg 到 PublicActor
        ↓
   PublicActor：
     - 验证公会名是否重复
     - 创建公会数据（GuildId、名称、创建者等）
     - 将创建者加入公会成员列表
     - 更新创建者的 BinaryData（通过消息通知）
        ↓
   PublicActor 发送 CreateGuildRespMsg 回 PlayerActor
        ↓
   PlayerActor 更新 BinaryData，发送响应给客户端
   ```

2. **加入公会流程**：
   ```
   客户端发送加入公会申请
        ↓
   PlayerActor 接收
        ↓
   发送 JoinGuildReqMsg 到 PublicActor
        ↓
   PublicActor：
     - 验证公会是否存在
     - 验证公会是否满员
     - 发送 GuildJoinRequestMsg 到公会会长/管理员的 PlayerActor
        ↓
   管理员同意/拒绝后：
     - 发送 GuildJoinResponseMsg 到 PublicActor
     - PublicActor 更新公会成员列表
     - PublicActor 通知申请者 PlayerActor 更新 BinaryData
   ```

3. **公会信息查询流程**：
   ```
   客户端请求公会信息
        ↓
   PlayerActor 接收
        ↓
   发送 QueryGuildInfoMsg 到 PublicActor
        ↓
   PublicActor：
     - 从公会数据获取基本信息
     - 从快照缓存获取成员展示数据
     - 组装公会信息
        ↓
   发送 GuildInfoRespMsg 回 PlayerActor
        ↓
   PlayerActor 发送给客户端
   ```

#### 2.3 关键实现点

- **数据一致性**：公会数据在 PublicActor，玩家公会信息在 BinaryData，通过消息同步
- **权限管理**：公会职位、权限存储在 PublicActor 的公会数据中
- **成员列表**：PublicActor 维护公会成员列表，支持快速查询

---

## 3. 拍卖行

### 功能需求
- 上架物品（设置价格、数量、时长）
- 浏览拍卖行（分类、搜索、排序）
- 购买物品（一口价、竞拍）
- 下架物品（主动下架、过期下架）

### 架构设计

**✅ 完全适合在 PublicActor 中实现**

**数据存储**：
- 拍卖行数据存储在 PublicActor 中（物品列表、价格、卖家信息等）
  - **持久化**：需要将 PublicActor 中的拍卖行数据持久化到数据库（可创建独立的 Auction 表，或使用 JSON 存储）
- 玩家上架记录存储在 `PlayerRoleBinaryData.AuctionData` 中（上架物品ID、数量等）
  - 需要在 `proto/csproto/system.proto` 中定义 `SiAuctionData` 结构
  - 需要在 `PlayerRoleBinaryData` 中添加 `auction_data` 字段
  - 需要创建 `AuctionSys`（EntitySystem，仅处理玩家上架记录），在 `OnInit` 中从 `BinaryData.AuctionData` 初始化

**流程设计**：

1. **上架物品流程**：
   ```
   客户端发送上架请求
        ↓
   PlayerActor 接收（验证物品是否存在、数量是否足够）
        ↓
   从背包扣除物品（BagSys）
        ↓
   发送 AuctionPutOnMsg 到 PublicActor
        ↓
   PublicActor：
     - 生成拍卖行物品ID
     - 存储物品信息（物品ID、数量、价格、卖家ID、过期时间等）
     - 添加到拍卖行列表
        ↓
   PublicActor 发送 AuctionPutOnRespMsg 回 PlayerActor
        ↓
   PlayerActor 更新 BinaryData，发送响应给客户端
   ```

2. **购买物品流程**：
   ```
   客户端发送购买请求
        ↓
   PlayerActor 接收（验证货币是否足够）
        ↓
   发送 AuctionBuyMsg 到 PublicActor
        ↓
   PublicActor：
     - 验证物品是否还在拍卖行
     - 验证价格是否正确
     - 扣除买家货币（通过消息查询买家 PlayerActor）
     - 给卖家发放货币（通过消息通知卖家 PlayerActor）
     - 从拍卖行移除物品
        ↓
   PublicActor 发送 AuctionBuyRespMsg 回买家 PlayerActor
   PublicActor 发送 AuctionSoldMsg 到卖家 PlayerActor（如果在线）
        ↓
   买家 PlayerActor：添加物品到背包，发送响应给客户端
   卖家 PlayerActor：更新货币，发送通知给客户端
   ```

3. **浏览拍卖行流程**：
   ```
   客户端请求拍卖行列表
        ↓
   PlayerActor 接收
        ↓
   发送 QueryAuctionListMsg 到 PublicActor（携带筛选条件）
        ↓
   PublicActor：
     - 根据条件筛选物品
     - 从快照缓存获取卖家展示数据（可选）
     - 组装拍卖行列表
        ↓
   发送 AuctionListRespMsg 回 PlayerActor
        ↓
   PlayerActor 发送给客户端
   ```

#### 3.3 关键实现点

- **过期处理**：PublicActor 定期检查过期物品，自动下架并返还给卖家
- **货币交易**：通过消息机制在 PlayerActor 之间转账，保证原子性
- **物品交付**：购买成功后，通过消息将物品添加到买家背包

---

## 4. 排行榜

### 功能需求
- 多种排行榜类型（战力、等级、公会等）
- 查询前 N 名
- 查询自己的排名

### 架构设计

**✅ 完全适合在 PublicActor 中实现**

**核心原则**：排行榜只存储 key(int64) 和 value(int64)，通过 Actor 消息机制无锁获取展示数据。

**架构方案**：采用 **PublicActor + 快照数据 + 异步消息查询** 的混合方案（方案C）。

**数据存储**：
- 排行榜核心数据存储在 PublicActor 中（key-value 排序）
- 快照数据缓存存储在 PublicActor 中（角色展示数据快照）

**流程设计**：

1. **快照数据管理**：
   - `PublicActor` 维护快照数据缓存
   - 角色进入游戏时，`PlayerActor` 向 `PublicActor` 发送快照注册消息
   - 角色属性变化时（如战力变化），`PlayerActor` 主动发送快照更新消息
   - 角色离线时，`PublicActor` 清理对应的快照数据

2. **排行榜查询流程**：
   ```
   客户端请求排行榜
        ↓
   PlayerActor 接收请求
        ↓
   发送消息到 PublicActor（携带请求者 SessionId）
        ↓
   PublicActor 处理：
    1. 从排行榜数据获取前 N 名的 key 列表
    2. 从快照缓存中查找展示数据
    3. 如果快照不存在或过期，向对应 PlayerActor 发送异步查询消息
    4. 等待所有数据就绪后，组装排行榜数据
    5. 通过回调消息发送回请求的 PlayerActor
        ↓
   PlayerActor 收到回调后，发送排行榜数据给客户端
   ```

3. **排行榜更新机制**：
   - 当角色属性变化时（如战力提升），`PlayerActor` 发送更新消息到 `PublicActor`
   - `PublicActor` 更新排行榜的 value 值，并重新排序
   - 同时更新快照缓存

**优势**：
- ✅ **完全无锁**：所有操作都在各自 Actor 的串行上下文中执行
- ✅ **Actor 边界清晰**：通过消息传递，不直接引用 `PlayerRole`
- ✅ **性能优化**：只在数据变化时更新，避免定时轮询的开销
- ✅ **数据实时性**：属性变化时立即更新，保证数据新鲜度
- ✅ **扩展性强**：支持多种排行榜类型，通过 `RankType` 区分
- ✅ **容错性好**：快照不存在时可以异步查询，保证数据完整性

---

## PublicActor 统一设计

### 数据结构定义

**所有公共数据结构必须在 Proto 中定义**，位置如下：

#### 在 `proto/csproto/social_def.proto` 中定义：

```protobuf
// 排行榜类型枚举
enum RankType {
    RankTypeNil = 0;
    RankTypeRoleCombatPower = 1;  // 角色战力榜
    RankTypeRoleLevel = 2;        // 角色等级榜
    RankTypeGuildCombatPower = 3; // 工会战力榜
    // ... 更多排行榜类型
}

// 排行榜项
message RankItem {
    int64 key = 1;   // roleId 或 guildId
    int64 value = 2; // 战力、等级等
}

// 排行榜数据
message RankData {
    RankType rank_type = 1;
    repeated RankItem items = 2;  // 已排序的列表
    int64 updated_at = 3;         // 最后更新时间
}

// 角色排行榜快照
message PlayerRankSnapshot {
    uint64 role_id = 1;
    string role_name = 2;
    uint32 job = 3;
    uint32 sex = 4;
    uint32 level = 5;
    int64 combat_power = 6;  // 战力
    string avatar = 7;       // 头像
    map<uint32, uint32> fashion = 8; // 时装数据
    int64 updated_at = 9;     // 快照更新时间
}

// 公会职位枚举
enum GuildPosition {
    GuildPositionNil = 0;
    GuildPositionMember = 1;      // 成员
    GuildPositionViceLeader = 2;   // 副会长
    GuildPositionLeader = 3;       // 会长
}

// 公会成员信息
message GuildMember {
    uint64 role_id = 1;
    uint32 position = 2;  // 职位（使用 GuildPosition 枚举）
    int64 join_time = 3;  // 加入时间
    // ... 其他成员相关信息
}

// 公会数据（用于 PublicActor 内部，需要持久化）
message GuildData {
    uint64 guild_id = 1;
    string guild_name = 2;
    uint64 creator_id = 3;
    uint32 level = 4;
    int64 create_time = 5;
    repeated GuildMember members = 6;  // 成员列表（权威数据源）
    string announcement = 7;
    // ... 其他公会信息
}

// 拍卖行物品
message AuctionItem {
    uint64 auction_id = 1;    // 拍卖行物品ID
    uint32 item_id = 2;       // 物品ID
    uint32 count = 3;         // 数量
    int64 price = 4;          // 价格
    uint64 seller_id = 5;     // 卖家ID
    int64 expire_time = 6;    // 过期时间
    int64 create_time = 7;    // 上架时间
}
```

#### 在 `proto/csproto/system.proto` 中添加：

```protobuf
// 好友系统数据
message SiFriendData {
    repeated uint64 friend_list = 1;           // 好友ID列表
    repeated uint64 friend_request_list = 2;   // 好友申请列表
}

// 公会系统数据（玩家公会信息）
message SiGuildData {
    uint64 guild_id = 1;      // 所属公会ID
    uint32 position = 2;      // 职位（使用 GuildPosition 枚举）
    int64 join_time = 3;      // 加入公会时间
}

// 拍卖行系统数据（玩家上架记录）
message SiAuctionData {
    repeated uint64 auction_id_list = 1;  // 上架物品ID列表
}
```

#### 在 `proto/csproto/player.proto` 的 `PlayerRoleBinaryData` 中添加：

```protobuf
message PlayerRoleBinaryData {
    // ... 现有字段 ...
    SiFriendData friend_data = 17;      // 好友数据
    SiGuildData guild_data = 18;        // 公会数据（玩家部分）
    SiAuctionData auction_data = 19;     // 拍卖行数据（玩家部分）
}
```

#### 在 `proto/csproto/rpc.proto` 中添加 PublicActor 消息定义：

```protobuf
// PublicActor 内部消息（用于 Actor 间通信）
// 注意：这些消息仅用于服务端内部，不需要发送给客户端

// 在线状态注册
message RegisterOnlineMsg {
    uint64 role_id = 1;
    string session_id = 2;
}

// 在线状态注销
message UnregisterOnlineMsg {
    uint64 role_id = 1;
}

// 聊天世界消息
message ChatWorldMsg {
    uint64 sender_id = 1;
    string sender_name = 2;
    string content = 3;
}

// 聊天私聊消息
message ChatPrivateMsg {
    uint64 sender_id = 1;
    uint64 target_id = 2;
    string content = 3;
}

// 好友申请消息
message AddFriendReqMsg {
    uint64 requester_id = 1;
    uint64 target_id = 2;
}

// 创建公会消息
message CreateGuildMsg {
    uint64 creator_id = 1;
    string guild_name = 2;
}

// 拍卖行上架消息
message AuctionPutOnMsg {
    uint64 seller_id = 1;
    uint32 item_id = 2;
    uint32 count = 3;
    int64 price = 4;
    int64 duration = 5;
}

// 更新排行榜快照消息
message UpdateRankSnapshotMsg {
    uint64 role_id = 1;
    PlayerRankSnapshot snapshot = 2;
}

// 更新排行榜值消息
message UpdateRankValueMsg {
    RankType rank_type = 1;
    int64 key = 2;
    int64 value = 3;
}

// 查询排行榜请求消息
message QueryRankReqMsg {
    RankType rank_type = 1;
    int32 top_n = 2;
    string requester_session_id = 3;  // 请求者的 SessionId
}
```

**注意**：`PublicActor` 结构体本身是 Go 代码实现细节，不需要在 Proto 中定义。上述 Proto 定义的是 PublicActor 使用的数据结构和消息格式。

---

## 总结

### 所有功能都适合在 PublicActor 中实现

**理由**：
1. ✅ **完全无锁**：所有操作都在 Actor 串行上下文中执行
2. ✅ **数据隔离**：不直接引用 PlayerRole，通过消息传递数据
3. ✅ **性能优化**：快照缓存减少查询频率
4. ✅ **扩展性强**：易于添加新功能
5. ✅ **符合架构**：遵循 Actor 模型和无锁设计原则

### 实现注意事项

#### EntitySystem 实现模式

所有玩家数据相关的系统都需要实现 EntitySystem 模式：

1. **创建系统结构**：
   ```go
   type FriendSys struct {
       *BaseSystem
       friendData *protocol.SiFriendData  // 使用 protocol 包中的 Proto 定义
   }
   ```

2. **实现 OnInit**：
   ```go
   func (fs *FriendSys) OnInit(ctx context.Context) {
       playerRole, _ := GetIPlayerRoleByContext(ctx)
       binaryData := playerRole.GetBinaryData()
       
       if binaryData.FriendData == nil {
           binaryData.FriendData = &protocol.SiFriendData{
               FriendList: make([]uint64, 0),
           }
       }
       fs.friendData = binaryData.FriendData  // 直接引用 BinaryData 中的数据（无锁）
   }
   ```

3. **注册系统**：
   ```go
   // 在系统初始化时注册
   entitysystem.RegisterSystemFactory(
       uint32(protocol.SystemId_SysFriend),
       func() iface.ISystem { return NewFriendSys() },
   )
   ```

4. **数据访问**：
   - 直接引用 `BinaryData` 中的数据结构（无锁）
   - 所有修改直接作用于 `BinaryData`，自动持久化

#### PublicActor 持久化策略

PublicActor 中的全局数据需要持久化：

1. **公会数据持久化**：
   - 方案A：创建独立的 `Guild` 表，使用 GORM 管理
   - 方案B：将公会数据序列化为 JSON，存储在数据库的 `guild_data` 表中
   - 建议：使用方案A，便于查询和维护

2. **拍卖行数据持久化**：
   - 方案A：创建独立的 `AuctionItem` 表，使用 GORM 管理
   - 方案B：将拍卖行数据序列化为 JSON，存储在数据库的 `auction_data` 表中
   - 建议：使用方案A，便于查询和过期清理

3. **持久化时机**：
   - 数据变更时立即持久化（同步写入）
   - 或定期批量持久化（异步写入，需要处理宕机恢复）

4. **数据恢复**：
   - GameServer 启动时，从数据库加载所有公会数据和拍卖行数据到 PublicActor
   - 需要处理数据一致性（如公会成员已离线等）

### 实现优先级建议

1. **Phase 1**：PublicActor 框架 + 在线状态管理 + 持久化框架
2. **Phase 2**：聊天系统（世界聊天 + 私聊）
3. **Phase 3**：好友系统（FriendSys EntitySystem + PublicActor 在线状态）
4. **Phase 4**：排行榜系统
5. **Phase 5**：公会系统（GuildSys EntitySystem + PublicActor 公会管理 + 持久化）
6. **Phase 6**：拍卖行（AuctionSys EntitySystem + PublicActor 拍卖行管理 + 持久化）

### 关键代码位置

**PublicActor 相关**：
- `server/service/gameserver/internel/publicactor`：PublicActor 实现
- `server/service/gameserver/internel/publicactor/chat_sys.go`：聊天系统
- `server/service/gameserver/internel/publicactor/friend_sys.go`：好友系统（PublicActor 部分）
- `server/service/gameserver/internel/publicactor/guild_sys.go`：公会系统（PublicActor 部分）
- `server/service/gameserver/internel/publicactor/auction_sys.go`：拍卖行系统（PublicActor 部分）
- `server/service/gameserver/internel/publicactor/rank_sys.go`：排行榜系统

**PlayerActor EntitySystem 相关**：
- `server/service/gameserver/internel/playeractor/entitysystem/friend_sys.go`：好友系统（PlayerActor 部分，EntitySystem）
- `server/service/gameserver/internel/playeractor/entitysystem/guild_sys.go`：公会系统（PlayerActor 部分，EntitySystem，仅处理玩家公会信息）
- `server/service/gameserver/internel/playeractor/entitysystem/auction_sys.go`：拍卖行系统（PlayerActor 部分，EntitySystem，仅处理玩家上架记录）

**Proto 定义**：
- `proto/csproto/social_def.proto`：需要创建此文件，定义社交经济相关的枚举和数据结构（`ChatType`、`ChatMessage`、`RankType`、`RankItem`、`RankData`、`PlayerRankSnapshot`、`GuildPosition`、`GuildMember`、`GuildData`、`AuctionItem` 等）
- `proto/csproto/system.proto`：需要添加 `SiFriendData`、`SiGuildData`、`SiAuctionData` 定义
- `proto/csproto/player.proto`：需要在 `PlayerRoleBinaryData` 中添加对应字段（`friend_data`、`guild_data`、`auction_data`）
- `proto/csproto/rpc.proto`：需要添加 PublicActor 内部消息定义（`RegisterOnlineMsg`、`ChatWorldMsg`、`AddFriendReqMsg` 等）

**数据库持久化**：
- `server/internal/database/guild.go`：公会数据持久化（可选，如果使用数据库存储）
- `server/internal/database/auction.go`：拍卖行数据持久化（可选，如果使用数据库存储）

