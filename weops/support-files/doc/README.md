## 嘉为蓝鲸oracledb插件使用说明

## 使用说明

### 插件功能

采集器连接oracle数据库，执行SQL查询语句，将结果解析到prometheus数据格式的监控指标。
实际收集的指标取决于数据库的配置和版本。

### 版本支持

操作系统支持: linux, windows

是否支持arm: 支持

**组件支持版本：**

Oracle Database: `11g`, `12c`, `18c`, `19c`, `21c`

部署模式支持: `standalone(单点)`, `RAC(集群)`, `dataGuard(DG)`

**是否支持远程采集:**

是

### 参数说明


| **参数名**              | **含义**                    | **是否必填** | **使用举例**       |
|----------------------|---------------------------|----------|----------------|
| --host               | 数据库主机IP                   | 是        | 127.0.0.1      |
| --port               | 数据库服务端口                   | 是        | 1521           |
| USER                 | 数据库用户名(环境变量)              | 是        |                |
| PASSWORD             | 数据库密码(环境变量)               | 是        |                |
| SERVICE_NAME         | 数据库服务名(环境变量)              | 是        | ORCLCDB        |
| --isRAC              | 是否为rac集群架构(开关参数), 默认不开启   | 否        |                |
| --isASM              | 是否有ASM磁盘组(开关参数), 默认不开启    | 否        |                |
| --isDataGuard        | 是否为DataGuard(开关参数), 默认不开启 | 否        |                |
| --isArchiveLog       | 是否采集归档日志指标, 默认不开启         | 否        |                |
| --query.timeout      | 查询超时秒数，默认使用5s             | 否        | 5              |
| --log.level          | 日志级别                      | 否        | info           |
| --web.listen-address | exporter监听id及端口地址         | 否        | 127.0.0.1:9601 |

### 使用指引

1. 查看Oracle数据库服务名和域名注意！**对于oracle数据库12版本，DSN中数据库名后必须加入域名，其他版本一般不需要**ORCLCDB是Oracle数据库的一个服务名称（Service Name），它用于唯一标识数据库实例中的一个服务。例: "oracle://system:Weops123!@db12c-oracle-db.oracle:1521/ORCLCDB.localdomain"

   - 查看当前数据库实例的 `SERVICE_NAME` 参数的值。

     ```sql
     SELECT value FROM v$parameter WHERE name = 'service_names';
     ```
   - 查看当前数据库实例的 `DB_DOMAIN` 参数的值。如果返回结果为空，表示未设置特定的域名。

     ```sql
     SELECT value FROM v$parameter WHERE name = 'db_domain';
     ```
2. 若出现unknown service error

   - 需检查监听器的当前状态，确保监听器正在运行并监听正确的端口，运行命令 `lsnrctl status`。
   - 确认监听器配置文件（`lsnrctl status`会输出监听器配置状态等信息，寻找配置文件，通常是 listener.ora）中是否正确定义了服务名称，并与您尝试连接的服务名称匹配。
   - `lsnrctl` 在oracle数据库12版本中，此命令一般存放于 `/u01/app/oracle/product/12.2.0/dbhome_1/` ； 在oracle数据库19版本中，一般存放于 `/opt/oracle/product/19c/dbhome_1/bin`
3. 连接Oracle数据库
   使用操作系统的身份认证（通常是超级用户或管理员），直接以 sysdba 角色登录到数据库

   ```shell
   sqlplus / as sysdba
   ```

   使用指定账户登录

   ```shell
   sqlplus username/password@host:port/service_name
   ```
