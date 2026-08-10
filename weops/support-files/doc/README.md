## 嘉为蓝鲸oracledb插件使用说明

## 使用说明

### 插件功能

采集器通过 Oracle 服务地址连接数据库，执行只读 SQL 查询，并将查询结果转换为 Prometheus 指标。
实际可采集指标取决于 Oracle 版本、部署架构、功能开关及监控账号授权。

### 版本支持

操作系统支持: linux, windows

是否支持arm: 支持

**组件支持版本：**

Oracle Database: `11g`, `12c`, `18c`, `19c`, `21c`

部署模式：`standalone（单点）`、`RAC（集群）`、`Data Guard（DG）`

**是否支持远程采集:**

是

### 配置指引

#### 1. 配置前确认

配置采集器前，先向数据库管理员确认以下信息：

- Oracle 主机名或 IP、监听端口，默认端口为 `1521`；
- 监听器实际注册的服务名；
- 目标是非 CDB、CDB 根容器还是指定 PDB；
- 是否启用 RAC、ASM、Data Guard 或归档日志采集；
- 监控平台的采集节点到 Oracle 监听端口的网络是否连通。

#### 2. 连接参数

| **参数名**           | **含义**                                  | **是否必填** | **使用举例**   |
|----------------------|-------------------------------------------|--------------|----------------|
| --host               | 数据库主机IP                              | 是           | 127.0.0.1      |
| --port               | 数据库服务端口                            | 是           | 1521           |
| USER                 | 数据库用户名(环境变量),特殊字符不需要转义 | 是           |                |
| PASSWORD             | 数据库密码(环境变量),特殊字符不需要转义   | 是           |                |
| SERVICE_NAME         | 数据库服务名(环境变量)                    | 是           | ORCLCDB        |
| --isRAC              | 是否为rac集群架构(开关参数), 默认不开启   | 否           |                |
| --isASM              | 是否有ASM磁盘组(开关参数), 默认不开启     | 否           |                |
| --isDataGuard        | 是否为DataGuard(开关参数), 默认不开启     | 否           |                |
| --isArchiveLog       | 是否采集归档日志指标, 默认不开启          | 否           |                |
| --query.timeout      | 查询超时秒数，默认使用5s                  | 否           | 5              |
| --log.level          | 日志级别                                  | 否           | info           |
| --web.listen-address | Exporter 监听地址和端口                   | 否           | :9161          |

> `oracledb_redo_log_switches_1h` 中的“最近 1 小时”是 SQL 统计窗口，不是采集周期。指标仍会在每次抓取 `/metrics` 时重新计算。窗口值适合直接告警和比较近期切换压力，历史趋势由监控系统保留每次窗口值。未使用实例启动以来的累计值，是因为累计值随运行时长单调增长且会在实例重启后归零，不适合直接设置通用阈值。

#### 3. 确认服务名

应使用监听器实际注册的完整服务名。服务名是否包含域名取决于数据库和监听器配置，不由 Oracle 版本单独决定。

管理员可执行：

```sql
SELECT value FROM v$parameter WHERE name = 'service_names';
SELECT value FROM v$parameter WHERE name = 'db_domain';
SELECT name, network_name, con_id FROM v$services ORDER BY con_id, name;
```

数据库主机上还可以执行：

```shell
lsnrctl status
```

配置前先使用监控账号验证连接：

```shell
sqlplus 'username/password@//host:1521/service_name'
```

#### 4. 选择监控账号所在容器

| 数据库场景 | 建议账号 |
|---|---|
| Oracle 11g 或非 CDB | 创建普通本地用户，例如 `WEOPS_MONITOR` |
| 仅监控一个 PDB | 切换到目标 PDB 后创建本地用户，例如 `WEOPS_MONITOR` |
| 需要从 CDB 根容器监控多个容器 | 由 DBA 创建公共用户，例如 `C##WEOPS_MONITOR` |

不要仅因为 Oracle 版本为 12c 及以上就使用 `C##` 前缀；是否使用公共用户取决于采集范围和实际连接的容器。

#### 5. 创建账号并授权

以下 SQL 必须由数据库管理员在采集器实际连接的容器中执行。将 `username` 和 `password` 替换为实际值；密码包含特殊字符时使用双引号。对象授权必须使用底层对象名 `V_$...`，不能对同义词 `V$...` 授权。

