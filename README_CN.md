# StellOrbit Go SDK

[English](./README.md) | 简体中文

`stellorbit-go-sdk` 是 [`stellhub/stellorbit-service`](https://github.com/stellhub/stellorbit-service) 的 Go 客户端 SDK。`stellorbit-service` 是 Stell 体系中的服务治理引擎，负责路由、负载均衡、重试、流量迁移和服务生命周期策略。

SDK 通过 [`stellnula-go-sdk`](https://github.com/stellhub/stellnula-go-sdk) 从 [`stellnula-service`](https://github.com/stellhub/stellnula-service) 消费治理规则，在本地维护不可变规则注册表，并向 Go 应用和中间件适配层暴露规则 Provider。

## 定位

本仓库面向需要消费 StellOrbit 治理能力的应用、平台服务和中间件组件，提供 Go 客户端实现。

核心包不直接实现熔断状态机、限流算法、鉴权拦截器或路由执行器。这些动作应由应用中间件、HTTP handler、gRPC interceptor、网关或框架适配层基于 SDK 返回的规则自行执行。

## 能力

- 从 StellNula bootstrap 治理规则。
- 订阅 `governance/service-governance` 规则通道。
- 构建本地不可变治理规则注册表。
- 提供基础规则 Provider：
  - `RouteRuleProvider`
  - `CircuitBreakerRuleProvider`
  - `AuthorizationRuleProvider`
  - `RateLimitRuleProvider`
- 当更新部分无效时保留 last-known-good 规则。
- 通过兼容 HTTP API 请求服务间路由决策。
- 通过兼容 HTTP API 查询服务生命周期和流量策略。
- 支持 API key 请求头认证。
- 兼容调用使用标准 Go `net/http` transport。
- 支持请求超时配置。
- 所有客户端调用都传播 `context.Context`。

## 当前状态

| 项目 | 值 |
| --- | --- |
| 稳定性 | 早期开发 |
| 语言 | Go |
| 规则来源 | `stellnula-service` |
| 传输 | `net/http`、StellNula 数据面 |
| 目标服务 | `stellorbit-service` |
| 维护方 | StellHub |

## 安装

```bash
go get github.com/stellhub/stellorbit-go-sdk
```

## 快速开始

从 StellNula 消费治理规则：

```go
package main

import (
	"context"
	"fmt"
	"log"

	stellorbit "github.com/stellhub/stellorbit-go-sdk"
)

func main() {
	client, err := stellorbit.NewClient(stellorbit.Options{
		StellnulaEndpoint: "http://localhost:8060",
		AppID:             "payment-service",
		ClientID:          "payment-service-local",
		Env:               "dev",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if err := client.Start(context.Background()); err != nil {
		log.Fatal(err)
	}

	rules, err := client.Routes().Find(context.Background(), stellorbit.RouteRuleQuery{
		ServiceName: "payment-service",
		RouteKey:    "tenant-a",
		Attributes: map[string]string{
			"env": "dev",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rules)
}
```

SDK 订阅固定通道：

| 字段 | 值 |
| --- | --- |
| `namespace` | `governance` |
| `group` | `service-governance` |

治理规则配置由 `stellorbit-service` 以聚合 payload 的形式发布。每个 StellNula 配置项使用固定 `configId`，格式为 `stellorbit.<applicationCode>.<ruleType>`，并且必须携带 `schemaVersion=stellorbit.governance.aggregate.v1`。SDK 从聚合 payload 的 `rules[]` 中解析规则；`rules[].ruleId` 是规则身份，聚合层 `configId` 用于整组规则的替换、删除和回退。聚合 root 必须包含对应类型的 validator 字段：`routes`、`limit`、`breaker` 或 `auth`。

原始 StellOrbit HTTP API 仍可用于兼容调用：

```go
package main

import (
	"context"
	"fmt"
	"log"

	stellorbit "github.com/stellhub/stellorbit-go-sdk"
)

func main() {
	client, err := stellorbit.NewClient(stellorbit.Options{
		Endpoint: "http://localhost:8080",
		APIKey:   "local-dev-api-key",
	})
	if err != nil {
		log.Fatal(err)
	}

	response, err := client.Route(context.Background(), stellorbit.RouteRequest{
		ServiceName: "payment-service",
		RouteKey:    "tenant-a",
		Attributes: map[string]string{
			"env":    "dev",
			"region": "local",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response.Body)
}
```

## API 表面

| 方法 | 职责 |
| --- | --- |
| `Start(ctx)` | 启动基于 StellNula 的治理规则源 |
| `Rules()` | 返回当前不可变治理规则注册表 |
| `Routes()` | 返回 `RouteRuleProvider` |
| `CircuitBreakers()` | 返回 `CircuitBreakerRuleProvider` |
| `Authorizations()` | 返回 `AuthorizationRuleProvider` |
| `RateLimits()` | 返回 `RateLimitRuleProvider` |
| `Route(ctx, request)` | 向 StellOrbit 请求路由决策 |
| `LifecyclePolicy(ctx, serviceName)` | 获取服务生命周期治理策略 |
| `TrafficPolicy(ctx, serviceName)` | 获取路由、重试和流量迁移策略 |

## 包结构

| 包 | 职责 |
| --- | --- |
| `github.com/stellhub/stellorbit-go-sdk` | 公开 SDK facade、客户端生命周期、options、兼容别名 |
| `github.com/stellhub/stellorbit-go-sdk/governance` | 规则模型、registry、parser、matcher、query、provider、source contract |
| `internal/source/stellnula` | 基于 StellNula 的治理规则源 |
| `internal/httpapi` | 旧版 StellOrbit HTTP API 的兼容客户端 |

## 开发

运行验证：

```bash
go test ./...
```

## 仓库范围

SDK 首版会刻意保持小而清晰。未来版本可以继续补充：

- 强类型策略模型。
- HTTP middleware 和 gRPC interceptor 适配器。
- 重试、故障转移和负载均衡辅助能力。
- 当服务契约稳定后引入 OpenAPI 生成 DTO。
- 针对本地 `stellorbit-service` 的集成测试。
- 面向客户端侧指标和 tracing 的可观测性 hooks。

## License

许可证会在首个稳定版本前确定。
