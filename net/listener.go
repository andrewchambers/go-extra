package net

import (
	"context"
	"net"
	"sync/atomic"
	"time"

	"github.com/andrewchambers/go-extra/errors"
	"golang.org/x/sync/semaphore"
)

type ConnectionManagerOptions struct {
	Domain        string
	Addr          string
	MaxConcurrent int64
}

func (l *ConnectionManagerOptions) Sanitize() {
	if l.Domain == "" {
		l.Domain = "tcp"
	}

	if l.MaxConcurrent <= 0 {
		l.MaxConcurrent = 1
	}
}

func NewConnectionManager(options ConnectionManagerOptions) (*ConnectionManager, error) {
	options.Sanitize()

	l, err := net.Listen(options.Domain, options.Addr)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return &ConnectionManager{
		connectionLimit: semaphore.NewWeighted(options.MaxConcurrent),
		l:               l,
	}, nil
}

// Like net.Listener, but supporting an upper bound
// on concurrent connections. The returned connections
// also track read/write traffic.
type ConnectionManager struct {
	connectionLimit *semaphore.Weighted
	l               net.Listener
}

func (cm *ConnectionManager) Accept() (net.Conn, error) {

	err := cm.connectionLimit.Acquire(context.TODO(), 1)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	conn, err := cm.l.Accept()
	if err != nil {
		cm.connectionLimit.Release(1)
		return nil, errors.Wrap(err)
	}

	return &Connection{
		manager: cm,
		Conn:    conn,
	}, nil
}

func (cm *ConnectionManager) Release(c *Connection) {
	cm.connectionLimit.Release(1)
}

func (cm *ConnectionManager) Close() error {
	return cm.l.Close()
}

func (cm *ConnectionManager) Addr() net.Addr {
	return cm.l.Addr()
}

type Connection struct {
	manager    *ConnectionManager
	Conn       net.Conn
	readCount  uint64
	writeCount uint64
}

// Safe for concurrent use
func (conn *Connection) ReadCount() uint64 {
	return atomic.LoadUint64(&conn.readCount)
}

// Safe for concurrent use
func (conn *Connection) WriteCount() uint64 {
	return atomic.LoadUint64(&conn.writeCount)
}

func (conn *Connection) Read(buf []byte) (int, error) {
	n, err := conn.Conn.Read(buf)
	atomic.AddUint64(&conn.readCount, uint64(n))
	return n, err
}

func (conn *Connection) Write(buf []byte) (int, error) {
	n, err := conn.Conn.Write(buf)
	atomic.AddUint64(&conn.writeCount, uint64(n))
	return n, err
}

func (conn *Connection) Close() error {
	conn.manager.Release(conn)
	return conn.Conn.Close()
}

func (conn *Connection) LocalAddr() net.Addr {
	return conn.Conn.LocalAddr()
}

func (conn *Connection) RemoteAddr() net.Addr {
	return conn.Conn.RemoteAddr()
}

func (conn *Connection) SetDeadline(t time.Time) error {
	return conn.Conn.SetDeadline(t)
}

func (conn *Connection) SetReadDeadline(t time.Time) error {
	return conn.Conn.SetReadDeadline(t)
}

func (conn *Connection) SetWriteDeadline(t time.Time) error {
	return conn.Conn.SetWriteDeadline(t)
}
