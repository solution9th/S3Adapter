CREATE TABLE IF NOT EXISTS `%s`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `os_access_key` char(30) NOT NULL COMMENT '本地key',
  `os_screct_key` char(40) NOT NULL COMMENT '本地key',
  `engine_type` char(10) NOT NULL COMMENT '对象存储',
  `engine_region` char(20) NOT NULL COMMENT '引擎具体region',
  `engine_access_key` char(40) NOT NULL COMMENT '引擎具体的key',
  `engine_secret_key` char(40) NOT NULL COMMENT '引擎具体的key',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '添加时间',
  `app_name` varchar(10) NOT NULL COMMENT '应用名称',
  `app_remark` varchar(20) NOT NULL COMMENT '应用备注',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4

