## simple-file-sync

简单文件同步工具

### 简介

一个简单的文件同步工具，用于从服务器到客户端的文件同步。它只处理文件的添加和修改事件。

### 使用方法

#### 配置文件 (推荐)

最简单的使用方式是创建配置文件。在项目目录中创建 `simple-file-sync.toml`：

```toml
mode = "all"
remote_dir = "/home/users/backup"
server_addr = "http://server:8120/receiver"
server_token = "your-token"
```

然后运行：
```bash
simple-file-sync client
```

本地目录会自动设置为配置文件所在的目录。详细配置说明请参考 [配置文件指南](docs/configuration.md)。

#### 命令行模式

#### 服务器端

启动服务器以接收文件上传：

```bash
simple-file-sync server --port=8080 --limit-dir=/Users/wudanyang/xxx --token=123456
```

#### 客户端

启动客户端以上传文件：

```bash
simple-file-sync client --local-dir=/Users/wudanyang/xxx --remote-dir=/home/users/wudanyang/wudanyang/xxx --server-addr=http://xxx:xxx/receiver --server-token=123456
```
