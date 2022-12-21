# vec init

1. 向量化：数据库所有专利 -> 调用算法模型接口将特定字段向量化 -> 将所有向量存到同一个文件中，并在 Redis 中保存向量行号与向量id的映射
2. 索引构建：调用 Diskann ，将向量文件作为输入，生成索引文件
3. 查询：用户输入字符串，调用算法模型接口将字符串向量化 -> 调用 Diskann ，根据索引文件进行查询，得到最相似的 top k 个向量id -> 根据向量id在 Redis 中查询向量行号 -> 根据向量行号在数据库中查询专利信息

## 1 特性

* [x] 天然支持多字段向量化，只需改动配置文件即可
* [x] 将数据库查询 -> 向量化 -> 向量存储 -> redis 更新全部并行化
* [x] channel 通信，全流程异步
* [ ] Dockerfile
* [x] 数据库并发查询与自动调速
* [ ] 支持任意DDL数据库
* [ ] 支持 PostgreSQL, Oracle

## 2 先决条件

### 1. 环境

* Go 1.19+
* Redis

### 2. 数据

* Mysql 并建立好数据库，导入好数据

### 3. 模型服务接口

需要自己准备一个模型服务接口，接口的输入是n个字符串，输出是n个m维向量，向量的维度任意。模型服务需满足（主要是key）：

POST 请求，URL 定义在 application.yaml 中，Request Body：

```json
{
    "strarr": [
        "string1",
        "string2",
        "string3"
    ]
}
```

Response Body：

```json
{
    "data": [
        [
            0.1,
            0.2,
            0.3
        ],
        [
            0.4,
            0.5,
            0.6
        ],
        [
            0.7,
            0.8,
            0.9
        ]
    ]
}
```

## 3 运行

把 `application.example.yml` 改名为 `application.yml` ，并修改其中的配置

### 3.1 初始化（向量化、索引构建）

```bash
go run cmd/init/main.go
```

运行结束后，程序会完成以下内容：
1. **读取全库**: 通过配置文件中定义的数据库和专利字段(field)，读取全库数据
2. **调用向量化接口**: 对于每条专利的多个字段(field)，分别调用向量化接口。并将向量存储在 "`server.VectorDir`/init/<field>.fvecs" 文件中。假设配置中定义了 n 个字段，那么最后会生成 n 个 `<field>.fvecs` 文件，每个文件都是所有字段的该 field 的向量的集合。
3. **建立索引**：使用前面的 n 个 `<filed>.fvecs` 文件，调用上游的 Diskann 服务，对 n 个 field 建立 n 个索引。（其中 Diskann 服务和本项目有一定的耦合，对接口和部署方式的要求后续再补充）

### 3.2 相似专利查询

#### 3.2.1 运行

```bash
go run cmd/query/main.go
```

#### 3.2.2 查询请求格式

POST，URL 是 `ip:port/query`，其中 `port` 是在 application.yml 中的 `server.port` 定义的。

Body 示例：

```json
{
    "field": "abstract",
    "query": "这是请求内容，比如现在的field是abstract，表明要在摘要里搜数据，所以query传你想搜索的摘要"
}
```

其中，目前支持的 field 有："name", "abstract", "claim"。

Response: 

HTTP 状态码200为成功，其余请参照 code 和 msg。

```json
{
  "code": 200,
  "msg": "ok",
  "data": [
    {"id":  "1", "score": 0.9},
    {"id":  "2", "score": 0.8},
    {"id":  "3", "score": 0.7},
    {"id":  "4", "score": 0.6},
    {"id":  "5", "score": 0.5},
    {"id":  "6", "score": 0.4},
    {"id":  "7", "score": 0.3},
    {"id":  "8", "score": 0.2},
    {"id":  "9", "score": 0.1}
  ]
}
```

其中 data 是专利 id 和相似度的数组，按相似度从高到低排序。