```sql
-- 非CDB或PDB本地用户
CREATE USER username IDENTIFIED BY "password";

-- 如果需要CDB公共用户，由DBA改用以下形式并调整username
-- CREATE USER C##WEOPS_MONITOR IDENTIFIED BY "password" CONTAINER = ALL;

GRANT CREATE SESSION TO username;

-- 默认指标所需权限
GRANT SELECT ON V_$INSTANCE TO username;
GRANT SELECT ON V_$SESSION TO username;
GRANT SELECT ON V_$RESOURCE_LIMIT TO username;
GRANT SELECT ON V_$SYSSTAT TO username;
GRANT SELECT ON V_$PROCESS TO username;
GRANT SELECT ON V_$SYSMETRIC TO username;
GRANT SELECT ON V_$WAITCLASSMETRIC TO username;
GRANT SELECT ON V_$SYSTEM_WAIT_CLASS TO username;
GRANT SELECT ON V_$SGA TO username;
GRANT SELECT ON V_$SGASTAT TO username;
GRANT SELECT ON V_$PGASTAT TO username;
GRANT SELECT ON V_$PARAMETER TO username;
GRANT SELECT ON V_$DATAFILE TO username;
GRANT SELECT ON V_$LOG_HISTORY TO username;
GRANT SELECT ON V_$EVENTMETRIC TO username;
GRANT SELECT ON V_$EVENT_NAME TO username;
GRANT SELECT ON V_$LOCKED_OBJECT TO username;
GRANT SELECT ON V_$ASM_DISKGROUP_STAT TO username;
GRANT SELECT ON DBA_TABLESPACE_USAGE_METRICS TO username;
GRANT SELECT ON DBA_TABLESPACES TO username;
GRANT SELECT ON DBA_INDEXES TO username;
GRANT SELECT ON DBA_OBJECTS TO username;
GRANT SELECT ON DBA_USERS TO username;

-- RAC模式额外权限（启用 --isRAC 时）
GRANT SELECT ON GV_$INSTANCE TO username;

-- ASM模式额外权限（启用 --isASM 时）
GRANT SELECT ON V_$ASM_DISK_STAT TO username;
GRANT SELECT ON V_$ASM_ALIAS TO username;
GRANT SELECT ON V_$ASM_DISKGROUP TO username;
GRANT SELECT ON V_$ASM_FILE TO username;

-- Data Guard和归档日志额外权限
GRANT SELECT ON V_$DATAGUARD_STATS TO username;
GRANT SELECT ON V_$DATABASE TO username;
GRANT SELECT ON V_$ARCHIVE_DEST TO username;
```

如果不启用 RAC、ASM、Data Guard 或归档日志采集，可以不执行对应的“额外权限”授权。不要为了简化配置直接授予 `DBA` 或 `SELECT ANY DICTIONARY`。

#### 6. 在监控平台中填写配置

在监控平台创建 Oracle Database 采集配置，并按实际数据库填写以下参数：

| 配置项 | 填写说明 | 示例 |
|---|---|---|
| 数据库主机（`--host`） | Oracle 监听器所在主机的 IP 或可解析域名 | `10.10.10.20` |
| 数据库端口（`--port`） | Oracle 监听端口 | `1521` |
| 用户名（`USER`） | 已完成上述授权的专用监控账号 | `WEOPS_MONITOR` |
| 密码（`PASSWORD`） | 监控账号密码 | 按实际密码填写 |
| 服务名（`SERVICE_NAME`） | 监听器注册的完整服务名，不要填写 SID 或完整连接串 | `ORCLCDB` |
| 查询超时（`--query.timeout`） | 单条 SQL 查询超时时间，建议先使用默认值 | `5` |

根据数据库部署架构选择功能开关：

| 数据库场景 | 需要启用的参数 |
|---|---|
| 单点数据库 | 无额外开关 |
| RAC | `--isRAC` |
| RAC 且使用 ASM | `--isRAC`、`--isASM` |
| Data Guard | `--isDataGuard` |
| 需要采集归档日志指标 | `--isArchiveLog` |

配置时注意：

