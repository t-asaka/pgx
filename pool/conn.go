package pool

import (
	"context"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
)

// Conn is an acquired *pgx.Conn from a Pool.
type Conn struct {
	res *puddle.Resource
}

// Release returns c to the pool it was acquired from. Once Release has been called, other methods must not be called.
// However, it is safe to call Release multiple times. Subsequent calls after the first will be ignored.
func (c *Conn) Release() {
	if c.res == nil {
		return
	}

	conn := c.Conn()
	res := c.res
	c.res = nil

	go func() {
		if !conn.IsAlive() {
			res.Destroy()
			return
		}

		if conn.PgConn().TxStatus != 'I' {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, err := conn.Exec(ctx, "rollback")
			cancel()
			if err != nil {
				res.Destroy()
				return
			}
		}

		if conn.IsAlive() {
			res.Release()
		} else {
			res.Destroy()
		}
	}()
}

func (c *Conn) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return c.Conn().Exec(ctx, sql, arguments...)
}

func (c *Conn) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return c.Conn().Query(ctx, sql, args...)
}

func (c *Conn) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return c.Conn().QueryRow(ctx, sql, args...)
}

func (c *Conn) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return c.Conn().SendBatch(ctx, b)
}

func (c *Conn) Begin(ctx context.Context, txOptions *pgx.TxOptions) (*pgx.Tx, error) {
	return c.Conn().Begin(ctx, txOptions)
}

func (c *Conn) Conn() *pgx.Conn {
	return c.res.Value().(*pgx.Conn)
}
