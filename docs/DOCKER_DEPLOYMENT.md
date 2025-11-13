# Docker Deployment Guide

Complete guide for deploying CleanCast using Docker and Docker Compose.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Docker Run](#docker-run)
- [Docker Compose](#docker-compose)
- [Configuration](#configuration)
- [Volume Management](#volume-management)
- [Networking](#networking)
- [Security](#security)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)
- [Production Checklist](#production-checklist)

---

## Prerequisites

- Docker Engine 20.10+ or Docker Desktop
- Docker Compose 2.0+ (included with Docker Desktop)
- YouTube Data API v3 key
- At least 1GB of available storage (more for audio files)
- Recommended: 2GB RAM, 2 CPU cores

## Quick Start

The fastest way to get CleanCast running:

```bash
# 1. Create directory for configuration
mkdir -p ~/clean-cast/audio

# 2. Create .env file
cat > ~/clean-cast/.env << EOF
GOOGLE_API_KEY=your_api_key_here
AUDIO_DIR_PATH=$(pwd)/audio
EOF

# 3. Download docker-compose.yml
cd ~/clean-cast
curl -O https://raw.githubusercontent.com/ikoyhn/clean-cast/main/docker-compose.yml

# 4. Start the service
docker-compose up -d

# 5. Check status
docker-compose ps
docker-compose logs -f
```

Access the service at `http://localhost:8080`

---

## Docker Run

### Basic Deployment

Minimal deployment with required parameters only:

```bash
docker run -d \
  --name clean-cast \
  --restart unless-stopped \
  -p 8080:8080 \
  -e GOOGLE_API_KEY=your_api_key_here \
  -v /path/to/audio:/config \
  ikoyhn/clean-cast:latest
```

### Full Configuration

Complete deployment with all options:

```bash
docker run -d \
  --name clean-cast \
  --restart unless-stopped \
  -p 8080:8080 \
  \
  -e GOOGLE_API_KEY=your_api_key_here \
  \
  -e PORT=8080 \
  -e HOST=0.0.0.0 \
  \
  -e CRON="0 2 * * *" \
  -e SPONSORBLOCK_CATEGORIES="sponsor,intro,outro" \
  -e MIN_DURATION=5m \
  -e COOKIES_FILE=cookies.txt \
  \
  -e AUDIO_FORMAT=m4a \
  -e AUDIO_QUALITY=192k \
  \
  -e BACKUP_CRON="0 3 * * 0" \
  -e BACKUP_INCLUDE_AUDIO=false \
  \
  -e BACKUP_S3_BUCKET=my-backups \
  -e BACKUP_S3_REGION=us-east-1 \
  -e BACKUP_S3_ACCESS_KEY=AKIA... \
  -e BACKUP_S3_SECRET_KEY=secret... \
  \
  -v /path/to/audio:/config \
  -v /path/to/cookies.txt:/config/cookies.txt:ro \
  \
  --memory="2g" \
  --cpus="2.0" \
  \
  --health-cmd="wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1" \
  --health-interval=30s \
  --health-timeout=10s \
  --health-retries=3 \
  --health-start-period=40s \
  \
  ikoyhn/clean-cast:latest
```

### Platform-Specific Images

CleanCast supports multiple architectures:

```bash
# AMD64 (x86_64)
docker run --platform linux/amd64 ... ikoyhn/clean-cast:latest

# ARM64 (Apple Silicon, Raspberry Pi 4)
docker run --platform linux/arm64 ... ikoyhn/clean-cast:latest

# Auto-detect (recommended)
docker run ... ikoyhn/clean-cast:latest
```

---

## Docker Compose

### Basic docker-compose.yml

Minimal configuration:

```yaml
version: '3.8'

services:
  clean-cast:
    image: ikoyhn/clean-cast:latest
    container_name: clean-cast
    restart: unless-stopped

    ports:
      - "8080:8080"

    environment:
      - GOOGLE_API_KEY=${GOOGLE_API_KEY}

    volumes:
      - audio_data:/config

volumes:
  audio_data:
```

### Production docker-compose.yml

Full production configuration:

```yaml
version: '3.8'

services:
  clean-cast:
    image: ikoyhn/clean-cast:latest
    container_name: clean-cast
    restart: unless-stopped

    # Multi-platform build
    build:
      context: .
      dockerfile: Dockerfile
      platforms:
        - linux/amd64
        - linux/arm64

    ports:
      - "${HOST_PORT:-8080}:8080"

    environment:
      # Required
      - GOOGLE_API_KEY=${GOOGLE_API_KEY}

      # Server
      - PORT=${PORT:-8080}
      - HOST=0.0.0.0
      - CONFIG_DIR=/config
      - AUDIO_DIR=/config/

      # Content
      - CRON=${CRON:-0 0 * * 0}
      - SPONSORBLOCK_CATEGORIES=${SPONSORBLOCK_CATEGORIES:-sponsor}
      - MIN_DURATION=${MIN_DURATION:-5m}
      - COOKIES_FILE=${COOKIES_FILE:-}

      # Audio
      - AUDIO_FORMAT=${AUDIO_FORMAT:-m4a}
      - AUDIO_QUALITY=${AUDIO_QUALITY:-192k}

      # Backup
      - BACKUP_CRON=${BACKUP_CRON:-}
      - BACKUP_INCLUDE_AUDIO=${BACKUP_INCLUDE_AUDIO:-false}

      # S3 Backup
      - BACKUP_S3_BUCKET=${BACKUP_S3_BUCKET:-}
      - BACKUP_S3_REGION=${BACKUP_S3_REGION:-}
      - BACKUP_S3_ACCESS_KEY=${BACKUP_S3_ACCESS_KEY:-}
      - BACKUP_S3_SECRET_KEY=${BACKUP_S3_SECRET_KEY:-}

    volumes:
      - audio_data:/config
      # Optional: bind mount for cookies
      # - ./cookies.txt:/config/cookies.txt:ro

    # Resource limits
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M

    # Health check
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

    # Logging
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
        labels: "service=clean-cast"

    # Security
    security_opt:
      - no-new-privileges:true

    # Network
    networks:
      - clean-cast-network

volumes:
  audio_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: ${AUDIO_DIR_PATH:-./audio}

networks:
  clean-cast-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.25.0.0/16
```

### Docker Compose Commands

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# View logs for last 100 lines
docker-compose logs --tail=100 -f

# Check status
docker-compose ps

# Stop services
docker-compose stop

# Restart services
docker-compose restart

# Remove services (keeps volumes)
docker-compose down

# Remove services and volumes (deletes data!)
docker-compose down -v

# Update to latest image
docker-compose pull
docker-compose up -d

# View resource usage
docker-compose stats
```

---

## Configuration

### Using .env File

Create a `.env` file in the same directory as `docker-compose.yml`:

```bash
# .env file
GOOGLE_API_KEY=AIzaSyC...
HOST_PORT=8080
CRON=0 2 * * *
SPONSORBLOCK_CATEGORIES=sponsor,intro,outro
MIN_DURATION=10m
AUDIO_FORMAT=m4a
AUDIO_QUALITY=192k
AUDIO_DIR_PATH=/data/clean-cast/audio
BACKUP_CRON=0 3 * * 0
```

### Using Environment Variables

Set variables in your shell before running Docker:

```bash
export GOOGLE_API_KEY="your_api_key_here"
export AUDIO_DIR_PATH="/data/clean-cast/audio"
docker-compose up -d
```

### Using Docker Secrets (Swarm Mode)

For Docker Swarm deployments:

```bash
# Create secret
echo "your_api_key_here" | docker secret create google_api_key -

# Update docker-compose.yml
services:
  clean-cast:
    secrets:
      - google_api_key
    environment:
      - GOOGLE_API_KEY_FILE=/run/secrets/google_api_key

secrets:
  google_api_key:
    external: true
```

---

## Volume Management

### Bind Mount (Recommended for Development)

```yaml
volumes:
  - /host/path/to/audio:/config
```

**Pros**:
- Easy to access files on host
- Simple backup with standard tools
- Direct file management

**Cons**:
- Path must exist on host
- Permissions can be tricky

### Named Volume (Recommended for Production)

```yaml
volumes:
  - audio_data:/config

volumes:
  audio_data:
    driver: local
```

**Pros**:
- Docker manages the volume
- Better performance
- Portable across hosts

**Cons**:
- Harder to access files directly
- Need Docker commands for backup

### Volume Commands

```bash
# List volumes
docker volume ls

# Inspect volume
docker volume inspect clean-cast_audio_data

# Backup volume
docker run --rm \
  -v clean-cast_audio_data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/audio_backup.tar.gz -C /data .

# Restore volume
docker run --rm \
  -v clean-cast_audio_data:/data \
  -v $(pwd):/backup \
  alpine tar xzf /backup/audio_backup.tar.gz -C /data

# Remove volume (warning: deletes data!)
docker volume rm clean-cast_audio_data
```

---

## Networking

### Default Bridge Network

Simplest configuration, sufficient for most use cases:

```yaml
# No network configuration needed
```

### Custom Bridge Network

Better isolation and DNS:

```yaml
networks:
  - clean-cast-network

networks:
  clean-cast-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.25.0.0/16
```

### Host Network

Direct access to host network (less secure):

```yaml
network_mode: "host"
```

### Reverse Proxy Integration

#### With Nginx

```nginx
# nginx.conf
upstream clean-cast {
    server localhost:8080;
}

server {
    listen 80;
    server_name podcasts.example.com;

    location / {
        proxy_pass http://clean-cast;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

#### With Traefik

```yaml
services:
  clean-cast:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.clean-cast.rule=Host(`podcasts.example.com`)"
      - "traefik.http.routers.clean-cast.entrypoints=websecure"
      - "traefik.http.routers.clean-cast.tls.certresolver=letsencrypt"
      - "traefik.http.services.clean-cast.loadbalancer.server.port=8080"
```

---

## Security

### Security Best Practices

1. **Run as Non-Root User**

```dockerfile
# In Dockerfile
USER nobody:nobody
```

2. **Read-Only Root Filesystem**

```yaml
read_only: true
tmpfs:
  - /tmp
  - /var/run
```

3. **Drop Capabilities**

```yaml
cap_drop:
  - ALL
cap_add:
  - NET_BIND_SERVICE  # If binding to port < 1024
```

4. **Security Options**

```yaml
security_opt:
  - no-new-privileges:true
  - apparmor=docker-default
```

5. **Secrets Management**

Never commit secrets to git. Use:
- Docker secrets (Swarm)
- Environment files with restricted permissions
- External secret management (Vault, AWS Secrets Manager)

```bash
# Restrict .env file permissions
chmod 600 .env
```

### Firewall Configuration

```bash
# Allow only necessary ports
sudo ufw allow 8080/tcp comment 'CleanCast HTTP'
sudo ufw enable
```

---

## Monitoring

### Health Checks

Built-in health check:

```yaml
healthcheck:
  test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

Check health status:

```bash
# View health status
docker inspect --format='{{.State.Health.Status}}' clean-cast

# View health logs
docker inspect --format='{{range .State.Health.Log}}{{.Output}}{{end}}' clean-cast
```

### Prometheus Metrics

CleanCast exposes Prometheus metrics at `/metrics`:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'clean-cast'
    static_configs:
      - targets: ['clean-cast:8080']
```

### Logging

View and follow logs:

```bash
# All logs
docker logs clean-cast

# Follow logs
docker logs -f clean-cast

# Last 100 lines
docker logs --tail=100 clean-cast

# Since timestamp
docker logs --since=2025-01-15T10:00:00 clean-cast

# With timestamps
docker logs -t clean-cast
```

Configure logging driver:

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"

# Or use syslog
logging:
  driver: "syslog"
  options:
    syslog-address: "udp://127.0.0.1:514"
    tag: "clean-cast"
```

---

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker logs clean-cast

# Check for port conflicts
sudo netstat -tlnp | grep :8080

# Check file permissions
ls -la /path/to/audio

# Verify environment variables
docker inspect clean-cast | grep -A 20 Env
```

### High Memory Usage

```bash
# Check resource usage
docker stats clean-cast

# Set memory limits
docker update --memory 2g --memory-swap 2g clean-cast
```

### Audio Files Not Downloading

```bash
# Check network connectivity
docker exec clean-cast wget -O- https://www.youtube.com

# Check disk space
docker exec clean-cast df -h

# Check permissions
docker exec clean-cast ls -la /config
```

### Database Issues

```bash
# Backup database
docker exec clean-cast cp /config/sqlite.db /config/sqlite.db.backup

# Check database integrity
docker exec clean-cast sqlite3 /config/sqlite.db "PRAGMA integrity_check;"
```

---

## Production Checklist

Before deploying to production:

- [ ] Set a strong `GOOGLE_API_KEY`
- [ ] Configure appropriate resource limits
- [ ] Set up volume backups
- [ ] Configure health checks
- [ ] Set up log rotation
- [ ] Enable monitoring (Prometheus)
- [ ] Configure reverse proxy with HTTPS
- [ ] Set up automated backups
- [ ] Configure S3 backup if needed
- [ ] Test disaster recovery
- [ ] Document your configuration
- [ ] Set up alerting
- [ ] Review security settings
- [ ] Test with actual workload
- [ ] Plan for updates
- [ ] Configure log aggregation

---

## Advanced Topics

### Multi-Container Setup

Run multiple instances with different configurations:

```yaml
services:
  clean-cast-tech:
    image: ikoyhn/clean-cast:latest
    ports:
      - "8081:8080"
    environment:
      - GOOGLE_API_KEY=${GOOGLE_API_KEY}
    volumes:
      - tech_podcasts:/config

  clean-cast-news:
    image: ikoyhn/clean-cast:latest
    ports:
      - "8082:8080"
    environment:
      - GOOGLE_API_KEY=${GOOGLE_API_KEY}
    volumes:
      - news_podcasts:/config
```

### Docker Swarm Deployment

```yaml
version: '3.8'

services:
  clean-cast:
    image: ikoyhn/clean-cast:latest
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 10s
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M
    ports:
      - "8080:8080"
    networks:
      - clean-cast-overlay

networks:
  clean-cast-overlay:
    driver: overlay
```

Deploy to swarm:

```bash
docker stack deploy -c docker-compose.yml clean-cast
```

---

## Support

For issues or questions:
- GitHub Issues: https://github.com/ikoyhn/clean-cast/issues
- Discussions: https://github.com/ikoyhn/clean-cast/discussions
- Documentation: https://github.com/ikoyhn/clean-cast/tree/main/docs

---

**Last Updated**: 2025-01-15
