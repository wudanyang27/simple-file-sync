# Simple File Sync

一个简单的文件同步工具，支持实时监控本地文件变化并同步到远程服务器。

## 功能特点

- 实时监控文件变化
- 支持全量同步或仅同步git差异文件
- 支持文件忽略模式（正则表达式）
- 支持路径映射（正则表达式）
- 支持TOML配置文件
- 支持多远程目标配置

## 使用方法

### 命令行参数

```bash
simple-file-sync client --local-dir=/path/to/local --mode=all --remote-dir=/path/to/remote --server-addr=http://server/receiver --server-token=token
```

#### 主要参数

- `--mode`: 同步模式，支持 `all`（所有文件）或 `git`（仅git差异文件）
- `--local-dir`: 本地目录
- `--remote-dir`: 远程目录
- `--server-addr`: 服务器地址
- `--server-token`: 服务器验证令牌
- `--target`: 指定要使用的远程目标名称

#### 忽略和映射

- `--ignore`: 忽略模式，使用正则表达式，多个模式用逗号分隔。程序会自动添加 `^` 和 `$` 以实现全值匹配
  ```bash
  --ignore=".*/\\.tmp$,.*/__pycache__"
  ```

- `--mapping`: 路径映射，格式为 `源路径正则:目标路径模板`，多个映射用逗号分隔
  ```bash
  --mapping="/src(/.*)\\.js$:/build$1.min.js,/images(/.*):static/images$1"
  ```

### 配置文件

你可以使用TOML格式的配置文件来存储设置：

```bash
simple-file-sync client -c /path/to/config.toml
# 或
simple-file-sync client --config=/path/to/config.toml
```

如果不指定配置文件路径，程序会尝试读取当前目录下的 `simple-file-sync.toml`。

#### 示例配置文件

```toml
# 客户端模式：all - 同步所有文件，git - 仅同步git差异文件
mode = "all"

# 本地目录，将被监控和同步
local_dir = "/Users/username/projects/my-project"

# 当前激活的远程目标名称
active_target = "dev"

# 忽略模式 (全值匹配，正则表达式)
# 注意：程序会自动添加 ^ 和 $ 作为匹配边界，无需手动添加
# 忽略模式 (全值匹配，正则表达式)
ignore = [
    ".*/\\.tmp$",
    ".*/\\.log$",
    ".*/node_modules$",
    ".*/build$",
    ".*/.git/.*",       # 匹配所有.git目录内的任何文件/目录
    ".*/\\.git",        # 匹配.git目录（不带末尾斜杠）
    ".*/\\.DS_Store$",
    ".*/__pycache__",
    ".*/.gitignore",
    ".*/.idea",
    ".*/.vscode",
    ".*/.github",
]

# 路径映射 (正则表达式:目标路径格式)
path_mappings = [
  "/src(/.*)\\.js$:/build$1.min.js",
  "/images(/.*):static/images$1"
]

# 远程目标配置
[[remote_targets]]
name = "dev"
server_addr = "http://127.0.0.1:8120/receiver"
remote_dir = "/path/to/dev"
token = "dev-token"

[[remote_targets]]
name = "production"
server_addr = "http://example.com:8120/receiver"
remote_dir = "/path/to/production"
token = "production-token"
```

更多示例请参考 `examples/simple-file-sync.toml`。

### 服务器部分

启动文件同步接收服务器：

```bash
simple-file-sync server --port=8120 --token=your-secret-token --limit-dir=/path/to/allowed/dir
```

#### 服务器参数

- `--port`: 监听端口（默认 8120）
- `--token`: 验证令牌（默认 kfcvme50）
- `--limit-dir`: 限制上传文件的目录（默认为用户主目录）

## 优先级

命令行参数的优先级高于配置文件中的设置。如果同时指定了配置文件和命令行参数，命令行参数将覆盖配置文件中的对应设置。
