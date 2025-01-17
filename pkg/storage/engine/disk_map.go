// Copyright 2017 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package engine

import (
	"bytes"
	"context"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/diskmap"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/pkg/errors"
)

// defaultBatchCapacityBytes is the default capacity for a
// SortedDiskMapBatchWriter.
const defaultBatchCapacityBytes = 4096

// RocksDBMapBatchWriter batches writes to a RocksDBMap.
type RocksDBMapBatchWriter struct {
	// capacity is the number of bytes to write before a Flush() is triggered.
	capacity int

	// makeKey is a function that transforms a key into an MVCCKey with a prefix
	// to be written to the underlying store.
	makeKey func(k []byte) MVCCKey
	batch   Batch
	store   Engine
}

// RocksDBMapIterator iterates over the keys of a RocksDBMap in sorted order.
type RocksDBMapIterator struct {
	iter Iterator
	// makeKey is a function that transforms a key into an MVCCKey with a prefix
	// used to Seek() the underlying iterator.
	makeKey func(k []byte) MVCCKey
	// prefix is the prefix of keys that this iterator iterates over.
	prefix []byte
}

// RocksDBMap is a SortedDiskMap that uses RocksDB as its underlying storage
// engine.
type RocksDBMap struct {
	// TODO(asubiotto): Add memory accounting.
	prefix          []byte
	store           Engine
	allowDuplicates bool
	keyID           int64
}

var _ diskmap.SortedDiskMapBatchWriter = &RocksDBMapBatchWriter{}
var _ diskmap.SortedDiskMapIterator = &RocksDBMapIterator{}
var _ diskmap.SortedDiskMap = &RocksDBMap{}

// tempStorageID is the temp ID generator for a node. It generates unique
// prefixes for NewRocksDBMap. It is a global because NewRocksDBMap needs to
// prefix its writes uniquely, and using a global prevents users from having to
// specify the prefix themselves and correctly guarantee that it is unique.
var tempStorageID uint64

func generateTempStorageID() uint64 {
	return atomic.AddUint64(&tempStorageID, 1)
}

// NewRocksDBMap creates a new RocksDBMap with the passed in Engine as
// the underlying store. The RocksDBMap instance will have a keyspace prefixed
// by a unique prefix.
func NewRocksDBMap(e Engine) *RocksDBMap {
	return newRocksDBMap(e, false)
}

// NewRocksDBMultiMap creates a new RocksDBMap with the passed in Engine
// as the underlying store. The RocksDBMap instance will have a keyspace
// prefixed by a unique prefix. Unlike NewRocksDBMap, Puts with identical
// keys will write multiple entries (instead of overwriting previous entries)
// that will be returned during iteration.
func NewRocksDBMultiMap(e Engine) *RocksDBMap {
	return newRocksDBMap(e, true)
}

func newRocksDBMap(e Engine, allowDuplicates bool) *RocksDBMap {
	prefix := generateTempStorageID()
	return &RocksDBMap{
		prefix:          encoding.EncodeUvarintAscending([]byte(nil), prefix),
		store:           e,
		allowDuplicates: allowDuplicates,
	}
}

// makeKey appends k to the RocksDBMap's prefix to keep the key local to this
// instance and creates an MVCCKey, which is what the underlying storage engine
// expects. The returned key is only valid until the next call to makeKey().
func (r *RocksDBMap) makeKey(k []byte) MVCCKey {
	// TODO(asubiotto): We can make this more performant by bypassing MVCCKey
	// creation (have to generalize storage API). See
	// https://github.com/cockroachdb/cockroach/issues/16718#issuecomment-311493414
	prefixLen := len(r.prefix)
	r.prefix = append(r.prefix, k...)
	mvccKey := MVCCKey{Key: r.prefix}
	r.prefix = r.prefix[:prefixLen]
	return mvccKey
}

// makeKeyWithTimestamp makes a key appropriate for a Put operation. It is like
// makeKey except it respects allowDuplicates, which uses the MVCC timestamp
// field to assign a unique keyID so duplicate keys don't overwrite each other.
func (r *RocksDBMap) makeKeyWithTimestamp(k []byte) MVCCKey {
	mvccKey := r.makeKey(k)
	if r.allowDuplicates {
		r.keyID++
		mvccKey.Timestamp.WallTime = r.keyID
	}
	return mvccKey
}

// Put implements the SortedDiskMap interface.
func (r *RocksDBMap) Put(k []byte, v []byte) error {
	return r.store.Put(r.makeKeyWithTimestamp(k), v)
}

// Get implements the SortedDiskMap interface.
func (r *RocksDBMap) Get(k []byte) ([]byte, error) {
	if r.allowDuplicates {
		return nil, errors.New("Get not supported if allowDuplicates is true")
	}
	return r.store.Get(r.makeKey(k))
}

