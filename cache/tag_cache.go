package cache

type TagCache struct {
	store  CacheStore
	tagSet *TagSet
}

var _ Cache = new(TagCache)

func NewTagCache(store CacheStore, names ...string) Cache {
	return &TagCache{store: store, tagSet: NewTagSet(store, names)}
}

func (this *TagCache) TaggedItemKey(key string) string {
	return EncodeSha1(this.tagSet.GetNamespace()) + ":" + key
}

func (this *TagCache) Get(key string, _val interface{}) error {
	return this.store.Get(this.TaggedItemKey(key), _val)
}

func (this *TagCache) IsExist(key string) bool {
	return this.store.IsExist(this.TaggedItemKey(key))
}

// 更新过期时间
func (this *TagCache) Touch(key string, expire int64) error {
	return this.store.Touch(this.TaggedItemKey(key), expire)
}

func (this *TagCache) Set(key string, value interface{}, expire int64) error {
	return this.store.Set(this.TaggedItemKey(key), value, expire)
}

func (this *TagCache) Incr(key string) (int64, error) {
	return this.store.Incr(this.TaggedItemKey(key))
}

func (this *TagCache) Decr(key string) (int64, error) {
	return this.store.Decr(this.TaggedItemKey(key))
}

func (this *TagCache) Delete(key string) error {
	return this.store.Delete(this.TaggedItemKey(key))
}

func (this *TagCache) Flush() error {
	return this.tagSet.Reset()
}

// add Tags
func (this *TagCache) Tags(tags []string) Cache {
	this.tagSet.AddNames(tags)
	return this
}

func (this *TagCache) StartAndGC(opt Options) error {
	return this.store.StartAndGC(opt)
}