- 采集节点必须能够访问 Oracle 主机和监听端口；
- `SERVICE_NAME` 必须与监听器注册值一致，不能填写 `host:port/service_name` 形式的完整连接串；
- 监控单个 PDB 时，应填写该 PDB 对应的服务名，并使用该 PDB 内创建的监控账号；
- 不建议使用 `SYS`、`SYSTEM` 或具有 `DBA` 权限的账号进行日常采集；
- 只有实际使用 RAC、ASM、Data Guard 或归档日志采集时，才启用对应开关并授予对应权限。

#### 7. 配置验证

保存并下发配置后，在监控平台检查采集任务状态和采集日志：

- `oracledb_up = 1`：监控账号可以正常连接数据库；
- `oracledb_exporter_last_scrape_error = 0`：最近一次指标采集未发生 SQL 错误；
- 指标持续产生新数据：采集链路和数据上报正常。

如果 `oracledb_up = 0` 或最近一次采集状态异常，应先检查采集日志中的 Oracle 错误码，再按下表处理。

#### 8. 常见问题

| 现象 | 常见原因 | 检查方法 |
|---|---|---|
| `ORA-12514`、`unknown service` | `SERVICE_NAME` 与监听器注册值不一致 | 执行 `lsnrctl status`，并查询 `V$SERVICES.NETWORK_NAME` |
| `ORA-01017` | 用户名、密码、服务名或连接容器错误 | 使用同一账号执行 SQLPlus 连接测试 |
| `ORA-28000` | 监控账号被锁定 | 由 DBA 检查 `DBA_USERS.ACCOUNT_STATUS`，不要改用 SYS 账号绕过 |
| `ORA-00942` | 缺少动态性能视图或 DBA 视图权限 | 根据采集任务日志中的查询对象补充对应 `V_$...` 或 `DBA_...` 授权 |
| `oracledb_up 0` | 数据库不可达或登录失败 | 检查采集任务日志、监听端口、服务名和账号状态 |
| `oracledb_exporter_last_scrape_error 1` | 至少一个采集 SQL 失败或超时 | 检查日志中的 `Context`、Oracle 错误和 `--query.timeout` |
| 新增指标不存在 | 监控平台使用的插件版本尚未包含该指标 | 确认插件已升级到包含该指标的版本，并重新下发采集配置 |

### 指标简介

