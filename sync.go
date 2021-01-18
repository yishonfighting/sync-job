package sync_job

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	inited     bool
	m          sync.RWMutex
	ComRoutine *Routine
)

// 错误提示
var (
	ChannelFullErr = errors.New("通道已满")
)

// 操作参数设置
type option struct {
	worker int
	buffer int
}

// 供外部自定义接口
type Option func(*option)

// 设置操作数
func SetWorker(n int) Option {
	if n < 1 {
		panic("routine: worker should > 0")
	}

	return func(o *option) {
		o.worker = n
	}
}

// 设置缓冲数
func SetBuffer(n int) Option {
	if n < 1 {
		panic("routine: buffer should > 0")
	}

	return func(o *option) {
		o.buffer = n
	}
}

// 通道方法类型
type item struct {
	funcName string
	f        func(c context.Context)
	ctx      context.Context
}

// 通道结构体
type Routine struct {
	name   string
	ch     chan item
	option *option
	waiter sync.WaitGroup

	ctx   context.Context
	close func()
}

// 初始化全局channel
func InitRoutine(name string, opts ...Option) *Routine {
	if inited {
		return ComRoutine
	}

	m.Lock()
	defer m.Unlock()

	ComRoutine = newRoutine(name, opts...)

	inited = true
	return ComRoutine
}

// 获取全局channel
func GetRoutine() *Routine {
	if inited {
		return ComRoutine
	}

	return InitRoutine("routine", SetWorker(5), SetBuffer(100))
}

// 生成channel
func newRoutine(name string, opts ...Option) *Routine {
	if name == "" {
		name = "routine"
	}

	// 设置操作和缓冲的数量
	o := &option{
		worker: 1,
		buffer: 1024,
	}
	for _, opt := range opts {
		opt(o)
	}

	r := &Routine{
		name:   name,
		ch:     make(chan item, o.buffer),
		option: o,
	}
	r.ctx, r.close = context.WithCancel(context.Background())
	r.waiter.Add(o.worker)
	for i := 0; i < o.worker; i++ {
		go r.process()
	}
	return r
}

// 接收信息处理
func (r *Routine) process() {
	defer r.waiter.Done()
	for {
		select {
		case im := <-r.ch:
			wrapFunc(im.funcName, im.f)(im.ctx)
		case <-r.ctx.Done():
			return
		}
	}
}

// 封装方法
func wrapFunc(funcName string, f func(c context.Context)) (res func(context.Context)) {
	res = func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 日志
				// 监控
			}
		}()

		// 方法执行
		f(ctx)
	}
	return
}

// 执行方法
func (r *Routine) Do(c context.Context, funcName string, f func(c context.Context)) (err error) {
	if f == nil || r.ctx.Err() != nil {
		return r.ctx.Err()
	}

	select {
	case r.ch <- item{funcName: funcName, f: f, ctx: c}:
	default:
		err = ChannelFullErr
	}

	return
}

// 关闭
func (r *Routine) Close() error {
	if err := r.ctx.Err(); err != nil {
		return err
	}
	r.close()
	r.waiter.Wait()
	fmt.Println(fmt.Sprintf("Shutdown %s successfully", r.name))
	return nil
}

