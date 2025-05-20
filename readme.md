# Simple File Sync

一个简单的文件同步工具，支持实时监控本地文件变化并同步到远程服务器。

## 功能特点

- 实时监控文件变化
- 支持全量同步或仅同步git差异文件
- 支持文件忽略模式（正则表达式）
- 支持路径映射（正则表达式）
- 支持TOML配置文件

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

#### 忽略和映射

- `--ignore`: 忽略模式，使用正则表达式，多个模式用逗号分隔
  ```bash
  --ignore="\\.tmp$,node_modules,build"
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

# 远程目录，文件将被上传到这里
remote_dir = "/remote/path/my-project"

# 服务器地址
server_addr = "http://example.com:8120/receiver"

# 服务器验证令牌
server_token = "your-secret-token"

# 忽略模式 (正则表达式)
ignore = [
  "\\.tmp$",
  "\\.log$",
  "node_modules",
  "build"
]

# 路径映射 (正则表达式:目标路径格式)
path_mappings = [
  "/src(/.*)\\.js$:/build$1.min.js",
  "/images(/.*):static/images$1"
]
```

更多示例请参考 `examples/simple-file-sync.toml`。

## 优先级

命令行参数的优先级高于配置文件中的设置。如果同时指定了配置文件和命令行参数，命令行参数将覆盖配置文件中的对应设置。
