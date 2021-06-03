# 搜索协议

## 传输协议

| 名称 | 字节数 | 描述 |
| ---- | ---- | ---- |
| 帧头 | 2 | 0xAAAA
| 包总数L | 1 | |
| 包总数H | 1 | |
| 当前包号L | 1 | 当包总数大于1时，才有该字段 |
| 当前包号H | 1 | 当包总数大于1时，才有该字段 |
| 包序号L | 1 | 当包总数大于1时， 才有该字段， 用于表示包为一帧数据 |
| 包序号H | 1 | 当包总数大于1时， 才有该字段， 用于表示包为一帧数据 |
| 数据长度 | 1 | |
| 数据 | 数据长度-1 | |

## 应用协议

### 搜索

#### 请求

```json5
{
  "command": "search",
  "session": "int"
}
```

#### 回复

```json5
{
  "command": "search",
  "session": "int",
  "result": "bool",
  "error": "string",
  "data": {
    "version": {
      "version": "string",
      "commitId": "string",
      "buildTime": "string"
    },
    "license": {
      "fingerprint": "string",
      "status": "string",
      "isForever": "bool",
      "ValidDatetime": "string",
      "ValidDuration": "int"
    }
  }
}
```

### 获取网卡信息列表

#### 请求

```json5
{
  "command": "net_list",
  "session": "int",
  "username": "string",
  "password": "string"
}
```

#### 回复

```json5
{
  "command": "net_list",
  "session": "int",
  "result": "bool",
  "error": "string",
  "data": [{
    "index": "int",
    "name": "string",
    "description": "string",
    "mac": "string",
    "ip": "string",
    "mask": "string",
    "gateway": "string"
  }]
}
```

### 设置网卡信息

#### 请求

```json5
{
  "command": "set_net",
  "session": "int",
  "username": "string",
  "password": "string",
  "data": {
    "name": "string",
    "ip": "string",
    "mask": "string",
    "gateway": "string",
  }
}
```

#### 回复

```json5
{
  "command": "set_net",
  "session": "int",
  "result": "bool",
  "error": "string"
}
```

### 升级

#### 请求

```json5
{
  "command": "update",
  "session": "int",
  "data": "string", //升级文件base64
}
```

#### 回复

```json5
{
  "command": "update",
  "session": "int",
  "result": "bool",
  "error": "string"
}
```

### 升级通知

#### 请求

```json5
{
  "command": "update_notify",
  "session": "int",
  "data": {
    "status": "string",
    "percent": "int",
  }
}
```

#### 回复

```json5
{
  "command": "update_notify",
  "session": "int",
  "result": "bool",
  "error": "string"
}
```
