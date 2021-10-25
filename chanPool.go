package pool

import (
	"fmt"
)

type Pool struct {
	factory factoryFunc
	destory destoryFunc
	conns   chan interface{}
	closed  bool
}
type factoryFunc func() interface{}
type destoryFunc func(interface{}) error

//初始化并返回一个空pool
func NewPool(cap int, fa factoryFunc, de destoryFunc) *Pool {
	return &Pool{
		factory: fa,
		destory: de,
		conns:   make(chan interface{}, cap),
		closed: false,
	}
}

//Release关闭pool，并释放pool中的连接，此时Get和Put操作无效
func (p *Pool) Release() error {
	if p.closed {
		return nil
	}
	p.closed = true
	close(p.conns)
	lenth := len(p.conns)
	for i := 0; i < lenth; i++ {
		conn := <-p.conns
		if err := p.destory(conn); err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

//Get获取一个连接
//若pool是开启状态，则从pool中获取一个连接，若此时pool是空的，则生成一个新的连接
//若pool是关闭状态，则Get总是会返回nil
func (p *Pool) Get() interface{} {
	if p.closed {
		return nil
	}
	select {
	case conn := <-p.conns:
		return conn
	default:
		return p.factory()
	}
}

//Put归还一个连接
//若pool是开启状态，则归还一个连接到pool，若此时pool是满的，则释放这个连接
//若pool是关闭状态，则释放这个连接
func (p *Pool) Put(conn interface{}) (err error) {
	// defer func() {
	// 	if perr := recover(); perr != nil {
	// 		err = fmt.Errorf("pool:%v", perr)
	// 	}
	// }()
	if conn == nil {
		return nil
	}
	if p.closed {
		return p.destory(conn)
	}
	select {
	case p.conns <- conn:
		return nil
	default:
		return p.destory(conn)
	}
}

func (p *Pool) Len() int {
	return len(p.conns)
}

func (p *Pool) Cap() int {
	return cap(p.conns)
}
