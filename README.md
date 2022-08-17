# CMDB

多云管理

## 如何启动
+ 先启动注册中心
+ 在启动用户中心
+ 在启动CMDB

## 架构图

## 项目说明

```
├── protocol                       # 脚手架功能: rpc / http 功能加载
│   ├── grpc.go              
│   └── http.go    
├── client                         # 脚手架功能: grpc 客户端实现 
│   ├── client.go              
│   └── config.go    
├── cmd                            # 脚手架功能: 处理程序启停参数，加载系统配置文件
│   ├── root.go             
│   └── start.go                
├── conf                           # 脚手架功能: 配置文件加载
│   ├── config.go                  # 配置文件定义
│   ├── load.go                    # 不同的配置加载方式
│   └── log.go                     # 日志配置文件
├── dist                           # 脚手架功能: 构建产物
├── etc                            # 配置文件
│   ├── xxx.env
│   └── xxx.toml
├── apps                            # 具体业务场景的领域包
│   ├── all
│   │   |-- grpc.go                # 注册所有GRPC服务模块, 暴露给框架GRPC服务器加载, 注意 导入有先后顺序。  
│   │   |-- http.go                # 注册所有HTTP服务模块, 暴露给框架HTTP服务器加载。                    
│   │   └── internal.go            #  注册所有内部服务模块, 无须对外暴露的服务, 用于内部依赖。 
│   ├── book                       # 具体业务场景领域服务 book
│   │   ├── http                   # http 
│   │   │    ├── book.go           # book 服务的http方法实现，请求参数处理、权限处理、数据响应等 
│   │   │    └── http.go           # 领域模块内的 http 路由处理，向系统层注册http服务
│   │   ├── impl                   # rpc
│   │   │    ├── book.go          # book 服务的rpc方法实现，请求参数处理、权限处理、数据响应等 
│   │   │    └── impl.go           # 领域模块内的 rpc 服务注册 ，向系统层注册rpc服务
│   │   ├──  pb                    # protobuf 定义
│   │   │     └── book.proto       # book proto 定义文件
│   │   ├── app.go                 # book app 只定义扩展
│   │   ├── book.pb.go             # protobuf 生成的文件
│   │   └── book_grpc.pb.go        # pb/book.proto 生成方法定义
├── version                        # 程序版本信息
│   └── version.go                    
├── README.md                    
├── main.go                        # Go程序唯一入口
├── Makefile                       # make 命令定义
└── go.mod                         # go mod 依赖定义
```

## 资源提供方
---provider                        # 资源都在第三方，比如阿里云，腾讯云，
云商有很多资源，要完成与CMDB的映射和交互
1. 已经能从云商（第三方接口）查询的数据
2. 数据分页（每一操作 控制查询的数据量，因为云商也有分页限制和速率限制，从而保证接口的性能）
    + 设计一个pagger(分页查询器)，基础：query page_size,page_number,
        一直查询，直到下一页没有数据，就停止查询
        - Next() bool : 控制是否有下一页数据
        - Scan() : 获取当页的数据
    + 增加令牌桶限流
3. 目前分页的逻辑划分的不是很开，现在是腾讯云cvm的分页，
    > 如果还需要些其他厂商的分页，那么这些分页逻辑，还需要再继续写，因为云商都有这两个限制。
    写多了就不利于代码的维护，那就需要把代码的功能抽象出来，分页和scan查询的逻辑分离开。
    > 因为对不同的产品和资源的scan查询的逻辑是不同的，对于腾讯云查询的是cvm，对于阿里云查询的是ecs。但是像next、offset、setPageSize的逻辑都是一样的，比如还可以加一个方法，setRate设置速率的方法，就可以把这些公用的逻辑抽象成一个共用模板。
    > pagger接口里面的scan方法，如果当不是查询hostset，是查RDS分页，就需要把hostset换成RDSset。如果是一直增加新的pagger对象，就不太好了。那就需要抽象一个通用的pagger定义。