// NewIterator implements the SortedDiskMap interface.
func (r *RocksDBMap) NewIterator() diskmap.SortedDiskMapIterator {
	// NOTE: prefix is only false because we can't use the normal prefix
	// extractor. This iterator still only does prefix iteration. See
	// RocksDBMapIterator.Valid().
	return &RocksDBMapIterator{
		iter: r.store.NewIterator(IterOptions{
			UpperBound: roachpb.Key(r.prefix).PrefixEnd(),
		}),
		makeKey: r.makeKey,
		prefix:  r.prefix,
	}
}

// NewBatchWriter implements the SortedDiskMap interface.
func (r *RocksDBMap) NewBatchWriter() diskmap.SortedDiskMapBatchWriter {
	return r.NewBatchWriterCapacity(defaultBatchCapacityBytes)
}

// NewBatchWriterCapacity implements the SortedDiskMap interface.
func (r *RocksDBMap) NewBatchWriterCapacity(capacityBytes int) diskmap.SortedDiskMapBatchWriter {
	makeKey := r.makeKey
	if r.allowDuplicates {
		makeKey = r.makeKeyWithTimestamp
	}
	return &RocksDBMapBatchWriter{
		capacity: capacityBytes,
		makeKey:  makeKey,
		batch:    r.store.NewWriteOnlyBatch(),
		store:    r.store,
	}
}

// Clear implements the SortedDiskMap interface.
func (r *RocksDBMap) Clear() error {
	if err := r.store.ClearRange(
		MVCCKey{Key: r.prefix},
		MVCCKey{Key: roachpb.Key(r.prefix).PrefixEnd()},
	); err != nil {
		return errors.Wrapf(err, "unable to clear range with prefix %v", r.prefix)
	}
	// NB: we manually flush after performing the clear range to ensure that the
	// range tombstone is pushed to disk which will kick off compactions that
	// will eventually free up the deleted space.
	return r.store.Flush()
}

// Close implements the SortedDiskMap interface.
func (r *RocksDBMap) Close(ctx context.Context) {
	if err := r.Clear(); err != nil {
		log.Error(ctx, err)
	}
}

// Seek implements the SortedDiskMapIterator interface.
func (i *RocksDBMapIterator) Seek(k []byte) {
	i.iter.Seek(i.makeKey(k))
}

// Rewind implements the SortedDiskMapIterator interface.
func (i *RocksDBMapIterator) Rewind() {
	i.iter.Seek(i.makeKey(nil))
}

// Valid implements the SortedDiskMapIterator interface.
func (i *RocksDBMapIterator) Valid() (bool, error) {
	ok, err := i.iter.Valid()
	if err != nil {
		return false, err
	}
	if ok && !bytes.HasPrefix(i.iter.UnsafeKey().Key, i.prefix) {
		return false, nil
	}

	return ok, nil
}

// Next implements the SortedDiskMapIterator interface.
func (i *RocksDBMapIterator) Next() {
	i.iter.Next()
}

// Key implements the SortedDiskMapIterator interface.
func (i *RocksDBMapIterator) Key() []byte {
	return i.iter.Key().Key[len(i.prefix):]
}

// Value implements the SortedDiskMapIterator interface.
func (i *RocksDBMapIterator) Value() []byte {
	return i.iter.Value()
}

// UnsafeKey implements the SortedDiskMapIterator interface.
func (i *RocksDBMapIterator) UnsafeKey() []byte {
	return i.iter.UnsafeKey().Key[len(i.prefix):]
}

// UnsafeValue implements the SortedDiskMapIterator interface.
func (i *RocksDBMapIterator) UnsafeValue() []byte {
	return i.iter.UnsafeValue()
}

// Close implements the SortedDiskMapIterator interface.
func (i *RocksDBMapIterator) Close() {
	i.iter.Close()
}

// Put implements the SortedDiskMapBatchWriter interface.
func (b *RocksDBMapBatchWriter) Put(k []byte, v []byte) error {
	if err := b.batch.Put(b.makeKey(k), v); err != nil {
		return err
	}
	if b.batch.Len() >= b.capacity {
		return b.Flush()
	}
	return nil
}

// Flush implements the SortedDiskMapBatchWriter interface.
func (b *RocksDBMapBatchWriter) Flush() error {
	if b.batch.Empty() {
		return nil
	}
	if err := b.batch.Commit(false /* syncCommit */); err != nil {
		return err
	}
	b.batch = b.store.NewWriteOnlyBatch()
	return nil
}

// Close implements the SortedDiskMapBatchWriter interface.
func (b *RocksDBMapBatchWriter) Close(ctx context.Context) error {
	err := b.Flush()
	b.batch.Close()
	return err
}
