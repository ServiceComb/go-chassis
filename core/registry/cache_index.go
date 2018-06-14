package registry

import (
	"github.com/ServiceComb/go-chassis/core/common"
	"github.com/hashicorp/go-version"
	cache "github.com/patrickmn/go-cache"
	"k8s.io/apimachinery/pkg/util/sets"
)

// noIndexCache return cache without index
type noIndexCache struct {
	latestV map[string]string
	cache   *cache.Cache
}

func newNoIndexCache() *noIndexCache {
	return &noIndexCache{
		cache:   cache.New(DefaultExpireTime, 0),
		latestV: map[string]string{},
	}
}

func (n *noIndexCache) SetIndexTags(tags sets.String) {}
func (n *noIndexCache) Items() map[string]cache.Item  { return n.cache.Items() }
func (n *noIndexCache) Delete(k string)               { n.cache.Delete(k); delete(n.latestV, k) }

func (n *noIndexCache) Set(k string, x interface{}) {
	latestV, _ := version.NewVersion("0.0.0")
	items, ok := x.([]*MicroServiceInstance)
	if !ok {
		return
	}
	for _, item := range items {
		v, _ := version.NewVersion(item.version())
		if v != nil && latestV.LessThan(v) {
			n.latestV[k] = item.version()
			latestV = v
		}
	}
	// TODO: mutex should use
	n.cache.Set(k, x, 0)
}

func (n *noIndexCache) Get(k string, tags map[string]string) (interface{}, bool) {
	value, ok := n.cache.Get(k)
	if len(tags) == 0 || !ok {
		return value, ok
	}
	items, ok := value.([]*MicroServiceInstance)
	if !ok {
		return nil, false
	}

	n.setTagsBeforeQuery(k, tags)
	ret := make([]*MicroServiceInstance, 0, len(items))
	for _, item := range items {
		if item.Has(tags) {
			ret = append(ret, item)
		}
	}
	if len(ret) == 0 {
		return nil, false
	}
	return ret, true
}

func (n *noIndexCache) setTagsBeforeQuery(k string, tags map[string]string) {
	if v, ok := tags[common.BuildinTagVersion]; ok && v == common.LatestVersion {
		tags[common.BuildinTagVersion] = n.latestV[k]
	}
}

// indexCache return cache with index
type indexCache struct {
	cache *noIndexCache
	// index is one of Index implemention
	index Index
}

func newIndexCache() *indexCache {
	tags := sets.NewString(common.BuildinTagVersion, common.BuildinTagApp)
	return &indexCache{
		cache: newNoIndexCache(),
		index: newHashIndex(tags),
	}
}

// TODO: if tags rebuild, indexers should autoclear to remove
// index which is built from old tags
func (b *indexCache) SetIndexTags(tags sets.String) { b.index.SetTags(tags) }
func (b *indexCache) Items() map[string]cache.Item  { return b.cache.Items() }
func (b *indexCache) Delete(k string)               { b.cache.Delete(k); b.index.Delete(k) }

func (b *indexCache) Set(k string, x interface{}) {
	b.cache.Set(k, x)
	// no tags means no index need to be built
	if len(b.index.GetTags()) == 0 {
		return
	}
	b.index.Set(k, x)
}

func (b *indexCache) Get(k string, tags map[string]string) (interface{}, bool) {
	if len(tags) == 0 {
		return b.cache.Get(k, tags)
	}
	// reset version tag if exist
	b.cache.setTagsBeforeQuery(k, tags)
	return b.index.Get(k, tags)
}
