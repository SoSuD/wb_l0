package usecase

import (
	"context"
	"errors"
	"testing"

	"wb_l0/internal/orders/cache"
	"wb_l0/internal/orders/mocks"
	"wb_l0/models"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- no-op логгер под интерфейс logger.Logger ---\
type nopLogger struct{}

func (l nopLogger) DPanic(args ...interface{}) {
	panic("")
}

func (l nopLogger) DPanicf(template string, args ...interface{}) {
	panic("")
}

func (nopLogger) InitLogger()                   {}
func (nopLogger) Debug(args ...interface{})     {}
func (nopLogger) Debugf(string, ...interface{}) {}
func (nopLogger) Info(args ...interface{})      {}
func (nopLogger) Infof(string, ...interface{})  {}
func (nopLogger) Warn(args ...interface{})      {}
func (nopLogger) Warnf(string, ...interface{})  {}
func (nopLogger) Error(args ...interface{})     {}
func (nopLogger) Errorf(string, ...interface{}) {}
func (nopLogger) Panic(args ...interface{})     {}
func (nopLogger) Panicf(string, ...interface{}) {}
func (nopLogger) Fatal(args ...interface{})     {}
func (nopLogger) Fatalf(string, ...interface{}) {}

func makeOrder(id string) *models.Order {
	return &models.Order{
		OrderUid:        id,
		TrackNumber:     "TRK-" + id,
		Entry:           "WBIL",
		CustomerId:      "cust",
		DeliveryService: "svc",
		Shardkey:        "1",
		OofShard:        "1",
		Payment: models.Payment{
			Transaction: "tx-" + id,
			Currency:    "USD",
			Amount:      100,
		},
		Delivery: models.Delivery{
			Name: "Alice", City: "Moscow", Address: "Street 1",
		},
		Items: []models.Item{{ChrtId: 1, TrackNumber: "TRK-" + id, Price: 100, Name: "item"}},
	}
}

type bundle struct {
	repo  *mocks.Repository
	cache *cache.LRU[string, *models.Order]
	uc    *OrdersUC
}

func newBundle(t *testing.T) bundle {
	repo := mocks.NewRepository(t)
	c := cache.New[string, *models.Order](128, nil)
	u := NewOrdersUC(nopLogger{}, repo, c)
	return bundle{repo: repo, cache: c, uc: u}
}

// ------------------- Create -------------------

func TestCreate_OK_Caches(t *testing.T) {
	b := newBundle(t)
	ctx := context.Background()
	o := makeOrder("ord-1")

	b.repo.EXPECT().Create(mock.Anything, o).Return(nil).Once()

	err := b.uc.Create(ctx, o)
	require.NoError(t, err)

	got, ok := b.cache.Get("ord-1")
	require.True(t, ok)
	require.Equal(t, o, got)
}

func TestCreate_RepoError_NoCache(t *testing.T) {
	b := newBundle(t)
	ctx := context.Background()
	o := makeOrder("ord-2")

	b.repo.EXPECT().Create(mock.Anything, o).Return(errors.New("db fail")).Once()

	err := b.uc.Create(ctx, o)
	require.Error(t, err)

	_, ok := b.cache.Get("ord-2")
	require.False(t, ok)
}

// ------------------- GetByID -------------------

func TestGetByID_CacheHit(t *testing.T) {
	b := newBundle(t)
	ctx := context.Background()
	o := makeOrder("ord-3")
	b.cache.Put(o.OrderUid, o)

	// репо вызываться не должен (если включен expecter, мок упадёт при лишнем вызове)
	got, err := b.uc.GetByID(ctx, o.OrderUid)
	require.NoError(t, err)
	require.Equal(t, o, got)
}

func TestGetByID_Miss_RepoOK_CacheSet(t *testing.T) {
	b := newBundle(t)
	ctx := context.Background()
	o := makeOrder("ord-4")

	b.repo.EXPECT().GetByID(mock.Anything, o.OrderUid).Return(o, nil).Once()
	// b.repo.On("GetByID", mock.Anything, o.OrderUid).Return(o, nil).Once()

	got, err := b.uc.GetByID(ctx, o.OrderUid)
	require.NoError(t, err)
	require.Equal(t, o, got)

	cached, ok := b.cache.Get(o.OrderUid)
	require.True(t, ok)
	require.Equal(t, o, cached)
}

func TestGetByID_Miss_RepoError(t *testing.T) {
	b := newBundle(t)
	ctx := context.Background()

	b.repo.EXPECT().GetByID(mock.Anything, "missing").Return((*models.Order)(nil), errors.New("not found")).Once()
	// b.repo.On("GetByID", mock.Anything, "missing").Return((*models.Order)(nil), errors.New("not found")).Once()

	got, err := b.uc.GetByID(ctx, "missing")
	require.Error(t, err)
	require.Nil(t, got)

	_, ok := b.cache.Get("missing")
	require.False(t, ok)
}

// ------------------- PutLastCache -------------------

func TestPutLastCache_OK(t *testing.T) {
	b := newBundle(t)
	ctx := context.Background()
	o1 := makeOrder("ord-5")
	o2 := makeOrder("ord-6")

	b.repo.EXPECT().GetLastByCount(mock.Anything, 2).Return([]*models.Order{o1, o2}, nil).Once()
	// b.repo.On("GetLastByCount", mock.Anything, 2).Return([]*models.Order{o1, o2}, nil).Once()

	err := b.uc.PutLastCache(ctx, 2)
	require.NoError(t, err)

	g1, ok1 := b.cache.Get(o1.OrderUid)
	g2, ok2 := b.cache.Get(o2.OrderUid)
	require.True(t, ok1 && ok2)
	require.Equal(t, o1, g1)
	require.Equal(t, o2, g2)
}

func TestPutLastCache_RepoError(t *testing.T) {
	b := newBundle(t)
	ctx := context.Background()

	b.repo.EXPECT().GetLastByCount(mock.Anything, 10).Return(nil, errors.New("db down")).Once()
	// b.repo.On("GetLastByCount", mock.Anything, 10).Return(nil, errors.New("db down")).Once()

	err := b.uc.PutLastCache(ctx, 10)
	require.Error(t, err)
}
