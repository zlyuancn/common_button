create table common_button_app
(
    `id`       int unsigned auto_increment not null primary key,
    `appid`    int unsigned default 0      not null comment '用于区分业务', -- 这里是因为程序中要定义常量, 这里如果使用自增id可能导致测试环境和生产环境不一致
    `app_name` varchar(64)  default ''     not null comment '业务名, 用于开发/管理人员知道这个是做什么的, 无程序逻辑',
    `remark`   varchar(256) default ''     not null comment '备注'
) comment '区分app业务';

create unique index button_app_unique_index on common_button_app (`appid`);
