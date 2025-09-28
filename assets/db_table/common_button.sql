create table common_button
(
    `id`             int unsigned auto_increment                not null primary key comment '按钮id',
    `module_id`      int unsigned     default 0                 not null comment '用于区分业务模块',
    `scene_id`       varchar(32)      default ''                not null comment '业务下的场景/页面id',
    `common_task_id` int unsigned     default 0                 not null comment '通用任务id',
    `enabled`        tinyint unsigned default 0                 not null comment '状态：0=未发布, 1=已发布',
    `sort_value`     int              default 999               not null comment '排序值. 正序, 排序值相同时以创建时间正序',
    `extend`         varchar(8192)    default '{}'              not null comment '扩展数据, json',
    -- 以下字段无业务逻辑, 或者是透传到客户端字段
    `button_title`   varchar(64)      default ''                not null comment '按钮标题',
    `button_desc`    varchar(128)     default ''                not null comment '按钮描述/副标题',
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

create table common_button_module
(
    `module_id`   int unsigned default 0  not null comment '用于区分模块', -- 这里是因为程序中要定义常量, 这里如果使用自增id可能导致测试环境和生产环境不一致
    -- 以下字段无业务逻辑, 或者是透传到客户端字段
    `module_name` varchar(64)  default '' not null comment '模块名, 用于开发/管理人员知道这个是做什么的, 无程序逻辑',
    `remark`      varchar(256) default '' not null comment '备注'
) comment '区分业务模块';

create unique index button_module_unique_index on common_button_module (`module_id`);

create table common_button_scene
(
    `module_id`  int unsigned default 0  not null comment '用于区分模块',-- 这里是因为程序中要定义常量, 这里如果使用自增id可能导致测试环境和生产环境不一致
    `scene_id`   varchar(32)  default '' not null comment '子场景id',-- 这里是因为程序中要定义常量
    -- 以下字段无业务逻辑, 或者是透传到客户端字段
    `scene_name` varchar(64)  default '' not null comment '子场景名',
    `remark`     varchar(256) default '' not null comment '备注'
) comment '区分业务模块的一个场景/页面';

create unique index button_scene_unique_index on common_button_scene (`module_id`, `scene_id`);

create table common_task
(
    `id`          int unsigned auto_increment             not null primary key comment '任务id',
    `module_id`   int unsigned  default 0                 not null comment '用于区分业务模块, 在业务逻辑中不会检查这个字段与button匹配',
    `scene_id`    varchar(32)   default ''                not null comment '业务下的场景/页面id, 在业务逻辑中不会检查这个字段与button匹配',
    `template_id` int unsigned  default 0                 not null comment '模板id',
    `start_time`  datetime                                not null comment '任务开始时间',
    `end_time`    datetime                                not null comment '任务结束时间',
    `task_target` int unsigned  default 0                 not null comment '任务目标',
    `prize_ids`   varchar(128)  default ''                not null comment '奖品id列表，逗号分隔',
    `hide_rule`   varchar(64)   default ''                not null comment '隐藏规则列表，逗号分隔：1=完成后隐藏 2=领奖后隐藏',
    `extend`      varchar(1024) default '{}'              not null comment '任务扩展, 一般用于存放任务模板无法确认的参数, 这些参数是运营决定的, 比如最近x天的x是多少',
    -- 以下字段无业务逻辑, 或者是透传到客户端字段
    `remark`      varchar(256)  default ''                not null comment '备注',
    `ctime`       datetime      default current_timestamp not null comment '创建时间',
    `utime`       datetime      default current_timestamp not null comment '更新时间' on update current_timestamp
) comment ='通用任务';

create index button_moduel_scene_index on common_task (`module_id`, `scene_id`);
create index button_end_time_index on common_task (`end_time`);

create table common_task_template
(
    `id`          int unsigned auto_increment    not null primary key comment '任务模板id',
    `period_type` smallint unsigned default 0    not null comment '任务周期：0=无周期 1=自然日 2=自然周(第一天是周日) 3=自然周(第一天是周一) 4=自然月',
    `task_type`   smallint unsigned default 0    not null comment '任务类型：1=跳转任务 2=签到',
    `extend`      varchar(8192)     default '{}' not null comment '扩展数据, 一般用于存放任务模板数据的参数, 这些参数是开发者决定的, 比如第三方任务的id和secret',
    -- 以下字段无业务逻辑, 或者是透传到客户端字段
    `remark`      varchar(256)      default ''   not null comment '备注'
) comment ='通用任务模板'
