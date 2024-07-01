create table common_button_scene
(
    `id`         int unsigned auto_increment not null primary key,
    `module_id`  int unsigned default 0      not null comment '用于区分模块',-- 这里是因为程序中要定义常量, 这里如果使用自增id可能导致测试环境和生产环境不一致
    `scene_id`   int unsigned default 0      not null comment '子场景id',-- 这里是因为程序中要定义常量, 这里如果使用自增id可能导致测试环境和生产环境不一致
    -- 以下字段无业务逻辑, 或者是透传到客户端字段
    `scene_name` varchar(64)  default ''     not null comment '子场景名',
    `remark`     varchar(256) default ''     not null comment '备注'
) comment '区分业务模块的一个场景/页面';

create unique index button_scene_unique_index on common_button_scene (`module_id`, `scene_id`);
