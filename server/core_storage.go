// Copyright 2018 The Nakama Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"sort"

	"context"

	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/heroiclabs/nakama/api"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

var (
	ErrStorageRejectedVersion    = errors.New("Storage write rejected - version check failed.")
	ErrStorageRejectedPermission = errors.New("Storage write rejected - permission denied.")
	ErrStorageWriteFailed        = errors.New("Storage write failed.")
)

type storageCursor struct {
	Key    string
	UserID uuid.UUID
	Read   int32
}

// Internal representation for a batch of storage write operations.
type StorageOpWrites []*StorageOpWrite

type StorageOpWrite struct {
	OwnerID string
	Object  *api.WriteStorageObject
}

func (s StorageOpWrites) Len() int {
	return len(s)
}
func (s StorageOpWrites) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s StorageOpWrites) Less(i, j int) bool {
	s1, s2 := s[i], s[j]
	if s1.Object.Collection < s2.Object.Collection {
		return true
	}
	if s1.Object.Key < s2.Object.Key {
		return true
	}
	return s1.OwnerID < s2.OwnerID
}

// Internal representation for a batch of storage delete operations.
type StorageOpDeletes []*StorageOpDelete

type StorageOpDelete struct {
	OwnerID  string
	ObjectID *api.DeleteStorageObjectId
}

func (s StorageOpDeletes) Len() int {
	return len(s)
}
func (s StorageOpDeletes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s StorageOpDeletes) Less(i, j int) bool {
	s1, s2 := s[i], s[j]
	if s1.ObjectID.Collection < s2.ObjectID.Collection {
		return true
	}
	if s1.ObjectID.Key < s2.ObjectID.Key {
		return true
	}
	return s1.OwnerID < s2.OwnerID
}

func StorageListObjects(ctx context.Context, logger *zap.Logger, db *sql.DB, caller uuid.UUID, ownerID *uuid.UUID, collection string, limit int, cursor string) (*api.StorageObjectList, codes.Code, error) {
	var sc *storageCursor = nil
	if cursor != "" {
		sc = &storageCursor{}
		if cb, err := base64.RawURLEncoding.DecodeString(cursor); err != nil {
			logger.Warn("Could not base64 decode storage cursor.", zap.String("cursor", cursor))
			return nil, codes.InvalidArgument, errors.New("Malformed cursor was used.")
		} else {
			if err := gob.NewDecoder(bytes.NewReader(cb)).Decode(sc); err != nil {
				logger.Warn("Could not decode storage cursor.", zap.String("cursor", cursor))
				return nil, codes.InvalidArgument, errors.New("Malformed cursor was used.")
			}
		}
	}

	var result *api.StorageObjectList
	var resultErr error

	if caller == uuid.Nil {
		// Call from the runtime.
		if ownerID == nil {
			// List storage regardless of user.
			// TODO
			result, resultErr = StorageListObjectsAll(ctx, logger, db, true, collection, limit, cursor, sc)
		} else {
			// List for a particular user ID.
			result, resultErr = StorageListObjectsUser(ctx, logger, db, true, *ownerID, collection, limit, cursor, sc)
		}
	} else {
		// Call from a client.
		if ownerID == nil {
			// List publicly readable storage regardless of owner.
			result, resultErr = StorageListObjectsAll(ctx, logger, db, false, collection, limit, cursor, sc)
		} else if o := *ownerID; caller == o {
			// User listing their own data.
			result, resultErr = StorageListObjectsUser(ctx, logger, db, false, o, collection, limit, cursor, sc)
		} else {
			// User listing someone else's data.
			result, resultErr = StorageListObjectsPublicReadUser(ctx, logger, db, o, collection, limit, cursor, sc)
		}
	}

	if resultErr != nil {
		return nil, codes.Internal, resultErr
	}

	return result, codes.OK, nil
}

