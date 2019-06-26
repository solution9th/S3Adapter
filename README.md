# S3Adapter

[![GoDoc](https://godoc.org/github.com/solution9th/S3Adapter?status.svg)](https://godoc.org/github.com/solution9th/S3Adapter) [![Go Report Card](https://goreportcard.com/badge/github.com/solution9th/S3Adapter)](https://goreportcard.com/report/github.com/solution9th/S3Adapter) [![license](https://img.shields.io/github/license/solution9th/S3Adapter.svg)](https://github.com/solution9th/S3Adapter)

S3Adapter 是一款兼容 AWS S3 的轻量级聚合对象存储网关。

S3Adapter 实现了不同对象存储后端使用统一的 REST API 进行对象存储操作，并且支持多种对象存储后端，已经支持了：s3、cos(腾讯云)等。由于 REST API 兼容 AWS S3，所以可以使用 AWS S3 SDK 透明的进行对象存储操作。

![](docs/s3.jpg)

## 特性

- 兼容 AWS S3 REST 接口，可以使用 AWS S3 SDK 进行调用
- 支持多种对象存储后端，现已支持: S3(AWS)、COS(腾讯云)
- 支持虚拟托管类型和路径类型 URI
- 轻量级，全平台，单文件，易部署

## 快速开始

### 安装

以二进制安装为例，更多安装方法请看 **[安装](docs/README.md#安装)**

```shell
$ curl -LO https://github.com/solution9th/S3Adapter/release/v1.14.0/S3Adapter
$ chmod +x S3Adapter
# 查看版本信息
$ ./S3Adapter version
```

### 启动

需要配置信息，支持多种配置读取方式（命令行参数、环境变量、配置文件），而且还自带了默认配置文件（不推荐使用）

以命令行参数为例，更多配置文件配置方法请看 `S3Adapter web --help` 或者 [配置文件](/docs/README.md#配置文件)

```shell
$ ./S3Adapter web  \
    --endpoint=s3.newio.cc \
    --mysqlhost=127.0.0.1
```

### 创建应用

通过 AccessKey 、 SecretKey、后端引擎名称(s3,cos) 和 后端引擎地区 换取 Key

```shell
$ curl -X PUT \
  https://os-proxy.develenv.com \
  -H 'Content-Type: application/xml' \
  -H 'Host: os-proxy.develenv.com' \
  -d '<CreateApplicationConfiguration>
    <AccessKey>123</AccessKey>
    <SecretKey>123</SecretKey>
    <Engine>s3</Engine>
    <Region>beijing</Region>
    <AppName>Test</AppName>
    <AppRemark>Test</AppRemark>
</CreateApplicationConfiguration>'
```

请求

```xml
<CreateApplicationConfiguration>
    <AccessKey>后端引擎的AccessKey(必填)</AccessKey>
    <SecretKey>后端引擎的SecretKey(必填)</SecretKey>
    <Engine>后端引擎的名称(必填)</Engine>
    <Region>后端引擎的地区(必填)</Region>
    <AppName>应用名称</AppName>
    <AppRemark>应用备注</AppRemark>
</CreateApplicationConfiguration>
```

响应

```xml
<CreateApplicationResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
    <AccessKey>yP7m8qHqo6tGnpZnpstw</AccessKey>
    <SecretKey>QLLrLCnNc14jUZ3pumD62Sjmwir2r6Ip2vhB9ury</SecretKey>
</CreateApplicationResult>
```

调用 S3Adapter 的接口都需要使用 `yP7m8qHqo6tGnpZnpstw` `QLLrLCnNc14jUZ3pumD62Sjmwir2r6Ip2vhB9ury` 秘钥对

如果想更换后端存储则只需要创建新应用即可

### 使用

直接使用 AWS S3 SDK 进行调用

1. endpoint 地址改成 S3Adapter 的 endpoint 
2. AccessKey 和 SecretKey 使用换取后的 Key 即可

以 *CreateBucket* 为例


```go
package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {

	var (
		accessKey  = "<new accessKey>"
		secretKey  = "<new secretKey>"
		region     = "<region>"
		endPoint   = "<endpoint>"
		bucketName = "<bucket name>"
	)

	sess, err := session.NewSession(&aws.Config{
		Endpoint:    aws.String(endPoint),
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		panic(err)
	}

	client := s3.New(sess)

	output, err := client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(output.String())
}
```

## 文档

[简体中文](docs/README.md)

## 许可协议

MIT