4. 创建账户及授权  
   注意！创建账户时必须使用管理员账户
   
   > 创建账户类型有区别:  
   a) 在Oracle数据库中，使用C##前缀是为了创建一个包含大写字母和特殊字符的用户名，这样可以确保在创建和使用这些用户时不会发生命名冲突。C##前缀表示"Container Database"，用于标识这个用户是一个全局共享的用户，而不是只属于某个具体的Pluggable Database (PDB)。  
   b) 要决定是否在用户名前使用C##，主要取决于数据库的架构。在Oracle 12c及更高版本中，数据库被分为一个容器数据库（CDB）和一个或多个可插拔数据库（PDB）。如果你在CDB层面创建用户，可以选择使用C##前缀，表示这个用户是一个全局共享的用户。如果在PDB层面创建用户，通常不需要使用C##前缀，因为PDB内的用户空间是相互隔离的。  
   c) 在创建用户时是否使用C##前缀取决于你的特定需求和数据库架构。如果你的用户需要在不同的PDB之间共享，并且你希望避免命名冲突，那么可以考虑使用C##前缀。如果用户只在特定的PDB中使用，可能不需要这个前缀。  
   d) 使用 C## 前缀的情况：  
   `CREATE USER C##GlobalUser IDENTIFIED BY password CONTAINER = ALL;`  
   e) 不使用 C## 前缀的情况：   
   `CREATE USER LocalUser IDENTIFIED BY password;`


   ```sql
   # 新建用户
   CREATE USER [user] IDENTIFIED BY [password];

   # 修改用户的密码，密码若含特殊字符需使用双引号将密码括起来
   ALTER USER [user] IDENTIFIED BY [password];

   # 允许用户建立数据库会话
   GRANT CREATE SESSION TO [user];

   # uptime指标授权
   GRANT SELECT ON V_$instance to [user];

   # rac指标授权
   GRANT SELECT ON GV_$instance to [user];

   # sessions类指标授权
   GRANT SELECT ON V_$session to [user];

   # resource类指标授权
   GRANT SELECT ON V_$resource_limit to [user];

   # asm_diskgroup类指标授权
   GRANT SELECT ON V_$datafile to [user];
   GRANT SELECT ON V_$asm_diskgroup_stat to [user];

   # activity类指标授权
   GRANT SELECT ON V_$sysstat to [user];

   # process类指标授权
   GRANT SELECT ON V_$process to [user];

   # wait_time类指标授权
   GRANT SELECT ON V_$waitclassmetric to [user];
   GRANT SELECT ON V_$system_wait_class to [user];

   # tablespace类指标授权
   GRANT SELECT ON dba_tablespace_usage_metrics to [user];
   GRANT SELECT ON dba_tablespaces to [user];

   # asm_disk_stat类指标授权
   GRANT SELECT ON V_$asm_disk_stat to [user];
   GRANT SELECT ON V_$asm_diskgroup_stat to [user];
   GRANT SELECT ON V_$instance to [user];

   # asm_space_consumers类指标授权
   GRANT SELECT ON V_$asm_alias to [user];
   GRANT SELECT ON V_$asm_diskgroup to [user];
   GRANT SELECT ON V_$asm_file to [user];

   # sga类指标授权
   GRANT SELECT ON V_$sga TO [user];
   GRANT SELECT ON V_$sgastat TO [user];

   # pga类指标授权
   GRANT SELECT ON V_$pgastat TO [user];

   # dataguard类指标授权
   GRANT SELECT ON V_$dataguard_stats TO [user];
   
   # archived_log类指标授权
   GRANT SELECT ON V_$database to [user];
   GRANT SELECT ON V_$archive_dest to [user];
   GRANT SELECT ON V_$parameter to [user];
   GRANT SELECT ON V_$asm_diskgroup to [user];
   ```

### 指标简介

