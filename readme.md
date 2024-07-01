
<!-- TOC -->

- [common\_button 是什么](#common_button-是什么)
- [术语](#术语)
  - [module 模块](#module-模块)
  - [scene 场景/页面](#scene-场景页面)
  - [button 按钮](#button-按钮)
  - [task\_template 任务模板](#task_template-任务模板)
  - [task 任务](#task-任务)

<!-- /TOC -->

---

# common_button 是什么

common_button 是一个通用按钮库, 用于快速方便的创建app上使用的一些按钮. 运营可以在不对app发版的情况下增加/删除/修改按钮

![](./assets/button1.png)![](./assets/button2.png)![](./assets/button3.png)

---

# 术语

## module 模块

module 表示一个业务下的一个模块划分, 比如`商城/论坛/用户中心`表示不同的模块

## scene 场景/页面

scene 表示不同的场景/页面, 比如用户中心的`个人信息/用户协议`表示不同的页面. 也可以用来区分同一个页面中的不同位置.

## button 按钮

button 表示一个场景/页面上可以点击的按钮, 这些简单的按钮不能有复杂的业务逻辑

## task_template 任务模板

task_template 用于定义一个任务类型, 比如 跳转/签到/第三方任务, 它是开发者决定的.

## task 任务

task 表示一个任务, 是运营基于 task_template 创建的一个任务实体. task 要依赖于 button 才能显示在客户端上.

![](./assets/task.png)

---
