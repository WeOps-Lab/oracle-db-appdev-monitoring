apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: oracledb-exporter-dg-{{VERSION}}
  namespace: oracle
spec:
  serviceName: oracledb-exporter-dg-{{VERSION}}
  replicas: 1
  selector:
    matchLabels:
      app: oracledb-exporter-dg-{{VERSION}}
  template:
    metadata:
      annotations:
        telegraf.influxdata.com/interval: 1s
        telegraf.influxdata.com/inputs: |+
          [[inputs.cpu]]
            percpu = false
            totalcpu = true
            collect_cpu_time = true
            report_active = true

          [[inputs.disk]]
            ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]

          [[inputs.diskio]]

          [[inputs.kernel]]

          [[inputs.mem]]

          [[inputs.processes]]

          [[inputs.system]]
            fielddrop = ["uptime_format"]

          [[inputs.net]]
            ignore_protocol_stats = true

          [[inputs.procstat]]
          ## pattern as argument for pgrep (ie, pgrep -f <pattern>)
            pattern = "exporter"
        telegraf.influxdata.com/class: opentsdb
        telegraf.influxdata.com/env-fieldref-NAMESPACE: metadata.namespace
        telegraf.influxdata.com/limits-cpu: '300m'
        telegraf.influxdata.com/limits-memory: '300Mi'
      labels:
        app: oracledb-exporter-dg-{{VERSION}}
        exporter_object: oracledb
        object_mode: dg
        object_version: {{VERSION}}
        pod_type: exporter
    spec:
      nodeSelector:
        node-role: worker
      shareProcessNamespace: true
      containers:
      - name: oracledb-exporter-dg-{{VERSION}}
        image: registry-svc:25000/library/oracledb-exporter:latest
        imagePullPolicy: Always
        args:
        - --isDataGuard
        - --query.timeout=3
        env:
        - name: DATA_SOURCE_NAME
          valueFrom:
            configMapKeyRef:
              name: oracle-rac-dsn
              key: DATA_SOURCE_NAME_{{VERSION}}_dg
        securityContext:
          allowPrivilegeEscalation: false
          runAsUser: 0
        resources:
          requests:
            cpu: 100m
            memory: 100Mi
          limits:
            cpu: 1000m
            memory: 300Mi
        ports:
        - containerPort: 9161

---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: oracledb-exporter-dg-{{VERSION}}
  name: oracledb-exporter-dg-{{VERSION}}
  namespace: oracle
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "9161"
    prometheus.io/path: '/metrics'
spec:
  ports:
  - port: 9161
    protocol: TCP
    targetPort: 9161
  selector:
    app: oracledb-exporter-dg-{{VERSION}}