4. 为啥腾讯云里面都是指针？

# 任务管理
融合3个模块
+ secret
+ provider
+ host service
/task/的API
```
# 定义一个类型，表示这个是资源同步的任务
type: "sync"
# secret,是哪个厂商的， 比如腾讯云secret
secret_id: "xxx"
# operater 按照资源划分, 比如操作主机
resource_type: "host"
# 指定操作那个地域的资源
region: "shanghai"
```

任务的状态：
+ 状态: running
+ 开启时间
+ 介绍时间
+ 执行日志


## 快速开发
make脚手架
```sh
➜  CMDB git:(master) ✗ make help
dep                            Get the dependencies
lint                           Lint Golang files
vet                            Run go vet
test                           Run unittests
test-coverage                  Run tests with coverage
build                          Local build
linux                          Linux build
run                            Run Server
clean                          Remove previous build
help                           Display this help screen
```

1. 使用安装依赖的Protobuf库(文件)
```sh
# 把依赖的probuf文件复制到/usr/local/include（mcube是哪个版本就复制哪个版本的proto文件）

# 创建protobuf文件目录
$ make -pv /usr/local/include/github.com/infraboard/mcube/pb

# 找到最新的mcube protobuf文件
$ ls `go env GOPATH`/pkg/mod/github.com/infraboard/

# 复制到/usr/local/include
$ cp -rf pb  /usr/local/include/github.com/infraboard/mcube/pb
```

2. 添加配置文件(默认读取位置: etc/CMDB.toml)
```sh
$ 编辑样例配置文件 etc/CMDB.toml.book
$ mv etc/CMDB.toml.book etc/CMDB.toml
```

3. 启动服务
```sh
# 编译protobuf文件, 生成代码
$ make gen
# 如果是MySQL, 执行SQL语句(docs/schema/tables.sql)
$ make init
# 下载项目的依赖
$ make dep
# 运行程序
$ make run
```

4. 注册逻辑用的不是IOC了，没有放在项目内部，放在mcube外部一个公共的库

## 相关文档
Swagger http://127.0.0.1:8050/apidocs.json
并安装 json fmt等插件

## 项目分析
目标：通过云商账号同步不同区域的资产
亮点：1、之前是认为管理的，现在用自动化提高了效率。
    2、把这个做成了一个系统之后，就成为企业面向数字化的一个基石。可以和外部的一些系统进行对接。比如一些财务系统。
    3、作为一个资源系统，这套理念是借鉴了k8s的资源管理。
        a.例如resource，有固有属性、IP、尺寸等，不随资源的改变而变化。
        b.关联属性：例如所有者，会进行变化。
        c.通用属性：无论哪些资源都是必有的，设计为公共表提升解锁效率。
（注：把protobuf定义出来，需要专门画一个流程图，有哪些功能，功能对应的数据结构是什么样的，功能的这个字段是在描述什么场景。很多字段都是产品定义）
资源管理：1、手动管理（录入、修改）。
    2、自动管理（对接第三方，创建、更新、删除、同步）
用户：能通过Tag或资源的属性IP在cmdb里面解锁资源。
资源搜索框：ip/ins_name/ins_decr/tags=app=appv1
设计：1.定义业务模型protobuf的定义、数据库设计（很多情况下业务模型是和数据库有关联的）
      2.编译protobuf文件，接口/数据模型有哪些字段，要保证接口的兼容性
      3.接口的实现类，但没有对外通过协议暴露使用接口
      4.提供API，http(restful)、GRPC(因为是通过protobuf生成的，生成后的接口不仅是内部接口同时满足grpc服务的一个实例，然后就能注册到grpc框架内来提供服务，把这个类给grpc.server。所以在protocol里面有一个LoadGrpcApp，会把ioc里面所有托管的grpc注册给grpc.server，然后只需要把grpc.server启动起来)
      5.交给框架
使用接口的好处：写mock




