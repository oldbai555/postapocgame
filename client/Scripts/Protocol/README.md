# Protocol Buffer C# 代码生成说明

## 前置要求

1. **protoc 工具**：已包含在 `proto/` 目录下（`protoc.exe`）
2. **Google.Protobuf NuGet包**：已在 `client.csproj` 中添加

## 生成步骤

### Windows 方式

运行批处理脚本：
```bash
cd proto
.\genproto_csharp.bat
```

### 手动方式

```bash
cd proto
.\protoc.exe -I=csproto --csharp_out=..\client\Scripts\Protocol csproto\*.proto
```

## 生成后的文件

生成的C#协议代码将位于 `client/Scripts/Protocol/` 目录下，每个 `.proto` 文件会生成对应的 `.cs` 文件。

## 注意事项

1. 生成后需要在 `ProtocolHandler.cs` 中注册协议映射
2. 确保生成的代码命名空间与项目一致
3. 如果修改了 `.proto` 文件，需要重新生成并更新协议映射

## 协议映射注册示例

在 `ProtocolHandler.InitializeProtocolMaps()` 中注册：

```csharp
// C2S协议
RegisterC2SProtocol((int)C2SProtocol.C2SRegister, typeof(C2SRegisterReq));
RegisterC2SProtocol((int)C2SProtocol.C2SLogin, typeof(C2SLoginReq));

// S2C协议
RegisterS2CProtocol((int)S2CProtocol.S2CRegisterResult, typeof(S2CRegisterResult));
RegisterS2CProtocol((int)S2CProtocol.S2CLoginResult, typeof(S2CLoginResult));
```