func StorageListObjectsAll(ctx context.Context, logger *zap.Logger, db *sql.DB, authoritative bool, collection string, limit int, cursor string, storageCursor *storageCursor) (*api.StorageObjectList, error) {
	cursorQuery := ""
	params := []interface{}{collection, limit}
	if storageCursor != nil {
		cursorQuery = ` AND (collection, read, key, user_id) > ($1, 2, $3, $4) `
		params = append(params, storageCursor.Key, storageCursor.UserID)
	}

	var query string
	if authoritative {
		query = `
SELECT collection, key, user_id, value, version, read, write, create_time, update_time
FROM storage
WHERE collection = $1` + cursorQuery + `
LIMIT $2`
	} else {
		query = `
SELECT collection, key, user_id, value, version, read, write, create_time, update_time
FROM storage
WHERE collection = $1 AND read = 2` + cursorQuery + `
LIMIT $2`
	}

	var objects *api.StorageObjectList
	err := ExecuteRetryable(func() error {
		rows, err := db.QueryContext(ctx, query, params...)
		if err != nil {
			if err == sql.ErrNoRows {
				objects = &api.StorageObjectList{Objects: make([]*api.StorageObject, 0), Cursor: cursor}
				return nil
			} else {
				logger.Error("Could not list storage.", zap.Error(err), zap.String("collection", collection), zap.Int("limit", limit), zap.String("cursor", cursor))
				return err
			}
		}
		// rows.Close() called in storageListObjects

		objects, err = storageListObjects(rows, cursor)
		if err != nil {
			logger.Error("Could not list storage.", zap.Error(err), zap.String("collection", collection), zap.Int("limit", limit), zap.String("cursor", cursor))
			return err
		}
		return nil
	})

	return objects, err
}

func StorageListObjectsPublicReadUser(ctx context.Context, logger *zap.Logger, db *sql.DB, userID uuid.UUID, collection string, limit int, cursor string, storageCursor *storageCursor) (*api.StorageObjectList, error) {
	cursorQuery := ""
	params := []interface{}{collection, userID, limit}
	if storageCursor != nil {
		cursorQuery = ` AND (collection, read, key, user_id) > ($1, 2, $4, $5) `
		params = append(params, storageCursor.Key, storageCursor.UserID)
	}

	query := `
SELECT collection, key, user_id, value, version, read, write, create_time, update_time
FROM storage
WHERE collection = $1 AND read = 2 AND user_id = $2 ` + cursorQuery + `
LIMIT $3`

	var objects *api.StorageObjectList
	err := ExecuteRetryable(func() error {
		rows, err := db.QueryContext(ctx, query, params...)
		if err != nil {
			if err == sql.ErrNoRows {
				objects = &api.StorageObjectList{Objects: make([]*api.StorageObject, 0), Cursor: cursor}
				return nil
			} else {
				logger.Error("Could not list storage.", zap.Error(err), zap.String("collection", collection), zap.Int("limit", limit), zap.String("cursor", cursor))
				return err
			}
		}
		// rows.Close() called in storageListObjects

		objects, err = storageListObjects(rows, cursor)
		if err != nil {
			logger.Error("Could not list storage.", zap.Error(err), zap.String("collection", collection), zap.Int("limit", limit), zap.String("cursor", cursor))
			return err
		}
		return nil
	})

	return objects, err
}

func StorageListObjectsUser(ctx context.Context, logger *zap.Logger, db *sql.DB, authoritative bool, userID uuid.UUID, collection string, limit int, cursor string, storageCursor *storageCursor) (*api.StorageObjectList, error) {
	cursorQuery := ""
	params := []interface{}{collection, userID, limit}
	if storageCursor != nil {
		cursorQuery = ` AND (collection, read, key, user_id) > ($1, $4, $5, $6) `
		params = append(params, storageCursor.Read, storageCursor.Key, storageCursor.UserID)
	}

	query := `
SELECT collection, key, user_id, value, version, read, write, create_time, update_time
FROM storage
WHERE collection = $1 AND read > 0 AND user_id = $2 ` + cursorQuery + `
LIMIT $3`
	if authoritative {
		// disregard permissions
		query = `
SELECT collection, key, user_id, value, version, read, write, create_time, update_time
FROM storage
WHERE collection = $1 AND user_id = $2 ` + cursorQuery + `
LIMIT $3`
	}

	var objects *api.StorageObjectList
	err := ExecuteRetryable(func() error {
		rows, err := db.QueryContext(ctx, query, params...)
		if err != nil {
			if err == sql.ErrNoRows {
				objects = &api.StorageObjectList{Objects: make([]*api.StorageObject, 0), Cursor: cursor}
				return nil
			} else {
				logger.Error("Could not list storage.", zap.Error(err), zap.String("collection", collection), zap.Int("limit", limit), zap.String("cursor", cursor))
				return err
			}
		}
		// rows.Close() called in storageListObjects

		objects, err = storageListObjects(rows, cursor)
		if err != nil {
			logger.Error("Could not list storage.", zap.Error(err), zap.String("collection", collection), zap.Int("limit", limit), zap.String("cursor", cursor))
			return err
		}
		return nil
	})

	return objects, err
}

