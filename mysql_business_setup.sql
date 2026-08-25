-- CDN 注入播控：内容提供商 / 内容服务商
-- 依赖：先执行 mysql_login_setup.sql 创建库 kratos
--
-- 职责划分：
--   cp     内容提供商。版权/制作方，向 CDN 注入媒资（C2/HTTP/FTP）。
--   sp     内容服务商。IPTV/OTT 等分发平台，从 CDN 拉流给终端。
--   cp_sp  注入路由。一份 CP 内容可下发给多个 SP，一个 SP 也可聚合多个 CP。

USE `kratos`;

-- ---------------------------------------------------------------------------
-- 内容提供商（CP）
-- 协议侧 CPID 用 cp_code；注入鉴权、内容命名空间、配额都挂在本表。
-- ---------------------------------------------------------------------------
CREATE TABLE `cp` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '内容提供商ID',
  `cp_code` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'CP编码（协议侧 CPID，全局唯一）',
  `cp_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'CP全称',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '状态：0-禁用 1-正常 2-冻结',
  `create_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_cp_code` (`cp_code`),
  KEY `idx_status` (`status`),
  KEY `idx_create_date` (`create_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容提供商表';

-- ---------------------------------------------------------------------------
-- 内容服务商（SP）
-- 协议侧 SPID/CSPID 用 sp_code；播放域名、鉴权、覆盖区域挂在本表。
-- ---------------------------------------------------------------------------
CREATE TABLE `sp` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '内容服务商ID',
  `sp_code` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'SP编码（协议侧 SPID/CSPID，全局唯一）',
  `sp_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'SP全称',
  `sp_config` json DEFAULT NULL COMMENT 'SP配置',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '状态：0-禁用 1-正常 2-冻结',
  `create_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sp_code` (`sp_code`),
  KEY `idx_status` (`status`),
  KEY `idx_sp_type` (`sp_type`),
  KEY `idx_create_date` (`create_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容服务商表';

-- ---------------------------------------------------------------------------
-- CP-SP 绑定
-- 注入任务按此路由：CP 注入的内容只下发给已绑定且生效的 SP。
-- ---------------------------------------------------------------------------
CREATE TABLE `cp_sp` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '绑定ID',
  `cp_id` int(10) unsigned NOT NULL COMMENT '内容提供商ID',
  `sp_id` int(10) unsigned NOT NULL COMMENT '内容服务商ID',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '状态：0-禁用 1-正常',
  `create_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_cp_sp` (`cp_id`,`sp_id`),
  KEY `idx_sp_id` (`sp_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容提供商与内容服务商绑定表';
