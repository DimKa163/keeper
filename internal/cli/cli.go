package cli

import (
	"context"
	"sync"
	"time"

	"github.com/DimKa163/keeper/internal/cli/core"
)

const key = "github.com/DimKa163/keeper_MasterKey"

type CLI struct {
	keys     map[string]any
	context  context.Context
	mu       sync.RWMutex
	services *ServiceContainer
}

func NewCLI(ctx context.Context, service *ServiceContainer,
	masterKey []byte) *CLI {
	keys := make(map[string]any)
	keys[key] = masterKey
	return &CLI{
		context:  ctx,
		keys:     keys,
		services: service,
	}
}

func (c *CLI) Decoder() core.Decoder {
	return c.services.Decoder
}

func (c *CLI) Encoder() core.Encoder {
	return c.services.Encoder
}

func (c *CLI) Console() *Console {
	return c.services.Console
}

func (c *CLI) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.keys == nil {
		c.keys = make(map[string]any)
	}
	c.keys[key] = value
}

func (c *CLI) Get(key string) (value any, exist bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exist = c.keys[key]
	return
}

func (c *CLI) MasterKey() (masterKey []byte, exist bool) {
	var value any
	value, exist = c.Get(key)
	if !exist {
		return nil, false
	}
	masterKey, ok := value.([]byte)
	if !ok {
		panic("masterKey is not []byte")
	}
	return
}
func (c *CLI) Deadline() (deadline time.Time, ok bool) {
	return c.context.Deadline()
}
func (c *CLI) Done() <-chan struct{} {
	return c.context.Done()
}

func (c *CLI) Err() error {
	return c.context.Err()
}

func (c *CLI) Value(key any) any {
	if keyAsString, ok := key.(string); ok {
		if val, exists := c.Get(keyAsString); exists {
			return val
		}
	}
	return c.context.Value(key)
}