func StorageReadAllUserObjects(ctx context.Context, logger *zap.Logger, db *sql.DB, userID uuid.UUID) ([]*api.StorageObject, error) {
	query := `
SELECT collection, key, user_id, value, version, read, write, create_time, update_time
FROM storage
WHERE user_id = $1`

	var objects []*api.StorageObject
	err := ExecuteRetryable(func() error {
		rows, err := db.QueryContext(ctx, query, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				objects = make([]*api.StorageObject, 0)
				return nil
			} else {
				logger.Error("Could not read storage objects.", zap.Error(err), zap.String("user_id", userID.String()))
				return err
			}
		}
		defer rows.Close()

		funcObjects := make([]*api.StorageObject, 0)
		for rows.Next() {
			o := &api.StorageObject{CreateTime: &timestamp.Timestamp{}, UpdateTime: &timestamp.Timestamp{}}
			var createTime pq.NullTime
			var updateTime pq.NullTime
			var userID sql.NullString
			if err := rows.Scan(&o.Collection, &o.Key, &userID, &o.Value, &o.Version, &o.PermissionRead, &o.PermissionWrite, &createTime, &updateTime); err != nil {
				return err
			}

			o.CreateTime.Seconds = createTime.Time.Unix()
			o.UpdateTime.Seconds = updateTime.Time.Unix()

			o.UserId = userID.String
			funcObjects = append(funcObjects, o)
		}

		if rows.Err() != nil {
			logger.Error("Could not read storage objects.", zap.Error(err), zap.String("user_id", userID.String()))
			return rows.Err()
		}
		objects = funcObjects
		return nil
	})

	return objects, err
}

func storageListObjects(rows *sql.Rows, cursor string) (*api.StorageObjectList, error) {
	objects := make([]*api.StorageObject, 0)
	for rows.Next() {
		o := &api.StorageObject{CreateTime: &timestamp.Timestamp{}, UpdateTime: &timestamp.Timestamp{}}
		var createTime pq.NullTime
		var updateTime pq.NullTime
		var userID sql.NullString
		if err := rows.Scan(&o.Collection, &o.Key, &userID, &o.Value, &o.Version, &o.PermissionRead, &o.PermissionWrite, &createTime, &updateTime); err != nil {
			rows.Close()
			return nil, err
		}

		o.CreateTime.Seconds = createTime.Time.Unix()
		o.UpdateTime.Seconds = updateTime.Time.Unix()

		o.UserId = userID.String
		objects = append(objects, o)
	}
	rows.Close()

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	objectList := &api.StorageObjectList{
		Objects: objects,
		Cursor:  cursor,
	}

	if len(objects) > 0 {
		lastObject := objects[len(objects)-1]
		newCursor := &storageCursor{
			Key:  lastObject.Key,
			Read: lastObject.PermissionRead,
		}

		if lastObject.UserId != "" {
			newCursor.UserID = uuid.FromStringOrNil(lastObject.UserId)
		}

		cursorBuf := new(bytes.Buffer)
		if err := gob.NewEncoder(cursorBuf).Encode(newCursor); err != nil {
			return nil, err
		}
		objectList.Cursor = base64.RawURLEncoding.EncodeToString(cursorBuf.Bytes())
	}

	return objectList, nil
}

