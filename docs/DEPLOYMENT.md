# SVG Generator Service - 部署指南

## 快速开始

### 1. 环境准备

确保已安装以下软件：
- Docker (20.10+)
- Docker Compose (2.0+)

### 2. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置文件，填入你的 API 密钥
vim .env
```

### 3. 快速部署

```bash
# 使用部署脚本（推荐）
./scripts/docker.sh deploy

# 或者使用 docker-compose
docker-compose up -d
```

## 详细部署选项

### Docker 脚本部署

使用提供的部署脚本进行管理：

```bash
# 构建镜像
./scripts/docker.sh build

# 运行服务
./scripts/docker.sh run

# 查看状态
./scripts/docker.sh status

# 查看日志
./scripts/docker.sh logs

# 停止服务
./scripts/docker.sh stop

# 重启服务
./scripts/docker.sh restart

# 清理环境
./scripts/docker.sh clean
```

### Docker Compose 部署

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重新构建并启动
docker-compose up -d --build
```

### 手动 Docker 部署

```bash
# 构建镜像
docker build -t svg-generator .

# 运行容器
docker run -d \
  --name svg-generator-app \
  --env-file .env \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -v $(pwd)/logs:/app/logs \
  svg-generator
```

## 服务访问

部署成功后，服务将在以下端点可用：

- **API 服务**: http://localhost:8080
- **健康检查**: http://localhost:8080/health
- **API 文档**: http://localhost:8080/docs（如果已实现）

## API 使用示例

### 生成 SVG 图像

```bash
# 使用 Claude 生成器
curl -X POST "http://localhost:8080/generate" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一个简单的红色圆形",
    "provider": "claude"
  }'

# 使用 Recraft 生成器
curl -X POST "http://localhost:8080/generate" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Modern minimalist logo",
    "provider": "recraft"
  }'

# 使用 SVG.IO 生成器
curl -X POST "http://localhost:8080/generate" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Simple geometric pattern",
    "provider": "svgio"
  }'
```

## 配置说明

### 环境变量

| 变量名 | 描述 | 必需 | 默认值 |
|--------|------|------|--------|
| `PORT` | 服务端口 | 否 | 8080 |
| `GIN_MODE` | Gin 运行模式 | 否 | release |
| `CLAUDE_API_KEY` | Claude API 密钥 | 是 | - |
| `CLAUDE_BASE_URL` | Claude API 基础 URL | 否 | https://api.anthropic.com |
| `RECRAFT_API_KEY` | Recraft API 密钥 | 是 | - |
| `RECRAFT_BASE_URL` | Recraft API 基础 URL | 否 | https://external.api.recraft.ai |
| `SVGIO_API_KEY` | SVG.IO API 密钥 | 是 | - |
| `SVGIO_BASE_URL` | SVG.IO API 基础 URL | 否 | https://svg.io |
| `OPENAI_API_KEY` | OpenAI API 密钥（翻译） | 否 | - |
| `LOG_LEVEL` | 日志级别 | 否 | info |
| `TIMEOUT` | 请求超时时间 | 否 | 30s |

### 配置文件

`config.yaml` 文件包含服务的详细配置。你可以根据需要修改：

```yaml
server:
  port: 8080
  gin_mode: release

providers:
  claude:
    enabled: true
    timeout: 30s
  recraft:
    enabled: true
    timeout: 30s
  svgio:
    enabled: true
    timeout: 30s

logging:
  level: info
  format: json
```

## 监控和日志

### 健康检查

```bash
# 检查服务健康状态
curl http://localhost:8080/health
```

### 查看日志

```bash
# Docker 容器日志
docker logs svg-generator-app

# Docker Compose 日志
docker-compose logs -f

# 应用日志（如果挂载了日志目录）
tail -f logs/app.log
```

## 故障排除

### 常见问题

1. **容器启动失败**
   ```bash
   # 检查环境变量是否正确配置
   docker exec svg-generator-app env | grep API_KEY
   
   # 检查配置文件
   docker exec svg-generator-app cat /app/config.yaml
   ```

2. **API 调用失败**
   ```bash
   # 检查 API 密钥是否有效
   # 查看应用日志了解详细错误信息
   docker logs svg-generator-app
   ```

3. **端口冲突**
   ```bash
   # 修改 .env 文件中的 PORT 变量
   # 或在 docker run 时使用不同端口映射
   docker run -p 9080:8080 ...
   ```

### 性能优化

1. **资源限制**（docker-compose.yml 中已配置）
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '1.0'
         memory: 512M
       reservations:
         memory: 256M
   ```

2. **并发处理**
   - 可以通过 `docker-compose scale` 命令启动多个实例
   - 配置负载均衡器（如 nginx）进行请求分发

## 生产环境建议

1. **安全性**
   - 使用 HTTPS
   - 设置防火墙规则
   - 定期更新 API 密钥
   - 使用 secrets 管理敏感信息

2. **高可用**
   - 使用多实例部署
   - 配置健康检查和自动重启
   - 设置监控和告警

3. **备份**
   - 定期备份配置文件
   - 备份日志文件（如果需要）

4. **监控**
   - 集成 Prometheus + Grafana
   - 设置日志聚合（如 ELK Stack）
   - 配置错误追踪（如 Sentry）