| **指标ID**                               | **指标中文名**                | **维度ID**                                                                        | **维度含义**                                   | **单位**    |
|----------------------------------------|--------------------------|---------------------------------------------------------------------------------|--------------------------------------------|-----------|
| oracledb_up                            | Oracle数据库监控插件运行状态        | -                                                                               | -                                          | -         |
| oracledb_uptime_seconds                | Oracle数据库实例已运行时间         | inst_id, instance_name, node_name                                               | 实例ID, 实例名称, 节点名称                           | s         |
| oracledb_activity_execute_count        | Oracle数据库执行次数            | -                                                                               | -                                          | -         |
| oracledb_activity_parse_count_total    | Oracle数据库解析次数            | -                                                                               | -                                          | -         |
| oracledb_activity_user_commits         | Oracle数据库用户提交次数          | -                                                                               | -                                          | -         |
| oracledb_activity_user_rollbacks       | Oracle数据库用户回滚次数          | -                                                                               | -                                          | -         |
| oracledb_wait_time_application         | Oracle数据库应用类等待时间         | -                                                                               | -                                          | ms        |
| oracledb_wait_time_commit              | Oracle数据库提交等待时间          | -                                                                               | -                                          | ms        |
| oracledb_wait_time_concurrency         | Oracle数据库并发等待时间          | -                                                                               | -                                          | ms        |
| oracledb_wait_time_configuration       | Oracle数据库配置等待时间          | -                                                                               | -                                          | ms        |
| oracledb_wait_time_network             | Oracle数据库网络等待时间          | -                                                                               | -                                          | ms        |
| oracledb_wait_time_other               | Oracle数据库其他等待时间          | -                                                                               | -                                          | ms        |
| oracledb_wait_time_scheduler           | Oracle数据库调度程序等待时间        | -                                                                               | -                                          | ms        |
| oracledb_wait_time_system_io           | Oracle数据库系统I/O等待时间       | -                                                                               | -                                          | ms        |
| oracledb_wait_time_user_io             | Oracle数据库用户I/O等待时间       | -                                                                               | -                                          | ms        |
| oracledb_resource_current_utilization  | Oracle数据库当前资源使用量         | resource_name                                                                   | 资源类型                                       | -         |
| oracledb_resource_limit_value          | Oracle数据库资源限定值           | resource_name                                                                   | 资源类型                                       | -         |
| oracledb_process_count                 | Oracle数据库进程数             | -                                                                               | -                                          | -         |
| oracledb_sessions_value                | Oracle数据库会话数             | status, type                                                                    | 会话状态, 会话类型                                 | -         |
| oracledb_sga_total                     | Oracle数据库SGA总大小          | -                                                                               | -                                          | bytes     |
| oracledb_sga_free                      | Oracle数据库SGA可用大小         | -                                                                               | -                                          | bytes     |
| oracledb_sga_used_percent              | Oracle数据库SGA使用率          | -                                                                               | -                                          | percent   |
| oracledb_pga_total                     | Oracle数据库PGA总大小          | -                                                                               | -                                          | bytes     |
| oracledb_pga_used                      | Oracle数据库PGA已使用大小        | -                                                                               | -                                          | bytes     |
| oracledb_pga_used_percent              | Oracle数据库PGA使用率          | -                                                                               | -                                          | percent   |
| oracledb_tablespace_bytes              | Oracle数据库表已使用容量大小        | tablespace, type                                                                | 表空间名称，表空间类型                                | bytes     |
| oracledb_tablespace_max_bytes          | Oracle数据库表最大容量限制         | tablespace, type                                                                | 表空间名称，表空间类型                                | bytes     |
| oracledb_tablespace_free               | Oracle数据库表可用容量大小         | tablespace, type                                                                | 表空间名称，表空间类型                                | bytes     |
| oracledb_tablespace_used_percent       | Oracle数据库表空间使用率          | tablespace, type                                                                | 表空间名称，表空间类型                                | percent   |
| oracledb_rac_node                      | Oracle数据库RAC节点数量         | -                                                                               | -                                          | -         |
| oracledb_dataguard_transport_lag_delay | Oracle数据库DataGuard数据传输延迟 | -                                                                               | -                                          | -         |
| oracledb_dataguard_apply_lag_delay     | Oracle数据库DataGuard数据应用延迟 | -                                                                               | -                                          | -         |
| oracledb_asm_diskgroup_free            | Oracle数据库ASM磁盘组可用空间      | diskgroup_name                                                                  | 磁盘组名称                                      | bytes     |
| oracledb_asm_diskgroup_total           | Oracle数据库ASM磁盘组总容量       | diskgroup_name                                                                  | 磁盘组名称                                      | bytes     |
| oracledb_asm_diskgroup_usage           | Oracle数据库ASM磁盘组空间使用率     | diskgroup_name                                                                  | 磁盘组名称                                      | percent   |
| oracledb_asm_disk_stat_reads           | Oracle数据库ASM磁盘的读操作总数     | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径 | -         |
| oracledb_asm_disk_stat_writes          | Oracle数据库ASM磁盘的写操作总数     | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径 | -         |
| oracledb_asm_disk_stat_bytes_read      | Oracle数据库ASM磁盘的总读取字节数    | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径 | bytes     |
| oracledb_asm_disk_stat_read_time       | Oracle数据库ASM磁盘的读取时间总和    | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径 | ms        |
| oracledb_asm_disk_stat_write_time      | Oracle数据库ASM磁盘的写入时间总和    | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径 | ms        |
| oracledb_asm_disk_stat_bytes_written   | Oracle数据库ASM磁盘的总写入字节数    | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径 | bytes     |
| oracledb_asm_disk_stat_iops            | Oracle数据库ASM磁盘每秒IO       | inst_id, node_name, instance_name, diskgroup_name, disk_number, failgroup, path | 实例ID, 节点名称, 实例名称, 磁盘组名称, 磁盘编号, 故障组名称, 磁盘路径 | -         |
| oracledb_asm_space_consumers_files     | Oracle数据库ASM磁盘组上文件数量     | diskgroup_name, file_type, inst_id, instance_name, node_name                    | 磁盘组名称, 文件类型, 实例ID, 实例名称, 节点名称              | -         |
| oracledb_asm_space_consumers_size_mb   | Oracle数据库ASM磁盘组上文件大小     | diskgroup_name, file_type, inst_id, instance_name, node_name                    | 磁盘组名称, 文件类型, 实例ID, 实例名称, 节点名称              | mebibytes |
| oracledb_archived_log_total            | Oracle数据库归档日志总空间大小       | diskgroup_name                                                                  | 磁盘组名称                                      | bytes     |
| oracledb_archived_log_used             | Oracle数据库归档日志已使用空间大小     | diskgroup_name                                                                  | 磁盘组名称                                      | bytes     |
| oracledb_archived_log_usage_ratio      | Oracle数据库归档日志空间使用率       | diskgroup_name                                                                  | 磁盘组名称                                      | bytes     |

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

添加“小嘉”微信即可获取oracle数据库监控指标最佳实践礼包，其他更多问题欢迎咨询

<img src="https://wedoc.canway.net/imgs/img/小嘉.jpg" width="50%" height="50%">
