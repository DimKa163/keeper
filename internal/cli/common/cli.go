package common

import (
	"context"
	"errors"
)

const stateName = "recordState"

const (
	key           = "github.com/DimKa163/keeper_MasterKey"
	hostName      = "github.com/DimKa163/keeper_Hostname"
	userClientKey = "github.com/DimKa163/keeper_UsersClient"
	dataClientKey = "github.com/DimKa163/keeper_SyncDataClient"

	versionKey = "github.com/DimKa163/keeper_version"
)

var (
	ErrMasterKeyNotRegistered = errors.New("keeper: master key not registered")
	ErrServerNotRegistered    = errors.New("keeper: server not registered")
)

func SetHostName(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, hostName, value)
}

func SetMasterKey(ctx context.Context, masterKey []byte) context.Context {
	return context.WithValue(ctx, key, masterKey)
}

func GetMasterKey(ctx context.Context) ([]byte, error) {
	if v, ok := ctx.Value(key).([]byte); ok {
		return v, nil
	}
	return nil, ErrMasterKeyNotRegistered
}

func SetVersion(ctx context.Context, version int32) context.Context {
	return context.WithValue(ctx, versionKey, version)
}

func GetVersion(ctx context.Context) int32 {
	if v, ok := ctx.Value(versionKey).(int32); ok {
		return v
	}
	return -1
}

//type CLI struct {
//	context  context.Context
//	keys     map[string]any
//	mu       *sync.RWMutex
//	services *ServiceContainer
//}
//
//func NewCLI(ctx context.Context, service *ServiceContainer, hostname string,
//	masterKey []byte) *CLI {
//	keys := make(map[string]any)
//	keys[key] = masterKey
//	keys[hostName] = hostname
//	return &CLI{
//		context:  ctx,
//		keys:     keys,
//		services: service,
//		mu:       &sync.RWMutex{},
//	}
//}
//
//func (c *CLI) Decoder() core.Decoder {
//	return c.services.Decoder
//}
//
//func (c *CLI) Encoder() core.Encoder {
//	return c.services.Encoder
//}
//
//func (c *CLI) Console() *ConsoleLine {
//	return c.services.Console
//}
//
//func (c *CLI) GetVersion(db *sql.Tx) (int32, error) {
//	state, err := persistence.TxGetState(c, db, stateName)
//	if err != nil {
//		if !errors.Is(sql.ErrNoRows, err) {
//			return -1, err
//		}
//		state = &core.SyncState{
//			ID:    stateName,
//			Value: 0,
//		}
//	}
//	return state.Value, nil
//}
//
//func (c *CLI) GetRpcClient() (pb.SyncDataClient, error) {
//	val, exists := c.Get(dataClientKey)
//	if exists {
//		return val.(pb.SyncDataClient), nil
//	}
//	srv, err := persistence.GetServer(c, c.services.DB, true)
//	if err != nil {
//		if errors.Is(sql.ErrNoRows, err) {
//			return nil, ErrServerNotRegistered
//		}
//		return nil, err
//	}
//	conn, err := grpc.NewClient(srv.Address,
//		grpc.WithTransportCredentials(insecure.NewCredentials()))
//	if err != nil {
//		return nil, err
//	}
//	users := pb.NewUsersClient(conn)
//	c.Set(userClientKey, users)
//	login := c.keys[hostName].(string)
//	masterKey, _ := c.MasterKey()
//	interceptor := newInterceptor(users, login, string(masterKey))
//	protectedConn, err := grpc.NewClient(srv.Address,
//		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithChainUnaryInterceptor(interceptor.Handle()))
//	if err != nil {
//		return nil, err
//	}
//	dataRpcClient := pb.NewSyncDataClient(protectedConn)
//	c.Set(dataClientKey, dataRpcClient)
//	return dataRpcClient, nil
//}
//
//func (c *CLI) Set(key string, value any) {
//	c.mu.Lock()
//	defer c.mu.Unlock()
//	if c.keys == nil {
//		c.keys = make(map[string]any)
//	}
//	c.keys[key] = value
//}
//
//func (c *CLI) Get(key string) (value any, exist bool) {
//	c.mu.RLock()
//	defer c.mu.RUnlock()
//	value, exist = c.keys[key]
//	return
//}
//
//func (c *CLI) MasterKey() (masterKey []byte, exist bool) {
//	var value any
//	value, exist = c.Get(key)
//	if !exist {
//		return nil, false
//	}
//	masterKey, ok := value.([]byte)
//	if !ok {
//		panic("masterKey is not []byte")
//	}
//	return
//}
//func (c *CLI) Deadline() (deadline time.Time, ok bool) {
//	return c.context.Deadline()
//}
//func (c *CLI) Done() <-chan struct{} {
//	return c.context.Done()
//}
//
//func (c *CLI) Err() error {
//	return c.context.Err()
//}
//
//func (c *CLI) Value(key any) any {
//	if keyAsString, ok := key.(string); ok {
//		if val, exists := c.Get(keyAsString); exists {
//			return val
//		}
//	}
//	return c.context.Value(key)
//}
//
//type unaryIdentifyInterceptor struct {
//	users    pb.UsersClient
//	token    string
//	username string
//	userpass string
//}
//
//func newInterceptor(users pb.UsersClient, username, userpass string) *unaryIdentifyInterceptor {
//	return &unaryIdentifyInterceptor{
//		users:    users,
//		username: username,
//		userpass: userpass,
//	}
//}
//
//func (h *unaryIdentifyInterceptor) Handle() grpc.UnaryClientInterceptor {
//	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
//		var err error
//		if h.token == "" {
//			h.token, err = h.login(ctx)
//			if err != nil {
//				return err
//			}
//		}
//		md := metadata.New(map[string]string{"authorization": fmt.Sprintf("Bearer %s", h.token)})
//		err = invoker(metadata.NewOutgoingContext(ctx, md), method, req, reply, cc, opts...)
//		if err != nil {
//			if e, ok := status.FromError(err); ok {
//				if e.Code() == codes.Unauthenticated {
//					h.token, err = h.login(ctx)
//					//md = metadata.New(map[string]string{"authorization": fmt.Sprintf("Bearer %s", h.token)})
//					//ctx = metadata.NewOutgoingContext(ctx, md)
//					err = invoker(ctx, method, req, reply, cc, opts...)
//				}
//			}
//		}
//		return err
//	}
//}
//
//func (h *unaryIdentifyInterceptor) login(ctx context.Context) (string, error) {
//	var us pb.User
//	us.SetLogin(h.username)
//	us.SetPassword(h.userpass)
//	t, err := h.users.Login(ctx, &us)
//	if err != nil {
//		return "", err
//	}
//	return t.GetToken(), nil
//}
