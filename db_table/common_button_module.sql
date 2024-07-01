create table common_button_module
(
    `module_id`   int unsigned default 0      not null comment '用于区分模块', -- 这里是因为程序中要定义常量, 这里如果使用自增id可能导致测试环境和生产环境不一致
    -- 以下字段无业务逻辑, 或者是透传到客户端字段
    `module_name` varchar(64)  default ''     not null comment '模块名, 用于开发/管理人员知道这个是做什么的, 无程序逻辑',
    `remark`      varchar(256) default ''     not null comment '备注'
) comment '区分业务模块';

create unique index button_module_unique_index on common_button_module (`module_id`);
