package buffer

import (
	"context"

	"github.com/ClickHouse/ch-go"
	"github.com/ClickHouse/ch-go/chpool"
	"github.com/ClickHouse/ch-go/proto"
)

var _ Container[ClickHouseRow] = &ClickHouseContainer{}

type ClickHouseRow interface {
	Insert(proto.Input) error
}

type ClickHouseContainer struct {
	pool     *chpool.Pool
	Table    string
	size     int
	BulkSize int
	cols     proto.Input

	newInputFunc func() proto.Input
}

func NewClickHouseContainer(pool *chpool.Pool, table string, bulkSize int, newInputFunc func() proto.Input) (*ClickHouseContainer, error) {
	return &ClickHouseContainer{
		pool:     pool,
		Table:    table,
		size:     0,
		BulkSize: bulkSize,
		cols:     newInputFunc(),

		newInputFunc: newInputFunc,
	}, nil
}

func (container *ClickHouseContainer) put(element ClickHouseRow) error {
	if err := element.Insert(container.cols); err != nil {
		return err
	}
	container.size++
	return nil
}

func (container *ClickHouseContainer) flush() error {
	if container.size == 0 {
		return nil
	}

	dataToWrite := container.cols
	container.cols = container.newInputFunc()
	container.size = 0

	// TODO 异步插入如何确保失败时重试、恢复？
	if err := container.pool.Do(context.TODO(), ch.Query{
		Body:  dataToWrite.Into(container.Table),
		Input: dataToWrite,
	}); err != nil {
		return err
	}

	return nil
}

func (container *ClickHouseContainer) isFull() bool {
	if container.size == container.BulkSize {
		return true
	}
	return false
}

func (container *ClickHouseContainer) reset() {
	// TODO 如何处理cols中残留数据?
	container.size = 0
	container.cols.Reset()
}
