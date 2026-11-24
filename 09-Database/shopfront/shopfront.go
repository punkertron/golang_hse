//go:build !solution

package shopfront

import (
	"context"
	"strconv"

	"github.com/go-redis/redis/v8"
)

type MyCounters struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) Counters {
	return &MyCounters{rdb: rdb}
}

func (c *MyCounters) GetItems(ctx context.Context, ids []ItemID, userID UserID) ([]Item, error) {
	type itemCommands struct {
		count  *redis.StringCmd
		viewed *redis.BoolCmd
	}

	pipe := c.rdb.Pipeline()

	commands := make([]itemCommands, len(ids))
	for i, id := range ids {
		key := itemKey(id)
		commands[i] = itemCommands{
			count:  pipe.Get(ctx, key+":views"),
			viewed: pipe.SIsMember(ctx, key+":viewers", int64(userID)),
		}
	}

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}

	items := make([]Item, len(ids))

	for i, cmd := range commands {
		count, err := cmd.count.Int()
		if err != nil && err != redis.Nil {
			return nil, err
		}

		viewed, err := cmd.viewed.Result()
		if err != nil {
			return nil, err
		}

		items[i] = Item{
			ViewCount: count,
			Viewed:    viewed,
		}
	}

	return items, nil
}

func (c *MyCounters) RecordView(ctx context.Context, id ItemID, userID UserID) error {
	key := itemKey(id)
	pipe := c.rdb.TxPipeline()
	pipe.Incr(ctx, key+":views")
	pipe.SAdd(ctx, key+":viewers", int64(userID))
	_, err := pipe.Exec(ctx)

	return err
}

func itemKey(id ItemID) string {
	return "shopfront:item:" + strconv.FormatInt(int64(id), 10)
}
