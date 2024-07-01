create table common_task
(
    `id`          int unsigned auto_increment                not null primary key,
    `appid`       int unsigned     default 0                 not null comment '业务id',
    `scene_id`    int unsigned     default 0                 not null comment '业务下的场景id',
    `template_id` int unsigned     default 0                 not null comment '模板id',
    `start_time`  datetime                                   not null comment '任务开始时间',
    `end_time`    datetime                                   not null comment '任务结束时间',
    `task_target` int unsigned     default 0                 not null comment '任务目标',
    `prize_ids`   varchar(128)     default ''                not null comment '奖品id列表，逗号分隔',
    `hide_rule`   varchar(64)      default ''                not null comment '隐藏规则列表，逗号分隔：1完成后隐藏、2领奖后隐藏',
    `enabled`     tinyint unsigned default 0                 not null comment '状态：0未发布、1已发布',
    `extend`      varchar(1024)    default '{}'              not null comment '任务扩展, 一般用于存放任务模板无法确认的参数, 这些参数是运营决定的, 比如最近x天的x是多少',
    `remark`      varchar(256)     default ''                not null comment '备注',
    `ctime`       datetime         default current_timestamp not null comment '创建时间',
    `utime`       datetime         default current_timestamp not null comment '更新时间' on update current_timestamp
) comment ='通用任务';

create index button_app_scene_index on common_task (`appid`, `scene_id`);
