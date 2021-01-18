### sync-job

* 系统统一定义、管理goroutine

* 统一内存管理工具

* 支持定义Worker 数量的goroutine，进行异步处理

示例:
```golang
sync_job.GetRoutine().Do(c, "funcName", func(ctx context.Context) { SomeFunc(c, args...) })
```