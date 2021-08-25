package dataperm

import (
	"context"

	"gorm.io/gorm"
)

type contextKey struct{}

var dataPermApplierHolderKey = contextKey{}

// Applier 是数据权限适配器
type Applier interface {
	Apply(db *gorm.DB) *gorm.DB
	Domain() string
}

type nopApplier struct{}

func (n nopApplier) Apply(db *gorm.DB) *gorm.DB {
	return db
}
func (n nopApplier) Domain() string {
	return "nop"
}

type holder struct {
	m map[string]Applier
}

func newHolder() *holder {
	m := make(map[string]Applier, 0)
	return &holder{
		m: m,
	}
}
func (h *holder) Set(domain string, applier Applier) {
	h.m[domain] = applier
}
func (h holder) Get(domain string) (Applier, bool) {
	a, ok := h.m[domain]
	return a, ok
}

// Wrap 根据ctx和domain包装出一个可用的Applier
func Wrap(ctx context.Context, domain string) Applier {
	if h, ok := ctx.Value(dataPermApplierHolderKey).(*holder); ok {
		if applier, ok := h.Get(domain); ok {
			return applier
		}
	}
	return &nopApplier{}
}

// SetAppliers 设置Applier到Context
func SetAppliers(ctx context.Context, appliers []Applier) context.Context {
	h := newHolder()
	for _, applier := range appliers {
		h.Set(applier.Domain(), applier)
	}
	return context.WithValue(ctx, dataPermApplierHolderKey, h)
}
