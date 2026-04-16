package database

import (
	"context"

	"github.com/valkey-io/valkey-go"
)

type Valkey struct {
	client valkey.Client
}

func NewValkey(addr string) (*Valkey, error) {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{addr},
	})
	if err != nil {
		return nil, err
	}
	return &Valkey{client: client}, nil
}

func (v *Valkey) Client() valkey.Client {
	return v.client
}

func (v *Valkey) Close() {
	v.client.Close()
}

func (v *Valkey) Set(ctx context.Context, key, value string, expSeconds int64) error {
	cmd := v.client.B().Set().Key(key).Value(value).ExSeconds(expSeconds).Build()
	return v.client.Do(ctx, cmd).Error()
}

func (v *Valkey) Get(ctx context.Context, key string) (string, error) {
	cmd := v.client.B().Get().Key(key).Build()
	return v.client.Do(ctx, cmd).ToString()
}

func (v *Valkey) Del(ctx context.Context, key string) error {
	cmd := v.client.B().Del().Key(key).Build()
	return v.client.Do(ctx, cmd).Error()
}
