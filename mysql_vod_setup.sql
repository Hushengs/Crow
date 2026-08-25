-- VOD 媒资：分类 / 影片 / 节目 / 媒体
-- 依赖：先执行 mysql_system_setup.sql 创建库 kratos
--
-- 职责划分：
--   video_category 分类。支持多级树形分类。
--   video    影片。节目集/电影等标题级元数据。
--   episode  节目。挂在影片下的分集/单集，通过 video_id 归属。
--   media    媒体。实际可播文件/流资源；挂 episode_id，业务侧用 media_id 标识。

USE `kratos`;

-- ---------------------------------------------------------------------------
-- 影片分类（Video Category）
-- 邻接表结构，parent_id=0 表示根分类。
-- ---------------------------------------------------------------------------
CREATE TABLE `video_category` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '分类ID',
  `parent_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '父分类ID，根分类为0',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '分类名称',
  `sort_order` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '同级排序值',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '状态：0-禁用 1-正常',
  `create_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_parent_name` (`parent_id`,`name`),
  KEY `idx_parent_sort` (`parent_id`,`sort_order`,`id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='影片分类表';

-- ---------------------------------------------------------------------------
-- 影片（Video）
-- 标题级内容：电影、剧集、综艺等；分集与媒体均挂到本表。
-- ---------------------------------------------------------------------------
CREATE TABLE `video` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '影片ID',
  `category_id` int(10) unsigned NOT NULL COMMENT '所属分类ID',
  `video_code` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '影片编码（业务侧唯一）',
  `title` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '影片标题',
  `subtitle` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '副标题',
  `video_type` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '类型：1-电影 2-剧集 3-综艺 4-其他',
  `poster_vertical_url` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '海报竖图地址',
  `poster_horizontal_url` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '海报横图地址',
  `thumbnail_url` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '缩略图地址',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '简介',
  `year` smallint(5) unsigned DEFAULT NULL COMMENT '出品年份',
  `duration` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '总时长（秒，剧集可为各集合计）',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '状态：0-禁用1-正常',
  `create_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_video_code` (`video_code`),
  KEY `idx_category_id` (`category_id`),
  KEY `idx_title` (`title`),
  KEY `idx_video_type` (`video_type`),
  KEY `idx_status` (`status`),
  KEY `idx_create_date` (`create_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='影片表';

-- ---------------------------------------------------------------------------
-- 节目（Episode）
-- 影片下的分集/单集；电影可仅有一集，剧集按 episode_no 排序。
-- ---------------------------------------------------------------------------
CREATE TABLE `episode` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '节目ID',
  `video_id` int(10) unsigned NOT NULL COMMENT '所属影片ID',
  `episode_no` int(10) unsigned NOT NULL DEFAULT '1' COMMENT '集序号（从1开始）',
  `title` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '节目标题',
  `duration` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '时长（秒）',
  `description` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '节目简介',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '状态：0-禁用 1-正常',
  `create_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_video_episode_no` (`video_id`,`episode_no`),
  KEY `idx_video_id` (`video_id`),
  KEY `idx_status` (`status`),
  KEY `idx_create_date` (`create_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='节目表';

-- ---------------------------------------------------------------------------
-- 媒体（Media）
-- 可播资源（文件/流）；归属节目，冗余 video_id 便于按影片聚合查询。
-- ---------------------------------------------------------------------------
CREATE TABLE `media` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `video_id` int(10) unsigned NOT NULL COMMENT '所属影片ID',
  `episode_id` int(10) unsigned NOT NULL COMMENT '所属节目ID',
  `media_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '媒体编码（协议/CDN侧唯一）',
  `media_url` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '媒体地址（播放/拉取URL）',
  `file_format` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '封装/格式（如 mp4、ts、m3u8）',
  `bitrate` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '码率（kbps）',
  `resolution` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '分辨率（如 1920x1080）',
  `file_size` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '文件大小（字节）',
  `duration` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '时长（秒）',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '状态：0-禁用 1-正常',
  `create_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_media_id` (`media_id`),
  KEY `idx_video_id` (`video_id`),
  KEY `idx_episode_id` (`episode_id`),
  KEY `idx_status` (`status`),
  KEY `idx_create_date` (`create_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='媒体表';