## 代码步骤
1.创建apps/resource、apps/resource/pb文件夹，resource.proto文件：定义数据结构
    make gen生成pb代码,生成3个pb.go文件
2.创建apps/resource/impl文件夹, apps/resource/impl/impl.go并实现注册逻辑、
    apps/resource/impl/resource.go并实现初步方法、
    apps/resource/app.go文件并定义AppName
    创建apps/resource/api文件夹
3.在docs/schema/tabils.sql里面写SQL语句 (show create table xxx;查看库的创建语句) 写完执行make init
    创建apps/resource/impl/sql.go，写go代码SQL
4.创建apps/resource/impl/resource.go
    4.1 在apps/resource/impl/resource.go
        4.1.1 写 Search()函数 的具体逻辑
        4.1.2 单独封装Query sql builder
    4.2 在apps/resource/app.go文件中
        4.2.1 定义SearchRequest是否有Tag，添加HasTag函数
        4.2.2 定义Tag的比较符号
        4.2.3 多个值比较的关系RelationShip()
        4.2.4 ResourceSet/Resource 构造函数
        4.2.5 将string格式处理成以“ ，”相隔的列表
        4.2.6 为ResourceSet补充Add方法
    4.3 创建apps/resource/impl/dao.go文件
        4.3.1 查询Tag方法
        在apps/resource/app.go文件中：
        4.3.2 补充NewDefaultTag构造函数
        4.3.3 将当前resourceSet里所有的resourceId拿出来，在专门封装一个函数resourceId（）
        4.3.4 添加更新resource Tag属性的方法
        4.3.5 添加AddTag方法
5.创建apps/resource/api/http.go文件、  apps/resource/api/resource.go文件
    5.1 在apps/resource/api/http.go文件写暴露http层api的具体实现
    5.2 在apps/resource/api/resource.go文件写handler的具体实现
        5.2.1 实现SearchResource函数
    5.3 在apps/resource/app.go文件中，
        5.3.1 构建NewSearchRequestFromHTTP
        5.3.2 构建NewTagsFromString
        5.3.3 实现AddTag函数
        5.3.4 实现ParExpr函数
6.创建apps/host(主机信息录入)、apps/host/api、apps/host/impl、apps/host/pb文件夹、apps/host/app.go文件
    6.1 在apps/host/app.go定义模块名称常量
    6.2 创建apps/host/pb/host.proto,写protobuf,然后make gen
    6.3 创建apps/host/impl/impl.go写实体类
    6.4 创建apps/host/impl/host.go写接口的初步实现
    6.5 创建apps/host/impl/sql.go写具体sql语句
    6.6 创建apps/host/impl/dao.go写具体进行数据库save操作的代码
    6.7 在apps/host/app.go封装host.GenHash方法
    6.8 创建utils文件夹、utils/hash.go,在utils/hash.go封装hash工具
    6.9 在apps/resource/app.go封装补充Information hash方法
    6.10 创建apps/resource/impl/tx.go单独封装SaveResource逻辑
    6.11 在apps/resource/app.go单独封装IPToString转换逻辑
    6.12 apps/resource/impl/tx.go单独封装updateResourceTag逻辑
    6.13 在apps/host/app.go，封装describe.string格式处理
    6.14 在apps/host/impl/host.go写SyncHost方法的具体保存实现
    6.15 在apps/host/app.go，
        6.15.1 DescribeHostRequestWithID构造函数
        6.15.2 NewUpdateHostDataByIns构造函数
        6.15.2 实现put方法
    6.16 在apps/resource/impl/tx.go单独封装UpdateResource逻辑
    6.17 在apps/host/impl/host.go写DescribeHost的具体实现
    6.18 在apps/host/app.go，
        6.18.1 写Where方法
        6.18.2 写NewDefaultHost构造函数
        6.18.3 加载的时候就根据逗号拆开
