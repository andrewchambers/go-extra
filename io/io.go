package io

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type MergedReadWriteCloser struct {
	RC io.ReadCloser
	WC io.WriteCloser
}

func (m *MergedReadWriteCloser) Read(buf []byte) (int, error) {
	return m.RC.Read(buf)
}

func (m *MergedReadWriteCloser) Write(buf []byte) (int, error) {
	return m.WC.Write(buf)
}

func (m *MergedReadWriteCloser) Close() error {
	_ = m.RC.Close()
	_ = m.WC.Close()
	return nil
}

func ConnectedPipes() (*MergedReadWriteCloser, *MergedReadWriteCloser) {
	a, b := io.Pipe()
	x, y := io.Pipe()

	return &MergedReadWriteCloser{
			RC: a,
			WC: y,
		}, &MergedReadWriteCloser{
			RC: x,
			WC: b,
		}
}

type MeteredConn struct {
	Conn net.Conn
	// if accessed concurrently, Read with sync/atomic
	ReadCount int64
	// if accessed concurrently, Read with sync/atomic
	WriteCount int64
}

func NewMeteredConn(c net.Conn) *MeteredConn {
	return &MeteredConn{
		Conn: c,
	}
}

func (mConn *MeteredConn) Read(buf []byte) (int, error) {
	n, err := mConn.Conn.Read(buf)
	atomic.AddInt64(&mConn.ReadCount, int64(n))
	return n, err
}

func (mConn *MeteredConn) Write(buf []byte) (int, error) {
	n, err := mConn.Conn.Write(buf)
	atomic.AddInt64(&mConn.WriteCount, int64(n))
	return n, err
}

func (mConn *MeteredConn) Close() error {
	return mConn.Conn.Close()
}

func (mConn *MeteredConn) LocalAddr() net.Addr {
	return mConn.Conn.LocalAddr()
}

func (mConn *MeteredConn) RemoteAddr() net.Addr {
	return mConn.Conn.RemoteAddr()
}

func (mConn *MeteredConn) SetDeadline(t time.Time) error {
	return mConn.Conn.SetDeadline(t)
}

func (mConn *MeteredConn) SetReadDeadline(t time.Time) error {
	return mConn.Conn.SetReadDeadline(t)
}

func (mConn *MeteredConn) SetWriteDeadline(t time.Time) error {
	return mConn.Conn.SetWriteDeadline(t)
}

type MeteredWriter struct {
	W          io.Writer
	WriteCount int64
}

func (mw *MeteredWriter) Write(buf []byte) (int, error) {
	n, err := mw.W.Write(buf)
	atomic.AddInt64(&mw.WriteCount, int64(n))
	return n, err
}

type MeteredReader struct {
	R         io.Reader
	ReadCount int64
}

func (mw *MeteredReader) Read(buf []byte) (int, error) {
	n, err := mw.R.Read(buf)
	atomic.AddInt64(&mw.ReadCount, int64(n))
	return n, err
}

var ErrOutOfSpace error = errors.New("Out of space")

// Write to W until the given buffer is larger then N.
// In that case return ErrOutOfSpace
// Multiple writers are safe, but
// don't access N concurrently with writing.
type LimitedWriter struct {
	lock sync.Mutex
	N    int64
	W    io.Writer
}

func (w *LimitedWriter) Write(buf []byte) (int, error) {
	w.lock.Lock()
	if int64(len(buf)) > w.N {
		w.lock.Unlock()
		return 0, ErrOutOfSpace
	}
	w.N -= int64(len(buf))
	w.lock.Unlock()

	return w.W.Write(buf)
}
