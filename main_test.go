package gocache

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testData struct {
	message    string
	expiration time.Duration
	items      []Item
	expected   expected
}

type expected struct {
	items     []Item
	ok        bool
	itemCount int
}

type key struct {
	id    string
	value int
}

func (k key) Equals(keyComp Key) bool {
	if keyComp == nil {
		return false
	}

	return k.id == keyComp.(key).id
}

func TestSetItem(t *testing.T) {
	testData := []testData{
		{
			message: "Should add two items in cache for cache with no expiration",
			items: []Item{
				{Key: key{id: "first id", value: 1}, Value: "val 1"},
				{Key: key{id: "second id", value: 2}, Value: "val 2"},
			},
			expected: expected{
				items: []Item{
					{Key: key{id: "first id", value: 1}, Value: "val 1"},
					{Key: key{id: "second id", value: 2}, Value: "val 2"},
				},
				itemCount: 2,
			},
		},
		{
			message: "Should replace existing item when duplicate for cache with no expiration",
			items: []Item{
				{Key: key{id: "first id", value: 1}, Value: "val 1"},
				{Key: key{id: "second id", value: 2}, Value: "val 2"},
				{Key: key{id: "first id", value: 3}, Value: "val 3"},
			},
			expected: expected{
				items: []Item{
					{Key: key{id: "first id", value: 1}, Value: "val 3"},
					{Key: key{id: "second id", value: 2}, Value: "val 2"},
				},
				itemCount: 2,
			},
		},
	}

	for _, td := range testData {
		t.Run(td.message, func(t *testing.T) {
			c := New(td.expiration)

			for _, item := range td.items {
				c.Set(item.Key, item.Value)
			}

			assert.Equal(t, td.expected.itemCount, c.ItemCount())
			for _, item := range c.items {
				assert.Contains(t, td.expected.items, *item)
			}
		})
	}
}

func TestSet1000DifferentItemsShouldGet1000ItemsInCache(t *testing.T) {
	c := New(0)
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		id := strconv.Itoa(i)
		go func() {
			defer wg.Done()
			item := &Item{
				Key:   key{id: id, value: 1},
				Value: "value",
			}
			c.Set(item.Key, item.Value)
		}()
	}

	wg.Wait()

	assert.Equal(t, 1000, c.ItemCount())
}

func TestSet1000EqualItemsShouldGetOneItemInCache(t *testing.T) {
	c := New(0)
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			item := &Item{
				Key:   key{id: "id", value: 1},
				Value: "value",
			}
			c.Set(item.Key, item.Value)
		}()
	}

	wg.Wait()

	assert.Equal(t, 1, c.ItemCount())
}

func TestSetItemWithLazyEviction(t *testing.T) {
	testData := []testData{
		{
			message:    "Should update expiration and value on existing item for cache with expiration of 1m",
			expiration: 1 * time.Minute,
			items: []Item{
				{Key: key{id: "first id", value: 1}, Value: "val 1"},
				{Key: key{id: "first id", value: 2}, Value: "val 3"},
			},
			expected: expected{
				items: []Item{
					{Key: key{id: "first id", value: 1}, Value: "val 3"},
				},
				itemCount: 1,
			},
		},
		{
			message:    "Should evict old item and add new one for cache with expiration of 1ns",
			expiration: 1 * time.Nanosecond,
			items: []Item{
				{Key: key{id: "first id", value: 1}, Value: "val 1"},
				{Key: key{id: "first id", value: 2}, Value: "val 3"},
			},
			expected: expected{
				items: []Item{
					{Key: key{id: "first id", value: 2}, Value: "val 3"},
				},
				itemCount: 1,
			},
		},
	}

	for _, td := range testData {
		t.Run(td.message, func(t *testing.T) {
			c := New(td.expiration)

			for _, item := range td.items {
				c.Set(item.Key, item.Value)
				time.Sleep(1 * time.Nanosecond)
			}

			assert.Equal(t, td.expected.itemCount, c.ItemCount())
			assert.Equal(t, td.expected.items[0].Key, c.items[0].Key)
			assert.Equal(t, td.expected.items[0].Value, c.items[0].Value)
			assert.NotNil(t, c.items[0].expireAt)
		})
	}
}

type testDataWithExistingCache struct {
	message  string
	cache    Cache
	expected expected
}