7.在apps/host/impl/host.go
    7.1 写queryHost的具体实现
            在apps/host/app.go，
                添加 NewHostSet()
                添加 (s *HostSet) Add
                添加 (s *HostSet) UpdateTag
8.host API，创建apps/host/api/host.go、apps/host/api/http.go
    8.1 在apps/host/api/http.go，写具体暴露http接口代码
    8.2 在apps/host/api/host.go，
        写CreateHost具体逻辑
        写QueryHost具体逻辑
            在apps/host/app.go，添加NewQueryHostRequestFromHTTP
    8.3 在apps/host/api/host.go，写DescribeHost具体逻辑
9.创建 provider（不可能所有的数据通过api录进来，需要通过访问云商，通过云商的接口拉进来）、
  provider/txyun、provider/txyun/connectivity(建立连接)、
  provider/txyun/cvm(操作虚拟机)文件夹,创建provider/txyun/connectivity/client.go、
  provider/txyun/connectivity/client_tset.go、provider/txyun/cvm/operator.go、
  provider/txyun/cvm/operator_test.go文件
    9.1 在provider/txyun/connectivity/client.go写具体逻辑
    9.2 在provider/txyun/connectivity/client_test.go写测试用例
            在provider/txyun/connectivity/client.go
                补充Chack函数 
                补充 AccountId() 函数
    9.3 创建ptr_value.go，写指针转string函数
    9.4 provider/txyun/cvm/operator.go写定义CVM结构
    9.5 创建provider/txyun/cvm/query.go 添加query方法
    9.6 完成数据转换逻辑
        utils/ptr_value.go 补充指针转换逻辑
        provider/txyun/cvm/operator.go补充parseTime、transferTags方法
        apps/resource/app.go，补充 NewThirdTag

10.增加分页功能，创建provider/txyun/cvm/pagger.go
    10.1 在host/app.go 增加pagger对象的接口定义、构造函数
    10.2 在provider/txyun/cvm/pagger.go 定义pagger对象（结构体）、Next和scan方法
    10.3 在host/app.go 为hostset增加Length方法
    10.4 增加计算offset、设置分页参数方法
    10.5 修改Req 执行真正的下一页的offset
    
11.增加令牌桶限流
    11.1 等待分配令牌

增加task模块，由它来联合host/provider这两个模块处理事务（之前是人为写）。
    调用provider，把调用的结果，通过task再调host API写到数据库。
增加secret模块，因为provider的参数都是从环境变量获取的，有可能有几个 provider ，
    或者每一个provider有很多账号，每个账号有很多region区域。
    目前的cmdb系统是可以拿着凭证就可以去访问云商，如果是只读还好，要是后期要做关机操作，很多自动化的逻辑依赖云商的key进行操作。如果key一旦被暴露出来，拿着去买机器或者干其他乱七八糟的，就非常的危险。
    需要专门对provider的账号做管理，增加secret模块管理访问资源方的凭证。
    secret key入库的时候需要加密存储。
    在secret返回的时候需要做脱敏，即使是密文也不应该显示。
每次查询的时候由task模块去secret模块把凭证查出来。然后把凭证传递给provider，
    然后由provider去操作资源提供方查询资源，把返回的结果通过task再调host API写到数据库

访问第三方：1.需要知道第三方的地址
    2.需要凭证信息（凭证类型：API key/password/token）(API key：api_key,api_secret)

12.增加secret模块,apps/secret
    12.1 pb
        secret.proto ：定义secret对象和接口
    12.2 impl
        impl.go:
        dao.go:
        secrte.go:具体方法的实现
        sql.go:--用于过滤的字段最好加上索引
    12.3 app.go：定义模块名字,
        封装secrte.go里面的构造函数和加解密、脱敏等方法
    12.4 api
        http.go: 
        secret.go: 


13.增加task模块，apps/task
    impl
        task.go：实现具体方法
        mock.go: 定义需要依赖但是没完成的模块对象及其方法等
        host_sync.go: 封装耗时逻辑，后台运行