func StorageReadObjects(ctx context.Context, logger *zap.Logger, db *sql.DB, caller uuid.UUID, objectIDs []*api.ReadStorageObjectId) (*api.StorageObjects, error) {
	params := make([]interface{}, 0)

	whereClause := ""
	for _, id := range objectIDs {
		l := len(params)
		if whereClause != "" {
			whereClause += " OR "
		}

		if caller == uuid.Nil {
			// Disregard permissions if called authoritatively.
			whereClause += fmt.Sprintf(" (collection = $%v AND key = $%v AND user_id = $%v) ", l+1, l+2, l+3)
			if id.UserId == "" {
				params = append(params, id.Collection, id.Key, uuid.Nil)
			} else {
				params = append(params, id.Collection, id.Key, id.UserId)
			}
		} else if id.GetUserId() == "" {
			whereClause += fmt.Sprintf(" (collection = $%v AND key = $%v AND user_id = $%v AND read = 2) ", l+1, l+2, l+3)
			params = append(params, id.Collection, id.Key, uuid.Nil)
		} else {
			whereClause += fmt.Sprintf(" (collection = $%v AND key = $%v AND user_id = $%v AND (read = 2 OR (read = 1 AND user_id = $%v))) ", l+1, l+2, l+3, l+4)
			params = append(params, id.Collection, id.Key, id.UserId, caller)
		}
	}

	query := `
SELECT collection, key, user_id, value, version, read, write, create_time, update_time
FROM storage
WHERE
` + whereClause

	var objects *api.StorageObjects
	err := ExecuteRetryable(func() error {
		rows, err := db.QueryContext(ctx, query, params...)
		if err != nil {
			if err == sql.ErrNoRows {
				objects = &api.StorageObjects{Objects: make([]*api.StorageObject, 0)}
				return nil
			} else {
				logger.Error("Could not read storage objects.", zap.Error(err))
				return err
			}
		}
		defer rows.Close()

		funcObjects := &api.StorageObjects{Objects: make([]*api.StorageObject, 0)}
		for rows.Next() {
			o := &api.StorageObject{CreateTime: &timestamp.Timestamp{}, UpdateTime: &timestamp.Timestamp{}}
			var createTime pq.NullTime
			var updateTime pq.NullTime

			var userID sql.NullString
			if err := rows.Scan(&o.Collection, &o.Key, &userID, &o.Value, &o.Version, &o.PermissionRead, &o.PermissionWrite, &createTime, &updateTime); err != nil {
				return err
			}

			o.CreateTime.Seconds = createTime.Time.Unix()
			o.UpdateTime.Seconds = updateTime.Time.Unix()

			if uuid.FromStringOrNil(userID.String) != uuid.Nil {
				o.UserId = userID.String
			}
			funcObjects.Objects = append(funcObjects.Objects, o)
		}
		if err = rows.Err(); err != nil {
			logger.Error("Could not read storage objects.", zap.Error(err))
			return err
		}
		objects = funcObjects
		return nil
	})

	return objects, err
}

func StorageWriteObjects(ctx context.Context, logger *zap.Logger, db *sql.DB, authoritativeWrite bool, ops StorageOpWrites) (*api.StorageObjectAcks, codes.Code, error) {
	// Ensure writes are processed in a consistent order.
	sort.Sort(ops)

	var acks []*api.StorageObjectAck

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Could not begin database transaction.", zap.Error(err))
		return nil, codes.Internal, err
	}

	if err = crdb.ExecuteInTx(ctx, tx, func() error {
		acks = make([]*api.StorageObjectAck, 0, ops.Len())

		for _, op := range ops {
			ack, writeErr := storageWriteObject(ctx, logger, tx, authoritativeWrite, op.OwnerID, op.Object)
			if writeErr != nil {
				if writeErr == ErrStorageRejectedVersion || writeErr == ErrStorageRejectedPermission {
					return StatusError(codes.InvalidArgument, "Storage write rejected.", writeErr)
				}

				logger.Debug("Error writing storage objects.", zap.Error(err))
				return writeErr
			}

			acks = append(acks, ack)
		}
		return nil
	}); err != nil {
		if e, ok := err.(*statusError); ok {
			return nil, e.Code(), e.Cause()
		}
		logger.Error("Error writing storage objects.", zap.Error(err))
		return nil, codes.Internal, err
	}

	return &api.StorageObjectAcks{Acks: acks}, codes.OK, nil
}