func TestGetItem(t *testing.T) {
	testData := []testDataWithExistingCache{
		{
			message: "Should get item that exists in cache",
			cache: Cache{
				&cache{
					items: []*Item{
						{Key: key{id: "first id", value: 1}, Value: "val 1"},
						{Key: key{id: "second id", value: 2}, Value: "val 2"},
					},
				},
			},
			expected: expected{
				items: []Item{
					{Key: key{id: "first id", value: 1}, Value: "val 1"},
				},
				ok: true,
			},
		},
		{
			message: "Should get not ok when item does not exist in empty cache",
			cache:   Cache{&cache{}},
			expected: expected{
				items: []Item{
					{Key: key{id: "item does not exist", value: 1}, Value: "val 3"},
				},
				ok: false,
			},
		},
		{
			message: "Should get not ok when item does not exist in cache",
			cache: Cache{
				&cache{
					items: []*Item{
						{Key: key{id: "first different id", value: 1}, Value: "val 1"},
					},
				},
			},
			expected: expected{
				items: []Item{
					{Key: key{id: "item does not exist", value: 1}, Value: "val 3"},
				},
				ok: false,
			},
		},
	}

	for _, td := range testData {
		t.Run(td.message, func(t *testing.T) {
			keyToGet := key{
				id:    "first id",
				value: 0,
			}

			val, ok := td.cache.Get(keyToGet)
			assert.Equal(t, td.expected.ok, ok)
			if ok {
				assert.Equal(t, td.expected.items[0].Value, val.Value)
			}
		})
	}
}

func TestGetAllItems(t *testing.T) {
	testData := []testDataWithExistingCache{
		{
			message: "Should get all items that exist in cache",
			cache: Cache{
				&cache{
					items: []*Item{
						{Key: key{id: "first id", value: 1}, Value: "val 1"},
						{Key: key{id: "second id", value: 2}, Value: "val 2"},
					},
				},
			},
			expected: expected{
				items: []Item{
					{Key: key{id: "first id", value: 1}, Value: "val 1"},
					{Key: key{id: "second id", value: 2}, Value: "val 2"},
				},
				itemCount: 2,
			},
		},
		{
			message: "Should get empty items list if no items in cache",
			cache:   Cache{&cache{}},
			expected: expected{
				items:     []Item{},
				itemCount: 0,
			},
		},
	}

	for _, td := range testData {
		t.Run(td.message, func(t *testing.T) {
			items := td.cache.GetAll()
			assert.Equal(t, td.expected.itemCount, len(items))
			for i, item := range items {
				assert.Equal(t, td.expected.items[i].Value, item.Value)
				assert.Equal(t, td.expected.items[i].Key, item.Key)
			}
		})
	}
}

func TestGetAllItemsWIthExpiration(t *testing.T) {
	testData := []testDataWithExistingCache{
		{
			message: "Should get empty list if all items have expired",
			cache: Cache{
				&cache{
					items: []*Item{
						{Key: key{id: "first id", value: 1}, Value: "val 1", expireAt: time.Now().UnixNano() - 10},
						{Key: key{id: "second id", value: 2}, Value: "val 2", expireAt: time.Now().UnixNano() - 10},
					},
					expiration: 1 * time.Nanosecond,
				},
			},
			expected: expected{
				items:     []Item{},
				itemCount: 0,
			},
		},
		{
			message: "Should get only not expired items and remove the rest",
			cache: Cache{
				&cache{
					items: []*Item{
						{Key: key{id: "first id", value: 1}, Value: "val 1", expireAt: time.Now().UnixNano() + 10*time.Minute.Nanoseconds()},
						{Key: key{id: "second id", value: 2}, Value: "val 2", expireAt: time.Now().UnixNano() - 10},
						{Key: key{id: "third id", value: 3}, Value: "val 3", expireAt: time.Now().UnixNano() + 10*time.Minute.Nanoseconds()},
					},
					expiration: 1 * time.Nanosecond,
				},
			},
			expected: expected{
				items: []Item{
					{Key: key{id: "first id", value: 1}, Value: "val 1"},
					{Key: key{id: "third id", value: 3}, Value: "val 3"},
				},
				itemCount: 2,
			},
		},
	}

	for _, td := range testData {
		t.Run(td.message, func(t *testing.T) {
			items := td.cache.GetAll()
			assert.Equal(t, td.expected.itemCount, len(items))
			for i, item := range items {
				assert.Equal(t, td.expected.items[i].Value, item.Value)
				assert.Equal(t, td.expected.items[i].Key, item.Key)
			}
		})
	}
}

