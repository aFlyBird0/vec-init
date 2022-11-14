# vec init

* 本项目：数据库所有专利 -> 调用算法模型接口将特定字段向量化 -> 将所有向量存到同一个文件中，并在 Redis 中保存向量行号与向量id的映射
* 后续（索引构建）：调用 Diskann ，将向量文件作为输入，生成索引文件
* 后续（查询）：用户输入字符串，调用算法模型接口将字符串向量化 -> 调用 Diskann ，根据索引文件进行查询，得到最相似的 top k 个向量id -> 根据向量id在 Redis 中查询向量行号 -> 根据向量行号在数据库中查询专利信息

## features

* [x] 天然支持多字段向量化，只需改动配置文件即可
* [x] 将数据库查询 -> 向量化 -> 向量存储 -> redis 更新全部并行化
* [x] channel （缓冲池）
* [ ] Dockerfile
* [ ] 数据库并发查询与自动调速
* [ ] 支持任意DDL数据库
* [ ] 支持 PostgreSQL, Oracle

## run

把 `application.example.yml` 改名为 `application.yml` ，并修改其中的配置

```bash
go run main.go
```
