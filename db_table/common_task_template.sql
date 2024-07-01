create table common_task_template
(
    `id`          int unsigned auto_increment   not null primary key,
    `period_type` tinyint unsigned default 0    not null comment '任务周期：0=无周期 1=自然日 2=自然周',
    `task_type`   tinyint unsigned default 0    not null comment '任务类型：1=跳转任务 2=签到',
    `extend`      varchar(8192)    default '{}' not null comment '扩展数据, 一般用于存放任务模板数据的参数, 这些参数是开发者决定的, 比如第三方任务的id和secret',
    `remark`      varchar(256)     default ''   not null comment '备注'
) comment ='通用任务模板'