func TestDeleteItem(t *testing.T) {
	testData := []testDataWithExistingCache{
		{
			message: "Should have empty cache after deleting single item",
			cache: Cache{
				&cache{
					items: []*Item{
						{Key: key{id: "item to delete", value: 1}, Value: "val 1", expireAt: time.Now().UnixNano() - 10},
					},
					expiration: 1 * time.Nanosecond,
				},
			},
			expected: expected{
				items:     []Item{},
				itemCount: 0,
			},
		},
		{
			message: "Should have single item in cache after deleting single item",
			cache: Cache{
				&cache{
					items: []*Item{
						{Key: key{id: "first id", value: 1}, Value: "val 1", expireAt: time.Now().UnixNano() - 10},
						{Key: key{id: "item to delete", value: 2}, Value: "val 2", expireAt: time.Now().UnixNano() - 10},
						{Key: key{id: "third id", value: 3}, Value: "val 3", expireAt: time.Now().UnixNano() - 10},
					},
					expiration: 1 * time.Nanosecond,
				},
			},
			expected: expected{
				items: []Item{
					{Key: key{id: "first id", value: 1}, Value: "val 1"},
					{Key: key{id: "third id", value: 3}, Value: "val 3"},
				},
				itemCount: 2,
			},
		},
		{
			message: "Should not change cache if item does not exist",
			cache: Cache{
				&cache{
					items: []*Item{
						{Key: key{id: "first id", value: 1}, Value: "val 1", expireAt: time.Now().UnixNano() + 10*time.Minute.Nanoseconds()},
					},
					expiration: 1 * time.Nanosecond,
				},
			},
			expected: expected{
				items: []Item{
					{Key: key{id: "first id", value: 1}, Value: "val 1"},
				},
				itemCount: 1,
			},
		},
	}

	for _, td := range testData {
		t.Run(td.message, func(t *testing.T) {
			key := key{
				id:    "item to delete",
				value: 2,
			}

			td.cache.Delete(key)

			assert.Equal(t, td.expected.itemCount, td.cache.ItemCount())
			for i, item := range td.cache.GetAll() {
				assert.Equal(t, td.expected.items[i].Value, item.Value)
				assert.Equal(t, td.expected.items[i].Key, item.Key)
			}
		})
	}
}

func TestDeleteAllItems(t *testing.T) {
	testData := []testDataWithExistingCache{
		{
			message: "Should not do anything if cache already empty",
			cache: Cache{
				&cache{
					items:      []*Item{},
					expiration: 1 * time.Nanosecond,
				},
			},
			expected: expected{
				itemCount: 0,
			},
		},
		{
			message: "Should remove all items from non empty cache",
			cache: Cache{
				&cache{
					items: []*Item{
						{Key: key{id: "first id", value: 1}, Value: "val 1", expireAt: time.Now().UnixNano() + 10*time.Minute.Nanoseconds()},
						{Key: key{id: "second id", value: 2}, Value: "val 2", expireAt: time.Now().UnixNano() - 10},
						{Key: key{id: "third id", value: 3}, Value: "val 3", expireAt: time.Now().UnixNano() - 10},
					},
					expiration: 1 * time.Nanosecond,
				},
			},
			expected: expected{
				itemCount: 0,
			},
		},
	}

	for _, td := range testData {
		t.Run(td.message, func(t *testing.T) {
			td.cache.DeleteAll()

			assert.Equal(t, td.expected.itemCount, td.cache.ItemCount())
		})
	}
}

func TestEvictItems(t *testing.T) {
	testData := []testDataWithExistingCache{
		{
			message: "Should have empty cache after evicting all expired items",
			cache: Cache{
				&cache{
					items: []*Item{
						{Key: key{id: "first id", value: 1}, Value: "val 1", expireAt: time.Now().UnixNano() - 10},
						{Key: key{id: "second id", value: 2}, Value: "val 2", expireAt: time.Now().UnixNano() - 10},
					},
					expiration: 1 * time.Nanosecond,
				},
			},
			expected: expected{
				itemCount: 0,
			},
		},
		{
			message: "Should only have two unexpired items in cache",
			cache: Cache{
				&cache{
					items: []*Item{
						{Key: key{id: "first id", value: 1}, Value: "val 1", expireAt: time.Now().UnixNano() + 10*time.Minute.Nanoseconds()},
						{Key: key{id: "second id", value: 2}, Value: "val 2", expireAt: time.Now().UnixNano() - 10},
						{Key: key{id: "third id", value: 3}, Value: "val 3", expireAt: time.Now().UnixNano() + 10*time.Minute.Nanoseconds()},
					},
					expiration: 1 * time.Nanosecond,
				},
			},
			expected: expected{
				itemCount: 2,
			},
		},
	}

	for _, td := range testData {
		t.Run(td.message, func(t *testing.T) {
			td.cache.Evict()

			assert.Equal(t, td.expected.itemCount, td.cache.ItemCount())
		})
	}
}
