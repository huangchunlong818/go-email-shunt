
## 邮件域预热功能公共代码
参考文档(php的)
https://alidocs.dingtalk.com/i/nodes/XPwkYGxZV32YGrg7cYobvzvEWAgozOKL

#### 环境要求
本功能在以下环境下开发和测试：

● Go 1.18 以上

● MySQL

● Redis

#### 特点:
● 支持mailgun和sendgrid两种邮件发送客户端

● 支持 1-redis限流 2-平滑加权轮询 两种域名选择算法模型. 邮件发送频率很低建议使用算法模型2.

#### 详细说明
##### 数据表使用说明
发送域配置表
```sql
CREATE TABLE `email_domains` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `service_provider` tinyint unsigned NOT NULL COMMENT '服务商; 1: Mailgun, 2: Sendgrid',
    `domain` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '发件域',
    `api_key_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'API KEY-ID',
    `api_key` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'API 密钥',
    `priority` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '优先级, 越小越优先',
    `weight` tinyint unsigned DEFAULT '1' COMMENT '轮训算法权重',
    `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态; 0: 禁用, 1: 启用',
    `created_at` int unsigned NOT NULL DEFAULT '0',
    `updated_at` int unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

示例数据
● 新域名的优先级高，优先级为 0
```sql
INSERT INTO `email_domains` (`id`, `service_provider`, `domain`, `api_key_id`, `api_key`, `priority`, `weight`, `status`, `created_at`, `updated_at`) VALUES (1, 1, 'mail.parcelpanel.net', 'xxxx', 'xxxx', 0, 1, 1, 1716430043, 0);
```

● 旧域名的优先级为最低的，往大了设置，这里设置为 99
```sql
INSERT INTO `email_domains` (`id`, `service_provider`, `domain`, `api_key_id`, `api_key`, `priority`, `status`, `created_at`, `updated_at`) VALUES (13, 1, 'edms.parcelpanel.net', 'xxxx', 'xxxx', 99, 1, 1, 1716437113, 0);
```
发送规则表
```sql
CREATE TABLE `email_send_rules` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `domain_id` int unsigned NOT NULL DEFAULT '0',
    `started_at` int unsigned NOT NULL DEFAULT '0',
    `ended_at` int unsigned NOT NULL DEFAULT '0',
    `total_send_count` int unsigned NOT NULL DEFAULT '0' COMMENT '预计总发送量',
    `sent_count` int unsigned NOT NULL DEFAULT '0' COMMENT '已发数量',
    `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '启用状态; 0: 禁用, 1: 启用',
    `created_at` int unsigned NOT NULL DEFAULT '0',
    `updated_at` int unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```sql
示例数据
● 将旧发送域作为 fallback 发送域，开始结束时间分别设置为字段的最大最小值，started_at 设置为 0 表示一直可以发送，不会进入发送量控制的逻辑
```sql
INSERT INTO `email_send_rules` (`domain_id`, `started_at`, `ended_at`, `total_send_count`, `sent_count`, `status`, `created_at`, `updated_at`) VALUES (13, 0, 4294967295, 0, 4022820, 1, 1716858270, 0);
```
● 给 ID 为 1 的发送域新增一条 2024-05-28 10:00 - 2024-05-29 9:59 时间段的发送规则，预计发送 50 封，则每封发送间隔至少会间隔 1728s
```sql
INSERT INTO `email_send_rules` (`domain_id`, `started_at`, `ended_at`, `total_send_count`, `sent_count`, `status`, `created_at`, `updated_at`) VALUES (1, 1716861600, 1716947999, 50, 0, 1, 1716858595, 0);
```

##### 发件流程描述
1. 选择发送域
2. 获取邮件配置信息
3. 创建 Sender 实例 
4. 发送邮件
5. 记录日志
6. 标记发送规则为已完成

##### 发送示例
main.go中是使用示例
```go
//依次传入redis客户端, mysql连接客户端, 域名选择算法模型(1-redis限流 2-平滑加权轮询)
emailSenderSrv := sender.NewSenderManager(redisCli, db, 2)
send, err := emailSenderSrv.Send(content)
```
