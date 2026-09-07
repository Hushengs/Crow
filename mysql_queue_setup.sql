-- CDN 注入队列：任务 / 内容快照 / 下游日志
-- 依赖：先执行 mysql_system_setup.sql 创建库 kratos
-- 关联：mysql_vod_setup.sql（video / episode / media）、mysql_business_setup.sql（sp）
--
-- 职责划分：
--   inject_task     注入任务。影片、节目、媒体的新增/修改/删除各生成一条任务。
--   inject_content  注入内容。同一影片、节目、媒体仅一条；只保留最后一次媒资操作。
--   inject_log      注入日志。记录下发下游 CDN 的同步应答与异步回调。

USE `kratos`;

-- ---------------------------------------------------------------------------
-- 注入任务（Inject Task）
-- 工作队列：媒资每次新增、修改、删除都追加一条任务，由调度消费。
-- 状态仅表示排队进度：待注入、注入中。终态落在 inject_content。
-- ---------------------------------------------------------------------------
CREATE TABLE `inject_task` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `content_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '关联注入内容ID（inject_content.id）',
  `resource_type` tinyint(3) unsigned NOT NULL COMMENT '媒资类型：1-影片 2-节目 3-媒体',
  `resource_id` int(10) unsigned NOT NULL COMMENT '媒资主键（video.id / episode.id / media.id）',
  `video_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '所属影片ID（冗余，便于按片查询）',
  `episode_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '所属节目ID（影片级任务为0）',
  `media_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '所属媒体ID（影片/节目级任务为0）',
  `action` tinyint(3) unsigned NOT NULL COMMENT '操作：1-新增 2-修改 3-删除',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '状态：0-待注入 1-注入中',
  `create_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_status_id` (`status`,`id`),
  KEY `idx_content_id` (`content_id`),
  KEY `idx_resource` (`resource_type`,`resource_id`),
  KEY `idx_video_id` (`video_id`),
  KEY `idx_media_id` (`media_id`),
  KEY `idx_create_date` (`create_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='注入任务表';

-- ---------------------------------------------------------------------------
-- 注入内容（Inject Content）
-- 同一影片 / 节目 / 媒体全局仅一条。action 只覆盖为最后一次新增、修改或删除。
-- 调度根据本表决定是否继续下发；成功/失败为媒资当前注入结果。
-- ---------------------------------------------------------------------------
CREATE TABLE `inject_content` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '内容ID',
  `resource_type` tinyint(3) unsigned NOT NULL COMMENT '媒资类型：1-影片 2-节目 3-媒体',
  `resource_id` int(10) unsigned NOT NULL COMMENT '媒资主键（video.id / episode.id / media.id）',
  `video_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '所属影片ID（冗余，便于按片查询）',
  `episode_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '所属节目ID（影片级记录为0）',
  `media_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '所属媒体ID（影片/节目级记录为0）',
  `action` tinyint(3) unsigned NOT NULL COMMENT '最后一次操作：1-新增 2-修改 3-删除',
  `last_task_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '最近一次注入任务ID',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '状态：0-等待 1-注入中 2-成功 3-失败',
  `fail_reason` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '失败原因（成功时为空）',
  `create_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_resource` (`resource_type`,`resource_id`),
  KEY `idx_status` (`status`),
  KEY `idx_video_id` (`video_id`),
  KEY `idx_media_id` (`media_id`),
  KEY `idx_last_task_id` (`last_task_id`),
  KEY `idx_create_date` (`create_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='注入内容表';

-- ---------------------------------------------------------------------------
-- 注入日志（Inject Log）
-- 每次向下游 CDN（SP）发起注入写一条。同步状态看接口即时应答，异步状态看 CDN 回调。
-- ---------------------------------------------------------------------------
CREATE TABLE `inject_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `task_id` bigint(20) unsigned NOT NULL COMMENT '注入任务ID',
  `content_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '注入内容ID',
  `sp_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '下游内容服务商ID（sp.id）',
  `resource_type` tinyint(3) unsigned NOT NULL COMMENT '媒资类型：1-影片 2-节目 3-媒体',
  `resource_id` int(10) unsigned NOT NULL COMMENT '媒资主键（video.id / episode.id / media.id）',
  `video_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '所属影片ID（冗余，便于按片查询）',
  `episode_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '所属节目ID（影片级日志为0）',
  `media_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '所属媒体ID（影片/节目级日志为0）',
  `action` tinyint(3) unsigned NOT NULL COMMENT '操作：1-新增 2-修改 3-删除',
  `correlate_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '关联ID（请求/回调对账）',
  `sync_status` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '同步状态：0-等待 1-成功 2-失败',
  `async_status` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '异步状态：0-等待 1-成功 2-失败',
  `sync_message` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '同步应答摘要',
  `async_message` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '异步回调摘要',
  `create_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_content_id` (`content_id`),
  KEY `idx_sp_id` (`sp_id`),
  KEY `idx_correlate_id` (`correlate_id`),
  KEY `idx_resource` (`resource_type`,`resource_id`),
  KEY `idx_video_id` (`video_id`),
  KEY `idx_media_id` (`media_id`),
  KEY `idx_sync_async` (`sync_status`,`async_status`),
  KEY `idx_create_date` (`create_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='注入日志表';
