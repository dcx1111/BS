# 图片管理系统

基于 Docker Compose 的图片管理系统，支持图片上传、管理、标签、轮播等功能。

## 学号姓名

### 学号 3230103537
### 姓名 董宸轩
### 仓库地址 https://github.com/dcx1111/BS

## 前置要求

- Docker
- Docker Compose

## 快速开始

### 1. 配置环境变量

复制环境变量模板文件：

```bash
cp env.example .env
```

编辑 `.env` 文件，至少修改以下关键配置：

- `JWT_SECRET`: JWT 密钥（生产环境必须修改）
- `DB_PASSWORD`: 数据库密码（默认：rootpassword）
- `AI_API_KEY`: AI API 密钥（如需 AI 功能）
- `AI_ENABLED`: 是否启用 AI 功能（默认：false）

其他配置可使用默认值，详细说明见 `env.example`。
由于验收需要，部分内容已经在原有 `.env` 里写好了。

### 2. 启动服务

```bash
docker-compose up -d
```

### 3. 访问应用

- 前端：http://localhost
- 后端 API：http://localhost:8080/api/v1

## 提示

- git日志，录像等都在docs文件夹下
