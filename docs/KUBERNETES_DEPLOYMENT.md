# Kubernetes Deployment Guide

Complete guide for deploying CleanCast on Kubernetes clusters.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Namespace Setup](#namespace-setup)
- [Secret Management](#secret-management)
- [ConfigMap](#configmap)
- [Persistent Storage](#persistent-storage)
- [Deployment](#deployment)
- [Service](#service)
- [Ingress](#ingress)
- [Horizontal Pod Autoscaling](#horizontal-pod-autoscaling)
- [Monitoring](#monitoring)
- [Backup Strategy](#backup-strategy)
- [Production Checklist](#production-checklist)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

- Kubernetes cluster 1.20+ (Minikube, Kind, GKE, EKS, AKS, etc.)
- `kubectl` CLI tool installed and configured
- YouTube Data API v3 key
- Storage provisioner configured in your cluster
- Recommended: Helm 3+ for easier management

### Cluster Requirements

- Minimum: 2 nodes with 2GB RAM, 2 CPU cores each
- Storage: Dynamic volume provisioning or manual PV creation
- Load balancer support (for Service type LoadBalancer) or Ingress controller

---

## Quick Start

Deploy CleanCast with minimal configuration:

```bash
# 1. Create namespace
kubectl create namespace clean-cast

# 2. Create secret with API key
kubectl create secret generic clean-cast-secrets \
  --from-literal=GOOGLE_API_KEY=your_api_key_here \
  -n clean-cast

# 3. Apply all manifests
kubectl apply -f k8s/ -n clean-cast

# 4. Wait for deployment
kubectl rollout status deployment/clean-cast -n clean-cast

# 5. Get service endpoint
kubectl get service clean-cast -n clean-cast
```

---

## Namespace Setup

Create a dedicated namespace for CleanCast:

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: clean-cast
  labels:
    app: clean-cast
    environment: production
```

Apply:

```bash
kubectl apply -f namespace.yaml
```

---

## Secret Management

### Using Kubernetes Secrets

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: clean-cast-secrets
  namespace: clean-cast
type: Opaque
stringData:
  GOOGLE_API_KEY: "your_api_key_here"

---
# Optional: S3 backup credentials
apiVersion: v1
kind: Secret
metadata:
  name: clean-cast-s3-secrets
  namespace: clean-cast
type: Opaque
stringData:
  BACKUP_S3_ACCESS_KEY: "AKIA..."
  BACKUP_S3_SECRET_KEY: "secret..."
```

Apply:

```bash
kubectl apply -f secret.yaml
```

### Using External Secrets Operator

For better secret management:

```yaml
# external-secret.yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: clean-cast-secrets
  namespace: clean-cast
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  target:
    name: clean-cast-secrets
  data:
    - secretKey: GOOGLE_API_KEY
      remoteRef:
        key: clean-cast/google-api-key
```

### Using Sealed Secrets

For GitOps workflows:

```bash
# Install kubeseal
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.0/controller.yaml

# Create sealed secret
echo -n 'your_api_key_here' | kubectl create secret generic clean-cast-secrets \
  --dry-run=client \
  --from-file=GOOGLE_API_KEY=/dev/stdin \
  -o yaml | \
  kubeseal -o yaml > sealed-secret.yaml

kubectl apply -f sealed-secret.yaml -n clean-cast
```

---

## ConfigMap

Store non-sensitive configuration:

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: clean-cast-config
  namespace: clean-cast
data:
  # Server configuration
  PORT: "8080"
  HOST: "0.0.0.0"
  CONFIG_DIR: "/config"
  AUDIO_DIR: "/config/"

  # Content configuration
  CRON: "0 2 * * *"
  SPONSORBLOCK_CATEGORIES: "sponsor,intro,outro"
  MIN_DURATION: "5m"

  # Audio configuration
  AUDIO_FORMAT: "m4a"
  AUDIO_QUALITY: "192k"

  # Backup configuration
  BACKUP_CRON: "0 3 * * 0"
  BACKUP_INCLUDE_AUDIO: "false"

  # S3 configuration
  BACKUP_S3_BUCKET: "clean-cast-backups"
  BACKUP_S3_REGION: "us-east-1"
```

Apply:

```bash
kubectl apply -f configmap.yaml
```

---

## Persistent Storage

### Using Dynamic Provisioning

```yaml
# pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: clean-cast-audio-pvc
  namespace: clean-cast
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
  storageClassName: standard  # Use your cluster's storage class
```

### Using Static Provisioning

```yaml
# pv.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: clean-cast-audio-pv
spec:
  capacity:
    storage: 50Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: manual
  hostPath:
    path: /data/clean-cast/audio
    type: DirectoryOrCreate

---
# pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: clean-cast-audio-pvc
  namespace: clean-cast
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
  storageClassName: manual
  volumeName: clean-cast-audio-pv
```

### Using Network Storage (NFS)

```yaml
# pv-nfs.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: clean-cast-audio-nfs
spec:
  capacity:
    storage: 100Gi
  accessModes:
    - ReadWriteMany
  nfs:
    server: nfs-server.example.com
    path: "/exports/clean-cast"
  persistentVolumeReclaimPolicy: Retain
```

Apply:

```bash
kubectl apply -f pvc.yaml
```

---

## Deployment

### Basic Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: clean-cast
  namespace: clean-cast
  labels:
    app: clean-cast
spec:
  replicas: 1  # Single replica recommended due to SQLite
  selector:
    matchLabels:
      app: clean-cast
  template:
    metadata:
      labels:
        app: clean-cast
    spec:
      containers:
      - name: clean-cast
        image: ikoyhn/clean-cast:latest
        imagePullPolicy: Always

        ports:
        - name: http
          containerPort: 8080
          protocol: TCP

        env:
        # Load from ConfigMap
        - name: PORT
          valueFrom:
            configMapKeyRef:
              name: clean-cast-config
              key: PORT
        - name: HOST
          valueFrom:
            configMapKeyRef:
              name: clean-cast-config
              key: HOST
        - name: CRON
          valueFrom:
            configMapKeyRef:
              name: clean-cast-config
              key: CRON
        - name: SPONSORBLOCK_CATEGORIES
          valueFrom:
            configMapKeyRef:
              name: clean-cast-config
              key: SPONSORBLOCK_CATEGORIES
        - name: MIN_DURATION
          valueFrom:
            configMapKeyRef:
              name: clean-cast-config
              key: MIN_DURATION
        - name: AUDIO_FORMAT
          valueFrom:
            configMapKeyRef:
              name: clean-cast-config
              key: AUDIO_FORMAT
        - name: AUDIO_QUALITY
          valueFrom:
            configMapKeyRef:
              name: clean-cast-config
              key: AUDIO_QUALITY

        # Load from Secrets
        - name: GOOGLE_API_KEY
          valueFrom:
            secretKeyRef:
              name: clean-cast-secrets
              key: GOOGLE_API_KEY

        # Optional: S3 credentials
        - name: BACKUP_S3_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: clean-cast-s3-secrets
              key: BACKUP_S3_ACCESS_KEY
              optional: true
        - name: BACKUP_S3_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: clean-cast-s3-secrets
              key: BACKUP_S3_SECRET_KEY
              optional: true

        # Resource limits
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"

        # Liveness probe
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3

        # Readiness probe
        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3

        # Startup probe (for slow starts)
        startupProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 0
          periodSeconds: 10
          timeoutSeconds: 3
          failureThreshold: 30

        volumeMounts:
        - name: audio-storage
          mountPath: /config

      volumes:
      - name: audio-storage
        persistentVolumeClaim:
          claimName: clean-cast-audio-pvc

      # Security context
      securityContext:
        fsGroup: 1000
        runAsNonRoot: true
        runAsUser: 1000

      # Restart policy
      restartPolicy: Always
```

### Production Deployment with All Features

```yaml
# deployment-production.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: clean-cast
  namespace: clean-cast
  labels:
    app: clean-cast
    version: v2.0.0
spec:
  replicas: 1
  revisionHistoryLimit: 10
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
      maxSurge: 1

  selector:
    matchLabels:
      app: clean-cast

  template:
    metadata:
      labels:
        app: clean-cast
        version: v2.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"

    spec:
      # Service account for RBAC
      serviceAccountName: clean-cast

      # Pod anti-affinity (for multi-replica setups)
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - clean-cast
              topologyKey: kubernetes.io/hostname

      # Init container (optional)
      initContainers:
      - name: setup
        image: busybox:latest
        command: ['sh', '-c', 'mkdir -p /config/audio /config/backups && chmod 755 /config']
        volumeMounts:
        - name: audio-storage
          mountPath: /config

      containers:
      - name: clean-cast
        image: ikoyhn/clean-cast:latest
        imagePullPolicy: Always

        ports:
        - name: http
          containerPort: 8080
          protocol: TCP

        envFrom:
        - configMapRef:
            name: clean-cast-config
        - secretRef:
            name: clean-cast-secrets
        - secretRef:
            name: clean-cast-s3-secrets
            optional: true

        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"

        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3

        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3

        startupProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 0
          periodSeconds: 10
          timeoutSeconds: 3
          failureThreshold: 30

        volumeMounts:
        - name: audio-storage
          mountPath: /config
        - name: tmp
          mountPath: /tmp

        # Security
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 1000
          capabilities:
            drop:
            - ALL

      volumes:
      - name: audio-storage
        persistentVolumeClaim:
          claimName: clean-cast-audio-pvc
      - name: tmp
        emptyDir: {}

      securityContext:
        fsGroup: 1000
```

Apply:

```bash
kubectl apply -f deployment.yaml
```

---

## Service

### ClusterIP Service

Internal cluster access only:

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: clean-cast
  namespace: clean-cast
  labels:
    app: clean-cast
spec:
  type: ClusterIP
  ports:
  - port: 8080
    targetPort: http
    protocol: TCP
    name: http
  selector:
    app: clean-cast
```

### LoadBalancer Service

External access with cloud load balancer:

```yaml
# service-loadbalancer.yaml
apiVersion: v1
kind: Service
metadata:
  name: clean-cast
  namespace: clean-cast
  labels:
    app: clean-cast
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: http
    protocol: TCP
    name: http
  selector:
    app: clean-cast
  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 10800
```

### NodePort Service

Access via node IP:

```yaml
# service-nodeport.yaml
apiVersion: v1
kind: Service
metadata:
  name: clean-cast
  namespace: clean-cast
spec:
  type: NodePort
  ports:
  - port: 8080
    targetPort: http
    nodePort: 30080
    protocol: TCP
    name: http
  selector:
    app: clean-cast
```

Apply:

```bash
kubectl apply -f service.yaml
```

---

## Ingress

### Nginx Ingress

```yaml
# ingress-nginx.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: clean-cast
  namespace: clean-cast
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/proxy-body-size: "0"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "600"
spec:
  tls:
  - hosts:
    - podcasts.example.com
    secretName: clean-cast-tls
  rules:
  - host: podcasts.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: clean-cast
            port:
              number: 8080
```

### Traefik Ingress

```yaml
# ingress-traefik.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: clean-cast
  namespace: clean-cast
  annotations:
    kubernetes.io/ingress.class: traefik
    cert-manager.io/cluster-issuer: letsencrypt-prod
    traefik.ingress.kubernetes.io/router.middlewares: clean-cast-ratelimit@kubernetescrd
spec:
  tls:
  - hosts:
    - podcasts.example.com
    secretName: clean-cast-tls
  rules:
  - host: podcasts.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: clean-cast
            port:
              number: 8080
```

### TLS with Cert-Manager

```yaml
# certificate.yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: clean-cast-tls
  namespace: clean-cast
spec:
  secretName: clean-cast-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
  - podcasts.example.com
```

Apply:

```bash
kubectl apply -f ingress-nginx.yaml
```

---

## Horizontal Pod Autoscaling

**Note**: HPA is not recommended for CleanCast due to SQLite's single-writer limitation. Consider vertical scaling instead.

If using a shared database (e.g., PostgreSQL), HPA can be used:

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: clean-cast
  namespace: clean-cast
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: clean-cast
  minReplicas: 1
  maxReplicas: 5
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
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 30
      - type: Pods
        value: 2
        periodSeconds: 30
      selectPolicy: Max
```

---

## Monitoring

### ServiceMonitor for Prometheus Operator

```yaml
# servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: clean-cast
  namespace: clean-cast
  labels:
    app: clean-cast
spec:
  selector:
    matchLabels:
      app: clean-cast
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
    scrapeTimeout: 10s
```

### PodMonitor

```yaml
# podmonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: clean-cast
  namespace: clean-cast
spec:
  selector:
    matchLabels:
      app: clean-cast
  podMetricsEndpoints:
  - port: http
    path: /metrics
```

### Grafana Dashboard

Import dashboard configuration for CleanCast metrics visualization.

---

## Backup Strategy

### CronJob for Automated Backups

```yaml
# cronjob-backup.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: clean-cast-backup
  namespace: clean-cast
spec:
  schedule: "0 3 * * *"  # Daily at 3 AM
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 1
  concurrencyPolicy: Forbid
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: ikoyhn/clean-cast:latest
            command:
            - /bin/sh
            - -c
            - |
              curl -X POST http://clean-cast:8080/api/backup/create \
                -H "Content-Type: application/json" \
                -d '{"include_audio": false, "description": "Scheduled Kubernetes backup"}'
            env:
            - name: BACKUP_INCLUDE_AUDIO
              value: "false"
          restartPolicy: OnFailure
          serviceAccountName: clean-cast
```

### Volume Snapshot

If your storage provider supports snapshots:

```yaml
# volumesnapshot.yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: clean-cast-snapshot
  namespace: clean-cast
spec:
  volumeSnapshotClassName: default-snapshot-class
  source:
    persistentVolumeClaimName: clean-cast-audio-pvc
```

Create snapshot:

```bash
kubectl apply -f volumesnapshot.yaml
```

---

## Production Checklist

Before deploying to production:

### Security
- [ ] Create dedicated namespace
- [ ] Use Kubernetes Secrets for sensitive data
- [ ] Configure RBAC (ServiceAccount, Role, RoleBinding)
- [ ] Enable Pod Security Policies or Pod Security Standards
- [ ] Use NetworkPolicies to restrict traffic
- [ ] Enable TLS/HTTPS via Ingress
- [ ] Set resource limits and requests
- [ ] Run as non-root user
- [ ] Use read-only root filesystem

### Reliability
- [ ] Configure liveness and readiness probes
- [ ] Set up persistent storage with backups
- [ ] Configure appropriate resource requests/limits
- [ ] Set up monitoring with Prometheus
- [ ] Configure log aggregation
- [ ] Test disaster recovery procedures
- [ ] Document rollback procedures

### Performance
- [ ] Right-size resource requests and limits
- [ ] Configure persistent volume with appropriate IOPS
- [ ] Set up CDN if serving globally
- [ ] Configure caching strategies

### Operations
- [ ] Set up CI/CD pipeline
- [ ] Configure automated backups
- [ ] Set up alerting
- [ ] Document operational procedures
- [ ] Plan for updates and rollbacks
- [ ] Configure log rotation

---

## Troubleshooting

### Pod Not Starting

```bash
# Check pod status
kubectl get pods -n clean-cast

# View pod events
kubectl describe pod <pod-name> -n clean-cast

# Check logs
kubectl logs <pod-name> -n clean-cast

# Check previous container logs (if crashed)
kubectl logs <pod-name> -n clean-cast --previous
```

### Storage Issues

```bash
# Check PVC status
kubectl get pvc -n clean-cast

# Describe PVC
kubectl describe pvc clean-cast-audio-pvc -n clean-cast

# Check PV
kubectl get pv

# Check storage class
kubectl get storageclass
```

### Service Not Accessible

```bash
# Check service
kubectl get svc -n clean-cast

# Test service from within cluster
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -n clean-cast -- \
  curl http://clean-cast:8080/health

# Check endpoints
kubectl get endpoints -n clean-cast

# Check ingress
kubectl get ingress -n clean-cast
kubectl describe ingress clean-cast -n clean-cast
```

### Debug Pod

Create a debug pod in the same namespace:

```bash
kubectl run -it --rm debug --image=busybox --restart=Never -n clean-cast -- sh
```

### Common Issues

#### ImagePullBackOff
- Check image name and tag
- Verify image registry credentials
- Check network connectivity

#### CrashLoopBackOff
- Check application logs
- Verify environment variables
- Check resource limits

#### Pending Pod
- Check PVC binding
- Verify node resources
- Check node selectors/affinity

---

## Complete Example

All manifests in one file:

```bash
kubectl apply -f - <<EOF
---
apiVersion: v1
kind: Namespace
metadata:
  name: clean-cast
---
apiVersion: v1
kind: Secret
metadata:
  name: clean-cast-secrets
  namespace: clean-cast
type: Opaque
stringData:
  GOOGLE_API_KEY: "your_api_key_here"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: clean-cast-config
  namespace: clean-cast
data:
  PORT: "8080"
  HOST: "0.0.0.0"
  CRON: "0 2 * * *"
  SPONSORBLOCK_CATEGORIES: "sponsor"
  MIN_DURATION: "5m"
  AUDIO_FORMAT: "m4a"
  AUDIO_QUALITY: "192k"
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: clean-cast-audio-pvc
  namespace: clean-cast
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: clean-cast
  namespace: clean-cast
spec:
  replicas: 1
  selector:
    matchLabels:
      app: clean-cast
  template:
    metadata:
      labels:
        app: clean-cast
    spec:
      containers:
      - name: clean-cast
        image: ikoyhn/clean-cast:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: clean-cast-config
        - secretRef:
            name: clean-cast-secrets
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
        volumeMounts:
        - name: audio-storage
          mountPath: /config
      volumes:
      - name: audio-storage
        persistentVolumeClaim:
          claimName: clean-cast-audio-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: clean-cast
  namespace: clean-cast
spec:
  type: ClusterIP
  ports:
  - port: 8080
    targetPort: 8080
  selector:
    app: clean-cast
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: clean-cast
  namespace: clean-cast
  annotations:
    kubernetes.io/ingress.class: nginx
spec:
  rules:
  - host: podcasts.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: clean-cast
            port:
              number: 8080
EOF
```

---

## Support

For issues or questions:
- GitHub Issues: https://github.com/ikoyhn/clean-cast/issues
- Discussions: https://github.com/ikoyhn/clean-cast/discussions
- Documentation: https://github.com/ikoyhn/clean-cast/tree/main/docs

---

**Last Updated**: 2025-01-15
