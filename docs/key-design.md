# Etcd Key 设计

本文档说明配置中心在 Etcd 中的 Key 结构，以及版本号、current 指针的更新规则。

## 1. Key 结构

```text
/config/{app}/{env}/current
/config/{app}/{env}/version_counter
/config/{app}/{env}/versions/{version}/{group}/{key}
/config/{app}/{env}/meta/{version}
```

各字段含义：

| 字段 | 含义 | 示例 |
| --- | --- | --- |
| app | 应用名 | order-service |
| env | 环境 | dev / test / prod |
| version | 配置版本号 | v1 / v2 / v3 |
| group | 配置分组 | database / redis |
| key | 配置项名称 | host / port |

各类 Key 的作用：

| Key | 作用 |
| --- | --- |
| `/config/{app}/{env}/current` | 保存当前生效版本，例如 `v3` |
| `/config/{app}/{env}/version_counter` | 保存当前最大版本序号，例如 `3` |
| `/config/{app}/{env}/versions/{version}/{group}/{key}` | 保存某个版本下的具体配置项 |
| `/config/{app}/{env}/meta/{version}` | 保存某个版本的元数据 |

## 2. 示例

以 `order-service` 的 `prod` 环境为例：

```text
/config/order-service/prod/current = v2
/config/order-service/prod/version_counter = 2

/config/order-service/prod/versions/v2/database/host = 127.0.0.1
/config/order-service/prod/versions/v2/database/port = 3306
/config/order-service/prod/versions/v2/redis/addr = 127.0.0.1:6379

/config/order-service/prod/meta/v2 = {
  "version": "v2",
  "revision": 12,
  "publisher": "A",
  "comment": "update database config",
  "created_at": 1782787200
}
```

## 3. current 设计

`current` 用来记录当前正在生效的配置版本。

客户端查询配置时，不需要知道最新版本号，只需要先读取：

```text
/config/{app}/{env}/current
```

拿到当前版本后，再读取对应版本下的配置内容。

例如：

```text
/config/order-service/prod/current = v2
```

表示当前生效配置来自：

```text
/config/order-service/prod/versions/v2/
```

规则：

1. `current` 只保存版本号，不保存完整配置内容。
2. 每次发布新配置成功后，`current` 都要更新为新版本。
3. 历史版本不删除，方便查询、回滚和审计。

## 4. version_counter 设计

`version_counter` 用来生成下一个版本号。

规则：

1. 第一次发布时，如果 `version_counter` 不存在，则新版本为 `v1`。
2. 如果 `version_counter = 2`，下一次发布生成 `v3`。
3. 发布成功后，`version_counter` 更新为最新版本序号。

示例：

```text
第一次发布：
version_counter 不存在
生成 v1
写入 version_counter = 1
写入 current = v1

第二次发布：
读取 version_counter = 1
生成 v2
写入 version_counter = 2
写入 current = v2
```

## 5. 发布配置流程

发布配置时，服务端按下面流程处理：

1. 校验请求参数：`app`、`env`、`configs` 不能为空。
2. 读取 `/config/{app}/{env}/version_counter`。
3. 根据计数器生成新版本号，例如 `v1`、`v2`、`v3`。
4. 将配置内容写入 `/config/{app}/{env}/versions/{version}/{group}/{key}`。
5. 写入版本元数据 `/config/{app}/{env}/meta/{version}`。
6. 更新 `/config/{app}/{env}/current = {version}`。
7. 更新 `/config/{app}/{env}/version_counter = {version_number}`。
8. 返回新版本号和 Etcd revision。

写入顺序建议：

```text
先写 versions
再写 meta
最后更新 current 和 version_counter
```

原因是 `current` 代表当前可用版本，必须等配置内容和元数据都写完后才能切换。

## 6. 查询当前配置流程

查询当前配置时，服务端按下面流程处理：

1. 读取 `/config/{app}/{env}/current`，得到当前版本号。
2. 如果 `current` 不存在，返回配置不存在。
3. 拼接版本配置前缀 `/config/{app}/{env}/versions/{current}/`。
4. 使用前缀查询读取该版本下所有配置项。
5. 将 Etcd 内部路径转换成对外返回的配置结构。

示例：

```text
读取 current 得到 v2
读取 /config/order-service/prod/versions/v2/
返回 database.host、database.port、redis.addr 等配置项
```

## 7. 回滚流程

回滚不是简单地把 `current` 改回旧版本，而是把目标版本复制成一个新版本。

例如当前版本是 `v3`，要回滚到 `v1`：

1. 读取当前版本 `/config/{app}/{env}/current`，得到 `v3`。
2. 读取目标版本 `/config/{app}/{env}/versions/v1/` 下所有配置。
3. 读取 `version_counter`，生成新版本 `v4`。
4. 将 `v1` 的配置内容复制写入 `/config/{app}/{env}/versions/v4/`。
5. 写入 `/config/{app}/{env}/meta/v4`，说明这是一次回滚发布。
6. 更新 `/config/{app}/{env}/current = v4`。
7. 更新 `/config/{app}/{env}/version_counter = 4`。

回滚后：

```text
current = v4
version_counter = 4
versions/v4 的配置内容与 versions/v1 相同
```

这样设计的原因：

1. 回滚也是一次新的配置变更，应该留下记录。
2. 不直接把 `current` 指回旧版本，可以保留完整审计链路。
3. 后续仍然可以继续从最新版本号往后发布。

## 8. Day1 结论

Day1 需要确认的版本规则如下：

1. 当前生效版本由 `current` 决定。
2. 新版本号由 `version_counter` 自增生成。
3. 每次发布都写入新的 `versions/{version}`。
4. 每次发布都写入对应的 `meta/{version}`。
5. 回滚会生成新版本，不会直接修改 `current` 指向旧版本。
