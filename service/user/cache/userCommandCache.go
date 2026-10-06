package mencache

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
)

type userCommandCache struct {
	store *cache.CacheStore
}

func NewUserCommandCache(store *cache.CacheStore) *userCommandCache {
	return &userCommandCache{store: store}
}

func (u *userCommandCache) DeleteUserCache(ctx context.Context, id int) {
	key := fmt.Sprintf(userByIdCacheKey, id)

	cache.DeleteFromCache(ctx, u.store, key)
}

func (u *userCommandCache) DeleteUserAllCache(ctx context.Context) {
	u.store.InvalidateCache(ctx, "user:*")
}
