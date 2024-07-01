create table common_button
(
    `id`             int unsigned auto_increment                not null primary key,
    `module_id`      int unsigned     default 0                 not null comment '用于区分业务模块',
    `scene_id`       int unsigned     default 0                 not null comment '业务下的场景/页面id',
    `common_task_id` int unsigned     default 0                 not null comment '通用任务id',
    `enabled`        tinyint unsigned default 0                 not null comment '状态：0=未发布, 1=已发布',
    `sort_value`     int              default 0                 not null comment '正序, 排序值相同时以创建时间正序',
    `extend`         varchar(8192)    default '{}'              not null comment '扩展数据, json',
    -- 以下字段无业务逻辑, 或者是透传到客户端字段
    `button_title`   varchar(64)      default ''                not null comment '按钮标题',
    `button_desc`    varchar(128)     default ''                not null comment '按钮描述',
    `icon1`          varchar(128)     default ''                not null comment '图片1',
    `icon2`          varchar(128)     default ''                not null comment '图片2',
    `icon3`          varchar(128)     default ''                not null comment '图片3',
    `skip_value`     varchar(1024)    default ''                not null comment '跳转地址',
    `skip_title`     varchar(32)      default ''                not null comment '跳转按钮标题',
    `remark`         varchar(256)     default ''                not null comment '备注',
    `ctime`          datetime         default current_timestamp not null comment '创建时间',
    `utime`          datetime         default current_timestamp not null comment '更新时间' on update current_timestamp
) comment '通用按钮';

create index button_module_scene_index on common_button (`module_id`, `scene_id`, `sort_value`);