func storageWriteObject(ctx context.Context, logger *zap.Logger, tx *sql.Tx, authoritativeWrite bool, ownerID string, object *api.WriteStorageObject) (*api.StorageObjectAck, error) {
	var dbVersion sql.NullString
	var dbPermissionWrite sql.NullInt64
	var dbPermissionRead sql.NullInt64
	err := tx.QueryRowContext(ctx, "SELECT version, read, write FROM storage WHERE collection = $1 AND key = $2 AND user_id = $3", object.Collection, object.Key, ownerID).Scan(&dbVersion, &dbPermissionRead, &dbPermissionWrite)
	if err != nil {
		if err == sql.ErrNoRows {
			if object.Version != "" && object.Version != "*" {
				// Conditional write with a specific version but the object did not exist at all.
				return nil, ErrStorageRejectedVersion
			}
		} else {
			logger.Debug("Error in write storage object pre-flight.", zap.Any("object", object), zap.Error(err))
			return nil, err
		}
	}

	if dbVersion.Valid && (object.Version == "*" || (object.Version != "" && object.Version != dbVersion.String)) {
		// An object existed and it's a conditional write that either:
		// - Expects no object.
		// - Or expects a given version bit it does not match.
		return nil, ErrStorageRejectedVersion
	}

	if dbPermissionWrite.Valid && dbPermissionWrite.Int64 == 0 && !authoritativeWrite {
		// Non-authoritative write to an existing storage object with permission 0.
		return nil, ErrStorageRejectedPermission
	}

	newVersion := fmt.Sprintf("%x", md5.Sum([]byte(object.Value)))
	newPermissionRead := int32(1)
	if object.PermissionRead != nil {
		newPermissionRead = object.PermissionRead.Value
	}
	newPermissionWrite := int32(1)
	if object.PermissionWrite != nil {
		newPermissionWrite = object.PermissionWrite.Value
	}

	if dbVersion.Valid && dbVersion.String == newVersion && dbPermissionRead.Int64 == int64(newPermissionRead) && dbPermissionWrite.Int64 == int64(newPermissionWrite) {
		// Stored object existed, and exactly matches the new object's version and read/write permissions.
		ack := &api.StorageObjectAck{
			Collection: object.Collection,
			Key:        object.Key,
			Version:    newVersion,
		}
		if ownerID != uuid.Nil.String() {
			ack.UserId = ownerID
		}
		return ack, nil
	}

	var query string
	if dbVersion.Valid {
		// Updating an existing storage object.
		query = "UPDATE storage SET value = $4, version = $5, read = $6, write = $7, update_time = now() WHERE collection = $1 AND key = $2 AND user_id = $3::UUID"
	} else {
		// Inserting a new storage object.
		query = "INSERT INTO storage (collection, key, user_id, value, version, read, write, create_time, update_time) VALUES ($1, $2, $3::UUID, $4, $5, $6, $7, now(), now())"
	}

	res, err := tx.ExecContext(ctx, query, object.Collection, object.Key, ownerID, object.Value, newVersion, newPermissionRead, newPermissionWrite)
	if err != nil {
		logger.Debug("Could not write storage object, exec error.", zap.Any("object", object), zap.String("query", query), zap.Error(err))
		return nil, err
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		logger.Debug("Could not write storage object, rowsAffected error.", zap.Any("object", object), zap.String("query", query), zap.Error(err))
		return nil, ErrStorageWriteFailed
	}

	ack := &api.StorageObjectAck{
		Collection: object.Collection,
		Key:        object.Key,
		Version:    newVersion,
	}
	if ownerID != uuid.Nil.String() {
		ack.UserId = ownerID
	}

	return ack, nil
}

func StorageDeleteObjects(ctx context.Context, logger *zap.Logger, db *sql.DB, authoritativeDelete bool, ops StorageOpDeletes) (codes.Code, error) {
	// Ensure deletes are processed in a consistent order.
	sort.Sort(ops)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Could not begin database transaction.", zap.Error(err))
		return codes.Internal, err
	}

	if err = crdb.ExecuteInTx(ctx, tx, func() error {
		for _, op := range ops {
			params := []interface{}{op.ObjectID.Collection, op.ObjectID.Key, op.OwnerID}
			var query string
			if authoritativeDelete {
				// Deleting from the runtime.
				query = "DELETE FROM storage WHERE collection = $1 AND key = $2 AND user_id = $3"
			} else {
				// Direct client request to delete.
				query = "DELETE FROM storage WHERE collection = $1 AND key = $2 AND user_id = $3 AND write > 0"
			}
			if op.ObjectID.GetVersion() != "" {
				// Conditional delete.
				params = append(params, op.ObjectID.Version)
				query += fmt.Sprintf(" AND version = $4")
			}

			result, err := tx.ExecContext(ctx, query, params...)
			if err != nil {
				logger.Debug("Could not delete storage object.", zap.Error(err), zap.String("query", query), zap.Any("object_id", op.ObjectID))
				return err
			}

			if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
				return StatusError(codes.InvalidArgument, "Storage delete rejected.", errors.New("Storage delete rejected - not found, version check failed, or permission denied."))
			}
		}
		return nil
	}); err != nil {
		if e, ok := err.(*statusError); ok {
			return e.Code(), e.Cause()
		}
		logger.Error("Error deleting storage objects.", zap.Error(err))
		return codes.Internal, err
	}

	return codes.OK, nil
}