| **指标ID**                                       | **指标中文名**                | **维度ID**                                                                                    | **维度含义**                                                                 | **单位**    | **指标类型** | **计算指标** | **告警阈值** | **关键指标** |
|------------------------------------------------|--------------------------|---------------------------------------------------------------------------------------------|--------------------------------------------------------------------------|-----------|----------|----------|----------|----------|
| oracledb_up                                    | Oracle数据库监控插件运行状态        | -                                                                                           | -                                                                        | -         | gauge    | 原始指标     | == 0     | 关键指标     |
| oracledb_uptime_seconds                        | Oracle数据库实例已运行时间         | inst_id, instance_name, node_name                                                           | 实例ID, 实例名称, 节点名称                                                         | s         | gauge    | 原始指标     | -        | 关键指标     |
| oracledb_cache_hit_ratio_value                 | Oracle数据库缓存命中率           | cache_hit_type                                                                              | 类型                                                                       | percent   | gauge    | 原始指标     | < 90     | 关键指标     |
| oracledb_activity_execute_count                | Oracle数据库执行次数            | -                                                                                           | -                                                                        | -         | counter  | 原始指标     | -        |          |
| oracledb_activity_execute_rate                 | Oracle数据库执行速率            | -                                                                                           | -                                                                        | cps       | gauge    | 衍生指标     | -        |          |
| oracledb_activity_parse_count_total            | Oracle数据库解析次数            | -                                                                                           | -                                                                        | -         | counter  | 原始指标     | -        |          |
| oracledb_activity_parse_rate                   | Oracle数据库解析速率            | -                                                                                           | -                                                                        | cps       | gauge    | 衍生指标     | -        |          |
| oracledb_activity_user_commits                 | Oracle数据库用户提交次数          | -                                                                                           | -                                                                        | -         | counter  | 原始指标     | -        |          |
| oracledb_activity_user_commits_rate            | Oracle数据库用户提交速率          | -                                                                                           | -                                                                        | cps       | gauge    | 衍生指标     | -        |          |
| oracledb_activity_user_rollbacks               | Oracle数据库用户回滚次数          | -                                                                                           | -                                                                        | -         | counter  | 原始指标     | -        |          |
| oracledb_activity_user_rollbacks_rate          | Oracle数据库用户回滚速率          | -                                                                                           | -                                                                        | cps       | gauge    | 衍生指标     | -        |          |
| oracledb_wait_time_application                 | Oracle数据库应用类等待时间         | -                                                                                           | -                                                                        | ms        | gauge    | 原始指标     | -        |          |
| oracledb_wait_time_commit                      | Oracle数据库提交等待时间          | -                                                                                           | -                                                                        | ms        | gauge    | 原始指标     | -        |          |
| oracledb_wait_time_concurrency                 | Oracle数据库并发等待时间          | -                                                                                           | -                                                                        | ms        | gauge    | 原始指标     | -        |          |
| oracledb_wait_time_configuration               | Oracle数据库配置等待时间          | -                                                                                           | -                                                                        | ms        | gauge    | 原始指标     | -        |          |
| oracledb_wait_time_network                     | Oracle数据库网络等待时间          | -                                                                                           | -                                                                        | ms        | gauge    | 原始指标     | -        |          |
| oracledb_wait_time_other                       | Oracle数据库其他等待时间          | -                                                                                           | -                                                                        | ms        | gauge    | 原始指标     | -        |          |
| oracledb_wait_time_scheduler                   | Oracle数据库调度程序等待时间        | -                                                                                           | -                                                                        | ms        | gauge    | 原始指标     | -        |          |
| oracledb_wait_time_system_io                   | Oracle数据库系统I/O等待时间       | -                                                                                           | -                                                                        | ms        | gauge    | 原始指标     | -        |          |
| oracledb_wait_time_user_io                     | Oracle数据库用户I/O等待时间       | -                                                                                           | -                                                                        | ms        | gauge    | 原始指标     | -        |          |
| oracledb_resource_current_utilization          | Oracle数据库当前资源使用量         | resource_name                                                                               | 资源类型                                                                     | -         | gauge    | 原始指标     | -        |          |
| oracledb_resource_limit_value                  | Oracle数据库资源限定值           | resource_name                                                                               | 资源类型                                                                     | -         | gauge    | 原始指标     | -        |          |
| oracledb_process_count                         | Oracle数据库进程数             | -                                                                                           | -                                                                        | -         | gauge    | 原始指标     | -        |          |
| oracledb_sessions_value                        | Oracle数据库会话数             | status, type                                                                                | 会话状态, 会话类型                                                               | -         | gauge    | 原始指标     | -        | 关键指标     |
| oracledb_session_lock_seconds_in_wait          | Oracle数据库会话锁等待时间         | username, status, client_host, client_app, sid, serial_num, object_id                       | 用户名, 会话状态, 客户端主机, 客户端程序, 会话 SID, 会话序列号, 被锁对象 ID                          | s         | gauge    | 原始指标     | -        | 关键指标     |
| oracledb_session_blocked_seconds_in_wait       | Oracle数据库会话被阻塞等待时间       | username, status, client_host, client_app, sid, serial_num, sql_id, blocking_session, event | 用户名, 会话状态, 客户端主机, 客户端程序, 会话 SID, 会话序列号, SQL语句ID, 阻塞该会话的上游会话ID, 会话等待的事件名称 | s         | gauge    | 原始指标     | -        | 关键指标     |
| oracledb_datafile_status                       | Oracle数据库数据文件状态          | file_id, file_name, status_name                                                            | 数据文件ID, 数据文件完整路径, 数据文件原始状态                                           | -         | gauge    | 原始指标     | == 0     | 关键指标     |
| oracledb_redo_log_switches_1h                  | Oracle数据库最近1小时重做日志切换次数 | -                                                                                           | -                                                                        | 次         | gauge    | 原始指标     | -        |          |
| oracledb_log_file_sync_waiting_sessions        | Oracle数据库日志文件同步等待会话数    | -                                                                                           | -                                                                        | 个         | gauge    | 原始指标     | -        |          |
| oracledb_log_file_sync_wait_count              | Oracle数据库日志文件同步等待次数     | -                                                                                           | -                                                                        | 次         | gauge    | 原始指标     | -        |          |
| oracledb_log_file_sync_time_waited_ms          | Oracle数据库日志文件同步累计等待时间   | -                                                                                           | -                                                                        | ms        | gauge    | 原始指标     | -        |          |
| oracledb_log_file_sync_avg_wait_ms             | Oracle数据库日志文件同步平均等待时间   | -                                                                                           | -                                                                        | ms        | gauge    | 原始指标     | -        |          |
| oracledb_deadlocks_total                       | Oracle数据库实例启动后累计死锁次数    | -                                                                                           | -                                                                        | 次         | counter  | 原始指标     | -        | 关键指标     |
| oracledb_lock_waiting_sessions_count           | Oracle数据库当前锁等待会话数        | -                                                                                           | -                                                                        | 个         | gauge    | 原始指标     | -        | 关键指标     |
| oracledb_db_system_value                       | Oracle数据库系统资源            | resource_name                                                                               | 资源名称                                                                     | -         | gauge    | 原始指标     | -        |          |
| oracledb_session_usage_used_percent            | Oracle数据库当前会话限制使用率       | -                                                                                           | -                                                                        | percent   | gauge    | 原始指标     | -        | 关键指标     |
| oracledb_sga_total                             | Oracle数据库SGA总大小          | -                                                                                           | -                                                                        | bytes     | gauge    | 原始指标     | -        |          |
| oracledb_sga_free                              | Oracle数据库SGA可用大小         | -                                                                                           | -                                                                        | bytes     | gauge    | 原始指标     | -        |          |
| oracledb_sga_used_percent                      | Oracle数据库SGA使用率          | -                                                                                           | -                                                                        | percent   | gauge    | 原始指标     | > 80     |          |
| oracledb_pga_total                             | Oracle数据库PGA总大小          | -                                                                                           | -                                                                        | bytes     | gauge    | 原始指标     | -        |          |
| oracledb_pga_used                              | Oracle数据库PGA已使用大小        | -                                                                                           | -                                                                        | bytes     | gauge    | 原始指标     | -        |          |
| oracledb_pga_used_percent                      | Oracle数据库PGA使用率          | -                                                                                           | -                                                                        | percent   | gauge    | 原始指标     | > 70     |          |
| oracledb_tablespace_bytes                      | Oracle数据库表已使用容量大小        | tablespace, type                                                                            | 表空间名称, 表空间类型                                                             | bytes     | gauge    | 原始指标     | -        |          |
| oracledb_tablespace_max_bytes                  | Oracle数据库表最大容量限制         | tablespace, type                                                                            | 表空间名称, 表空间类型                                                             | bytes     | gauge    | 原始指标     | -        |          |
| oracledb_tablespace_free                       | Oracle数据库表可用容量大小         | tablespace, type                                                                            | 表空间名称, 表空间类型                                                             | bytes     | gauge    | 原始指标     | -        |          |
| oracledb_tablespace_used_percent               | Oracle数据库表空间使用率          | tablespace, type                                                                            | 表空间名称, 表空间类型                                                             | percent   | gauge    | 原始指标     | > 80     | 关键指标     |
| oracledb_rac_node                              | Oracle数据库RAC节点数量         | -                                                                                           | -                                                                        | -         | gauge    | 原始指标     | -        | 关键指标     |
| oracledb_dataguard_transport_lag_delay         | Oracle数据库DataGuard数据传输延迟 | -                                                                                           | -                                                                        | s         | gauge    | 原始指标     | -        | 关键指标     |
| oracledb_dataguard_apply_lag_delay             | Oracle数据库DataGuard数据应用延迟 | -                                                                                           | -                                                                        | s         | gauge    | 原始指标     | -        | 关键指标     |
| oracledb_asm_diskgroup_free                    | Oracle数据库ASM磁盘组可用空间      | diskgroup_name                                                                              | 磁盘组名称                                                                    | bytes     | gauge    | 原始指标     | -        |          |
| oracledb_asm_diskgroup_total                   | Oracle数据库ASM磁盘组总容量       | diskgroup_name                                                                              | 磁盘组名称                                                                    | bytes     | gauge    | 原始指标     | -        |          |
| oracledb_asm_diskgroup_usage                   | Oracle数据库ASM磁盘组空间使用率     | diskgroup_name                                                                              | 磁盘组名称                                                                    | percent   | gauge    | 原始指标     | > 80     |          |
| oracledb_asm_disk_stat_reads                   | Oracle数据库ASM磁盘的读操作总数     | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path             | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径                               | -         | counter  | 原始指标     | -        |          |
| oracledb_asm_disk_stat_reads_rate              | Oracle数据库ASM磁盘的读操作速率     | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path             | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径                               | cps       | gauge    | 衍生指标     | -        |          |
| oracledb_asm_disk_stat_writes                  | Oracle数据库ASM磁盘的写操作总数     | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path             | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径                               | -         | counter  | 原始指标     | -        |          |
| oracledb_asm_disk_stat_writes_rate             | Oracle数据库ASM磁盘的写操作速率     | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path             | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径                               | cps       | gauge    | 衍生指标     | -        |          |
| oracledb_asm_disk_stat_bytes_read              | Oracle数据库ASM磁盘的总读取字节数    | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path             | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径                               | bytes     | counter  | 原始指标     | -        |          |
| oracledb_asm_disk_stat_bytes_read_rate         | Oracle数据库ASM磁盘的读取传输速率    | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path             | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径                               | Bps       | gauge    | 衍生指标     | -        |          |
| oracledb_asm_disk_stat_read_time               | Oracle数据库ASM磁盘的读取时间总和    | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path             | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径                               | ms        | counter  | 原始指标     | -        |          |
| oracledb_asm_disk_stat_read_time_increase      | Oracle数据库ASM磁盘的读取时间      | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path             | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径                               | ms        | gauge    | 衍生指标     | -        |          |
| oracledb_asm_disk_stat_write_time              | Oracle数据库ASM磁盘的写入时间总和    | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path             | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径                               | ms        | counter  | 原始指标     | -        |          |
| oracledb_asm_disk_stat_write_time_increase     | Oracle数据库ASM磁盘的写入时间      | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path             | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径                               | ms        | gauge    | 衍生指标     | -        |          |
| oracledb_asm_disk_stat_bytes_written           | Oracle数据库ASM磁盘的总写入字节数    | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path             | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径                               | bytes     | counter  | 原始指标     | -        |          |
| oracledb_asm_disk_stat_bytes_written_rate      | Oracle数据库ASM磁盘的写入传输速率    | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path             | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径                               | Bps       | gauge    | 衍生指标     | -        |          |
| oracledb_asm_disk_stat_io                      | Oracle数据库ASM磁盘总IO        | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path             | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径                               | -         | counter  | 原始指标     | -        |          |
| oracledb_asm_disk_stat_iops                    | Oracle数据库ASM磁盘每秒IO       | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path             | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径                               | cps       | gauge    | 衍生指标     | -        |          |
| oracledb_asm_space_consumers_files             | Oracle数据库ASM磁盘组上文件数量     | diskgroup_name, file_type, inst_id, instance_name, node_name                                | 磁盘组名称, 文件类型, 实例ID, 实例名称, 节点名称                                            | -         | gauge    | 原始指标     | -        |          |
| oracledb_asm_space_consumers_size_mb           | Oracle数据库ASM磁盘组上文件大小     | diskgroup_name, file_type, inst_id, instance_name, node_name                                | 磁盘组名称, 文件类型, 实例ID, 实例名称, 节点名称                                            | mebibytes | gauge    | 原始指标     | -        |          |
| oracledb_archived_log_total                    | Oracle数据库归档日志总空间大小       | diskgroup_name                                                                              | 磁盘组名称                                                                    | bytes     | gauge    | 原始指标     | -        |          |
| oracledb_archived_log_used                     | Oracle数据库归档日志已使用空间大小     | diskgroup_name                                                                              | 磁盘组名称                                                                    | bytes     | gauge    | 原始指标     | -        |          |
| oracledb_archived_log_usage_ratio              | Oracle数据库归档日志空间使用率       | diskgroup_name                                                                              | 磁盘组名称                                                                    | percent   | gauge    | 原始指标     | > 80     |          |
| oracledb_unusable_index_count                  | Oracle数据库中不可用的索引数量       | -                                                                                           | -                                                                        | -         | gauge    | 原始指标     |          |          |
| oracledb_invalid_objects_count                 | Oracle数据库中无效对象数量         | owner, object_type, status                                                                  | 拥有者, 对象类型, 对象状态                                                          | -         | gauge    | 原始指标     |          |          |
| process_cpu_seconds_total                      | Oracle数据库监控探针进程CPU秒数总计   | -                                                                                           | -                                                                        | s         | gauge    | 原始指标     | -        |          |
| process_max_fds                                | Oracle数据库监控探针进程最大文件描述符数  | -                                                                                           | -                                                                        | -         | gauge    | 原始指标     | -        |          |
| process_open_fds                               | Oracle数据库监控探针进程打开文件描述符数  | -                                                                                           | -                                                                        | -         | gauge    | 原始指标     | -        |          |
| process_resident_memory_bytes                  | Oracle数据库监控探针进程常驻内存大小    | -                                                                                           | -                                                                        | bytes     | gauge    | 原始指标     | -        |          |
| process_virtual_memory_bytes                   | Oracle数据库监控探针进程虚拟内存大小    | -                                                                                           | -                                                                        | bytes     | gauge    | 原始指标     | -        |          |
| oracledb_exporter_last_scrape_duration_seconds | Oracle数据库监控探针最近一次抓取时长    | -                                                                                           | -                                                                        | s         | gauge    | 原始指标     | -        | 关键指标     |
| oracledb_exporter_last_scrape_error            | Oracle数据库监控探针最近一次抓取状态    | -                                                                                           | -                                                                        | -         | gauge    | 原始指标     | -        |          |
| oracledb_exporter_scrape_errors_total          | Oracle数据库监控探针采集错误总数      | collector                                                                                   | 采集器                                                                      | -         | gauge    | 原始指标     | -        | 关键指标     |
| oracledb_exporter_scrapes_total                | Oracle数据库监控探针抓取指标总数      | -                                                                                           | -                                                                        | -         | gauge    | 原始指标     | -        |          |

