CREATE DATABASE IF NOT EXISTS `kratos`
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE `kratos`;

CREATE TABLE `admin` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '管理员ID',
  `username` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '登录用户名',
  `password` char(60) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密码（存储password_hash加密后的密文）',
  `real_name` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '真实姓名',
  `role_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '角色ID（关联角色表，0表示无角色）',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '状态：0-禁用 1-正常 2-锁定（密码错误过多）',
  `last_login_time` datetime DEFAULT NULL COMMENT '最后登录时间',
  `password_updated_at` datetime DEFAULT NULL COMMENT '密码最后修改时间',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '备注',
  `create_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`create_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台管理员表';

CREATE TABLE `admin_operation_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `admin_id` int(10) unsigned NOT NULL COMMENT '操作人ID（关联管理员表）',
  `admin_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '操作人姓名（冗余字段，防止管理员被删除后丢失）',
  `module` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '操作模块（如：user, order, goods, system）',
  `action` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '操作动作（如：add, edit, delete, export, login）',
  `description` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '操作描述（人类可读的摘要）',
  `request_method` varchar(8) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '请求方法（GET/POST/PUT/DELETE）',
  `request_url` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '请求URL',
  `request_params` json DEFAULT NULL COMMENT '请求参数（JSON格式，存储完整入参）',
  `create_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_create_date` (`create_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台管理员操作日志表';

INSERT INTO `admin` (`username`, `password`)
VALUES ('admin', '123456')
ON DUPLICATE KEY UPDATE `password` = VALUES(`password`);
