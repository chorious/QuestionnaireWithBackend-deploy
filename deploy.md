# 云服务器部署指南

> 本仓库是 [QuestionnaireWithBackend](https://github.com/chorious/QuestionnaireWithBackend) 的云服务器部署版。
> 与原版的主要区别：Go 后端通过 `embed.FS` 嵌入前端 dist，单二进制部署；同时提供 Docker / systemd / nginx 多种部署方式。

---

## 快速开始（Docker Compose）

### 1. 克隆仓库

```bash
git clone https://github.com/chorious/QuestionnaireWithBackend-deploy.git
cd QuestionnaireWithBackend-deploy
```

### 2. 设置环境变量

```bash
cp .env.example .env
# 编辑 .env，设置 ADMIN_TOKEN
```

### 3. 一键启动

```bash
docker-compose up -d --build
```

服务运行在 `http://0.0.0.0:3000`，数据持久化在 `./data/`。

---

## 手动部署（systemd + nginx）

### 1. 服务器准备

- Ubuntu 22.04 LTS
- Go 1.23+, Node.js 20+, nginx

### 2. 构建

```bash
cd /opt
git clone https://github.com/chorious/QuestionnaireWithBackend-deploy.git questionnaire
cd questionnaire

# 构建前端
npm ci
npm run build

# 构建后端（单二进制，前端已嵌入）
cd backend-go
cp -r ../dist ./dist
CGO_ENABLED=0 go build -ldflags="-s -w" -o questionnaire-backend .
cd ..
```

### 3. 配置 systemd

```bash
sudo cp systemd/questionnaire.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable questionnaire
sudo systemctl start questionnaire
```

### 4. 配置 nginx（HTTPS）

```bash
sudo cp nginx.conf /etc/nginx/sites-available/questionnaire
sudo ln -s /etc/nginx/sites-available/questionnaire /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

申请 Let's Encrypt 证书：
```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

---

## 环境变量

| 变量 | 必填 | 说明 |
|---|---|---|
| `PORT` | 否 | 服务端口，默认 3000 |
| `ADMIN_TOKEN` | 否 | 管理接口鉴权令牌，不设置则 /submissions 等返回 401 |
| `DB_PATH` | 否 | SQLite 数据库路径，默认 `./data/data.db` |

---

## 目录结构

```
.
|-- backend-go/          # Go 后端源码（含 embed.FS）
|-- src/                 # React 前端源码
|-- Dockerfile           # 多阶段构建镜像
|-- docker-compose.yml   # 一键 Docker 部署
|-- nginx.conf           # nginx 反向代理配置
|-- systemd/             # systemd 服务文件
|-- deploy.md            # 本文件
```

---

## 数据备份

SQLite 数据库文件位置：
- Docker: `./data/data.db`
- systemd: `/opt/questionnaire/data/data.db`

建议定时备份：
```bash
sqlite3 /opt/questionnaire/data/data.db ".backup '/backup/data-$(date +%Y%m%d).db'"
```