### 版本日志

#### weops_oracledb_exporter 2.2.0

- weops调整

#### weops_oracledb_exporter 2.2.1

- 增加dataguard、归档日志类监控指标
- 增加rac、asm和dataguard指标采集开关
- 去除自定义文件

#### weops_oracledb_exporter 2.2.2

- DSN拆分
- 隐藏敏感参数
- process类监控指标中文名更正

#### weops_oracledb_exporter 2.2.3

- up指标中文名优化

#### weops_oracledb_exporter 3.1.1

- 合入官方版本v1.5.4
- 指标单位更正
  oracledb_archived_log_usage_ratio   Oracle数据库归档日志空间使用率    percent
- 新增部分指标
- 内置衍生指标

#### weops_oracledb_exporter 3.1.2
- 合入官方版本v1.5.5
- 新增指标
  oracledb_unusable_index_count	    Oracle数据库中不可用的索引数量
  oracledb_invalid_objects_count	Oracle数据库中无效对象数量

#### weops_oracledb_exporter 3.1.3
- 新增指标
  oracledb_session_lock_seconds_in_wait	Oracle数据库会话锁等待时间

#### weops_oracledb_exporter 3.1.4
- 新增指标
  oracledb_session_usage_used_percent	Oracle数据库当前会话限制使用率
  oracledb_session_blocked_seconds_in_wait	Oracle数据库会话被阻塞等待时间

#### weops_oracledb_exporter 3.1.5
- 新增指标
  oracledb_datafile_status	Oracle数据库数据文件状态
  oracledb_redo_log_switches_1h	Oracle数据库最近1小时重做日志切换次数
  oracledb_log_file_sync_waiting_sessions	Oracle数据库日志文件同步等待会话数
  oracledb_log_file_sync_wait_count	Oracle数据库日志文件同步等待次数
  oracledb_log_file_sync_time_waited_ms	Oracle数据库日志文件同步累计等待时间
  oracledb_log_file_sync_avg_wait_ms	Oracle数据库日志文件同步平均等待时间
  oracledb_deadlocks_total	Oracle数据库实例启动后累计死锁次数
  oracledb_lock_waiting_sessions_count	Oracle数据库当前锁等待会话数
