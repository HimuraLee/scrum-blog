SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for cate
-- ----------------------------
DROP TABLE IF EXISTS `cate`;
CREATE TABLE `cate` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '分类名',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO cate VALUE (1, 'other');
-- ----------------------------
-- Table structure for post
-- ----------------------------
DROP TABLE IF EXISTS `post`;
CREATE TABLE `post` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `cate_id` int(11) NOT NULL DEFAULT 0,
  `cate_name` varchar(255) NOT NULL DEFAULT '' COMMENT '分类名',
  `status` tinyint(4) NOT NULL DEFAULT 0 COMMENT '发布状态：0 草稿，1 已发布',
  `title` varchar(255) NOT NULL COMMENT '文章标题',
  `passwd` varchar(255) NOT NULL DEFAULT '' COMMENT '文章密码，空代表无密码',
  `filename` varchar(255) NOT NULL DEFAULT '' COMMENT '博客本地文件名',
  `markdown_content` longtext NOT NULL COMMENT '博客内容md格式',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `cate_post` (`cate_id`),
  KEY `create_time` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `post` VALUES ('1', '0', '', '0', '关于我', '', 'about-me', '::: tip 介绍\n暫時沒有個人介紹!\n:::', CURRENT_TIME, CURRENT_TIME);
INSERT INTO `post` VALUES ('2', '0', '', '0', '留言板', '', 'message-board', '::: tip\n欢迎大家在此留下你的建议和意见。\n:::', CURRENT_TIME, CURRENT_TIME);

DROP TABLE IF EXISTS `post_tag`;
CREATE TABLE `post_tag` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `post_id` int(11) NOT NULL,
  `tag_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `post_tag` (`post_id`,`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for tag
-- ----------------------------
DROP TABLE IF EXISTS `tag`;
CREATE TABLE `tag` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '标签名',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '用户名',
  `passwd` varchar(255) NOT NULL DEFAULT '' COMMENT '密码',
  `author_name` varchar(255) NOT NULL DEFAULT '' COMMENT '笔名',
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Records of user
-- ----------------------------
INSERT INTO `user` VALUES ('1', 'himura', 'f01a75sd5d43g1bae1s4dd69fe4lf6af3ddaj75k24g86sb0', 'himura', CURRENT_TIME, CURRENT_TIME);
