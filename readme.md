## simple-file-sync

简单文件同步工具

### 简介

一个简单的文件同步工具，用于从服务器到客户端的文件同步。它只处理文件的添加和修改事件。

### 使用方法

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
