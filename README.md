# go-bin

基于 Go 的轻量文件分享网页服务，支持分享文件、文本和链接。

## 安装

**Linux / macOS:**

```bash
curl -fsSL https://raw.githubusercontent.com/zaaack/go-bin/main/install.sh | bash
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/zaaack/go-bin/main/install.ps1 | iex
```

也可以从 [Releases](https://github.com/zaaack/go-bin/releases) 页面手动下载。

## 功能

- 公开分享会出现在列表页
- 私有分享使用随机 URL，仅凭链接访问
- 文件尽量保留原文件名展示
- 文本和链接在列表页展示前 2 行摘要
- 列表页支持下载文件、复制文本、复制 URL、打开 URL
- 详情页支持下载、复制下载链接、复制文本、复制 URL、打开 URL
- 支持置顶
- 支持过期时间和永不过期
- SQLite 存元数据，`uploads/` 存文件

## 启动

```powershell
$env:GO111MODULE = "on"
$env:GOPROXY = "https://goproxy.cn,direct"
go run ./cmd/go-bin serve
```

也可以先生成可执行文件：

```powershell
$env:GO111MODULE = "on"
$env:GOPROXY = "https://goproxy.cn,direct"
go generate .
```

## 参数

```powershell
go run ./cmd/go-bin serve \
  --addr :8080 \
  --db data.db \
  --uploads-dir uploads \
  --base-url http://localhost:8080 \
  --default-public=true \
  --default-pin=false \
  --default-expire=3mo
```

`--db` 支持指定 sqlite 文件位置。

`--default-expire` 支持：

- `never`
- `1d`
- `7d`
- `30d`
- `1mo`
- `3mo`
- `1y`

## 页面

- `/` 公开列表页
- `/new` 发布页
- `/s/{slug}` 详情页
- `/download/{slug}` 文件下载
