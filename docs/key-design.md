# Etcd Key 设计

- 当前配置指针: /config/{app}/{env}/current
- 版本计数器: /config/{app}/{env}/version_counter
- 配置内容: /config/{app}/{env}/versions/{version}/{group}/{key}
- 版本元数据: /config/{app}/{env}/meta/{version}