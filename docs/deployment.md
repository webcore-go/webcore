# Deployment Guide

This guide covers how to deploy the WebCoreGo API in various environments.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Local Development](#local-development)
3. [Docker Deployment](#docker-deployment)
4. [Kubernetes Deployment](#kubernetes-deployment)
5. [Cloud Deployment](#cloud-deployment)
6. [Production Considerations](#production-considerations)
7. [Monitoring and Logging](#monitoring-and-logging)
8. [Backup and Recovery](#backup-and-recovery)

## Prerequisites

Before deploying, ensure you have:

- Go 1.19 or higher
- Docker (optional, for containerized deployment)
- PostgreSQL 12 or higher
- Redis 6 or higher
- Nginx (optional, for reverse proxy)
- SSL certificate (for production)

## Local Development

### 1. Clone the Repository

```bash
git clone <repository-url>
cd webcore-go
```

### 2. Install Dependencies

```bash
go mod download
go mod tidy
```

### 3. Set Up Environment Variables

Create a `.env` file:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```bash
# Application
APP_ENV=development
APP_DEBUG=true
APP_PORT=3000

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=konsolidator
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your_jwt_secret
JWT_EXPIRES_IN=3600
JWT_REFRESH_EXPIRES_IN=86400

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=100
RATE_LIMIT_REQUESTS_PER_HOUR=1000

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json
LOG_OUTPUT=stdout
```

### 4. Start Services

Using Docker Compose:

```bash
docker-compose up -d
```

Or manually:

```bash
# Start PostgreSQL
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=konsolidator \
  -p 5432:5432 \
  postgres:13

# Start Redis
docker run -d --name redis \
  -p 6379:6379 \
  redis:6-alpine
```

### 5. Run Migrations

```bash
go run cmd/migrate/main.go
```

### 6. Start the Application

```bash
go run main.go
```

The API will be available at `http://localhost:3000`

## Docker Deployment

### 1. Build the Docker Image

```bash
# Build the image
docker build -t konsolidator-api:latest .

# Build with specific tag
docker build -t konsolidator-api:v1.0.0 .
```

### 2. Run with Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "3000:3000"
    environment:
      - APP_ENV=production
      - APP_DEBUG=false
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=konsolidator
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres
      - redis
    volumes:
      - ./logs:/app/logs
    restart: unless-stopped

  postgres:
    image: postgres:13
    environment:
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=konsolidator
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    restart: unless-stopped

  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - app
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

Start the services:

```bash
docker-compose up -d
```

### 3. Multi-Stage Build

For optimized production images:

```dockerfile
# Build stage
FROM golang:1.19-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy binary from build stage
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config

# Copy SSL certificates if needed
COPY --from=builder /app/ssl ./ssl

EXPOSE 3000

CMD ["./main"]
```

## Kubernetes Deployment

### 1. Create Kubernetes Resources

#### Namespace

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: konsolidator
```

#### ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: konsolidator-config
  namespace: konsolidator
data:
  config.yaml: |
    app:
      environment: production
      debug: false
      port: 3000
    
    database:
      host: postgres-service
      port: 5432
      user: postgres
      password: ${DB_PASSWORD}
      name: konsolidator
      sslmode: disable
      max_connections: 25
      max_idle_connections: 5
      max_lifetime_hours: 1
    
    redis:
      host: redis-service
      port: 6379
      password: ${REDIS_PASSWORD}
      db: 0
      pool_size: 10
      min_idle_connections: 5
      dial_timeout_seconds: 5
      read_timeout_seconds: 3
      write_timeout_seconds: 3
    
    jwt:
      secret: ${JWT_SECRET}
      expires_in: 3600
      refresh_expires_in: 86400
    
    rate_limit:
      requests_per_minute: 100
      requests_per_hour: 1000
    
    logging:
      level: info
      format: json
      output: stdout
```

#### Secret

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: konsolidator-secret
  namespace: konsolidator
type: Opaque
data:
  db-password: <base64-encoded-password>
  redis-password: <base64-encoded-password>
  jwt-secret: <base64-encoded-secret>
```

#### Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: konsolidator-api
  namespace: konsolidator
  labels:
    app: konsolidator-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: konsolidator-api
  template:
    metadata:
      labels:
        app: konsolidator-api
    spec:
      containers:
      - name: api
        image: konsolidator-api:v1.0.0
        ports:
        - containerPort: 3000
        env:
        - name: APP_ENV
          value: "production"
        - name: DB_HOST
          value: "postgres-service"
        - name: REDIS_HOST
          value: "redis-service"
        envFrom:
        - secretRef:
            name: konsolidator-secret
        - configMapRef:
            name: konsolidator-config
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: logs
          mountPath: /app/logs
      volumes:
      - name: logs
        emptyDir: {}
      restartPolicy: Always
```

#### Service

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: konsolidator-service
  namespace: konsolidator
spec:
  selector:
    app: konsolidator-api
  ports:
  - protocol: TCP
    port: 80
    targetPort: 3000
  type: ClusterIP
```

#### Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: konsolidator-ingress
  namespace: konsolidator
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - api.konsolidator.com
    secretName: konsolidator-tls
  rules:
  - host: api.konsolidator.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: konsolidator-service
            port:
              number: 80
```

### 2. Apply Resources

```bash
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f ingress.yaml
```

### 3. Verify Deployment

```bash
kubectl get pods -n konsolidator
kubectl get services -n konsolidator
kubectl get ingress -n konsolidator
```

## Cloud Deployment

### AWS Deployment

#### Using AWS ECS

1. Build and push Docker image to ECR:

```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <aws_account_id>.dkr.ecr.us-east-1.amazonaws.com

# Build and push
docker build -t konsolidator-api .
docker tag konsolidator-api:latest <aws_account_id>.dkr.ecr.us-east-1.amazonaws.com/konsolidator-api:latest
docker push <aws_account_id>.dkr.ecr.us-east-1.amazonaws.com/konsolidator-api:latest
```

2. Create ECS Task Definition:

```json
{
  "family": "konsolidator-api",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::<aws_account_id>:role/ecsTaskExecutionRole",
  "containerDefinitions": [
    {
      "name": "konsolidator-api",
      "image": "<aws_account_id>.dkr.ecr.us-east-1.amazonaws.com/konsolidator-api:latest",
      "portMappings": [
        {
          "containerPort": 3000,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "APP_ENV",
          "value": "production"
        },
        {
          "name": "DB_HOST",
          "value": "<rds-endpoint>"
        },
        {
          "name": "REDIS_HOST",
          "value": "<elasticache-endpoint>"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/konsolidator-api",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

3. Create ECS Service and Load Balancer

#### Using AWS EKS

Follow the Kubernetes deployment guide and deploy to EKS instead of local Kubernetes.

### Google Cloud Platform (GCP) Deployment

#### Using Google Kubernetes Engine (GKE)

1. Build and push Docker image to Google Container Registry (GCR):

```bash
# Build and push
docker build -t gcr.io/<project-id>/konsolidator-api:latest .
docker push gcr.io/<project-id>/konsolidator-api:latest
```

2. Update the Kubernetes deployment to use the GCR image.

#### Using Cloud Run

```bash
# Deploy to Cloud Run
gcloud run deploy konsolidator-api \
  --image gcr.io/<project-id>/konsolidator-api:latest \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars="APP_ENV=production,DB_HOST=<cloudsql-instance>,REDIS_HOST=<redis-instance>"
```

### Azure Deployment

#### Using Azure Kubernetes Service (AKS)

1. Build and push Docker image to Azure Container Registry (ACR):

```bash
# Build and push
docker build -t <acr-name>.azurecr.io/konsolidator-api:latest .
docker push <acr-name>.azurecr.io/konsolidator-api:latest
```

2. Update the Kubernetes deployment to use the ACR image.

#### Using Azure App Service

```bash
# Deploy to Azure App Service
az webapp create \
  --resource-group <resource-group> \
  --plan <app-service-plan> \
  --name konsolidator-api \
  --runtime "DOTNETCORE:6.0" \
  --deployment-container-image-name <acr-name>.azurecr.io/konsolidator-api:latest
```

## Production Considerations

### 1. Security

#### SSL/TLS Configuration

```nginx
# nginx.conf
server {
    listen 80;
    server_name api.konsolidator.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.konsolidator.com;
    
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

#### Environment Variables

Never commit sensitive information to version control. Use environment variables or secrets management:

```bash
# AWS Secrets Manager
aws secretsmanager get-secret-value --secret-id konsolidator-api-secret --query SecretString --output text

# Azure Key Vault
az keyvault secret show --name konsolidator-api-secret --vault-name <key-vault-name>
```

### 2. Performance Optimization

#### Database Optimization

```yaml
# config.yaml
database:
  # Connection pooling
  max_connections: 50
  max_idle_connections: 10
  max_lifetime_hours: 2
  
  # Query optimization
  max_open_files: 1000
  max_idle_time: "5m"
  
  # Performance monitoring
  slow_query_threshold: "100ms"
  enable_sql_logging: false
```

#### Redis Optimization

```yaml
redis:
  # Connection pooling
  pool_size: 20
  min_idle_connections: 5
  
  # Performance
  dial_timeout_seconds: 5
  read_timeout_seconds: 3
  write_timeout_seconds: 3
  pool_timeout_seconds: 4
  
  # Memory optimization
  max_memory_mb: 1024
  max_memory_policy: allkeys-lru
```

### 3. Monitoring and Logging

#### Structured Logging

```go
// Configure structured logging
logger := shared.NewLogger()
logger.SetLevel("info")
logger.SetFormat("json")
logger.SetOutput("stdout")
```

#### Health Checks

```yaml
# deployment.yaml
livenessProbe:
  httpGet:
    path: /health
    port: 3000
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /health
    port: 3000
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

### 4. Scaling

#### Horizontal Scaling

```yaml
# deployment.yaml
replicas: 3
```

#### Vertical Scaling

```yaml
# deployment.yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

#### Auto-scaling

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: konsolidator-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: konsolidator-api
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

## Monitoring and Logging

### 1. Application Metrics

#### Prometheus Integration

```go
// metrics.go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint"},
    )
)
```

#### Grafana Dashboard

Create a Grafana dashboard with:

- Request rate and latency
- Error rate by endpoint
- Database connection pool metrics
- Redis usage metrics
- Memory and CPU usage

### 2. Logging

#### Structured Logging

```go
// Configure structured logging
logger := shared.NewLogger()
logger.SetLevel("info")
logger.SetFormat("json")
logger.SetOutput("stdout")

// Log structured data
logger.Info("API request",
    "method", "GET",
    "path", "/api/v1/items",
    "status", 200,
    "duration", "15ms",
    "user_id", 123,
)
```

#### Log Aggregation

Use ELK stack or similar for log aggregation:

- Elasticsearch: Store and search logs
- Logstash: Process and transform logs
- Kibana: Visualize logs

### 3. Distributed Tracing

#### Jaeger Integration

```go
// tracing.go
package main

import (
    "github.com/uber/jaeger-client-go"
    "github.com/uber/jaeger-client-go/config"
)

func initTracing(serviceName string) (tracer *jaeger.Tracer, closer io.Closer, err error) {
    cfg := &config.Configuration{
        ServiceName: serviceName,
        Sampler: &config.SamplerConfig{
            Type:  jaeger.SamplerTypeConst,
            Param: 1,
        },
        Reporter: &config.ReporterConfig{
            LogSpans: true,
        },
    }
    
    tracer, closer, err = cfg.NewTracer()
    return
}
```

## Backup and Recovery

### 1. Database Backup

#### PostgreSQL

```bash
# Create backup
pg_dump -h <host> -U <user> -d <database> > backup.sql

# Automated backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -h localhost -U postgres -d konsolidator > /backups/konsolidator_$DATE.sql
gzip /backups/konsolidator_$DATE.sql
```

#### Restore

```bash
# Restore from backup
gunzip -c backup.sql.gz | psql -h localhost -U postgres -d konsolidator
```

### 2. Redis Backup

```bash
# Save Redis data
redis-cli BGSAVE

# Copy RDB file
cp /var/lib/redis/dump.rdb /backups/redis_backup_$(date +%Y%m%d).rdb
```

### 3. Configuration Backup

```bash
# Backup configuration files
tar -czf config_backup_$(date +%Y%m%d).tar.gz \
    /etc/nginx/ \
    /etc/konsolidator/ \
    /var/www/konsolidator/config/
```

### 4. Disaster Recovery Plan

1. **Regular Backups**: Daily database and configuration backups
2. **Off-site Storage**: Store backups in different geographic locations
3. **Testing**: Regularly test backup restoration
4. **Documentation**: Document recovery procedures
5. **Automation**: Automate backup and recovery processes

## Troubleshooting

### Common Issues

1. **Application Won't Start**
   - Check logs for errors
   - Verify configuration files
   - Ensure dependencies are running

2. **Database Connection Issues**
   - Check database connectivity
   - Verify credentials
   - Check connection pool settings

3. **Memory Issues**
   - Monitor memory usage
   - Adjust resource limits
   - Check for memory leaks

4. **Performance Issues**
   - Monitor response times
   - Check database queries
   - Review caching strategies

### Debug Mode

Enable debug mode in production for troubleshooting:

```bash
# Set environment variable
export APP_DEBUG=true

# Or in configuration
app:
  debug: true
```

### Log Analysis

```bash
# View application logs
kubectl logs -f deployment/konsolidator-api -n konsolidator

# View specific container logs
docker logs -f konsolidator-api

# Filter logs by error
grep -i error /var/log/konsolidator/app.log
```

## Support

For deployment support:

1. Check the troubleshooting section
2. Review logs for error messages
3. Verify system requirements
4. Contact the DevOps team

## Conclusion

This deployment guide provides comprehensive instructions for deploying the WebCoreGo API in various environments. Always follow best practices for security, performance, and reliability in production deployments.

Remember to:
- Use proper secrets management
- Monitor application performance
- Regularly update dependencies
- Test backup and recovery procedures
- Document deployment procedures