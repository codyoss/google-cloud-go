// Copyright 2014 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datastore

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	pb "cloud.google.com/go/datastore/apiv1/datastorepb"
	"cloud.google.com/go/internal/trace"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	"google.golang.org/api/transport"
	gtransport "google.golang.org/api/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	prodEndpointTemplate = "datastore.UNIVERSE_DOMAIN:443"
	prodUniverseDomain   = "googleapis.com"
	prodMtlsAddr         = "datastore.mtls.googleapis.com:443"
	userAgent            = "gcloud-golang-datastore/20160401"
)

// ScopeDatastore grants permissions to view and/or manage datastore entities
const ScopeDatastore = "https://www.googleapis.com/auth/datastore"

// DetectProjectID is a sentinel value that instructs NewClient to detect the
// project ID. It is given in place of the projectID argument. NewClient will
// use the project ID from the given credentials or the default credentials
// (https://developers.google.com/accounts/docs/application-default-credentials)
// if no credentials were provided. When providing credentials, not all
// options will allow NewClient to extract the project ID. Specifically a JWT
// does not have the project ID encoded.
const DetectProjectID = "*detect-project-id*"

// resourcePrefixHeader is the name of the metadata header used to indicate
// the resource being operated on.
const resourcePrefixHeader = "google-cloud-resource-prefix"

// reqParamsHeader is routing header required to access named databases
const reqParamsHeader = "x-goog-request-params"

// DefaultDatabaseID is ID of the default database denoted by an empty string
const DefaultDatabaseID = ""

var (
	gtransportDialPoolFn = gtransport.DialPool
	detectProjectIDFn    = detectProjectID
)

// Client is a client for reading and writing data in a datastore dataset.
type Client struct {
	connPool     gtransport.ConnPool
	client       pb.DatastoreClient
	dataset      string // Called dataset by the datastore API, synonym for project ID.
	databaseID   string // Default value is empty string
	readSettings *readSettings
	config       *datastoreConfig
}

// NewClient creates a new Client for a given dataset.  If the project ID is
// empty, it is derived from the DATASTORE_PROJECT_ID environment variable.
// If the DATASTORE_EMULATOR_HOST environment variable is set, client will use
// its value to connect to a locally-running datastore emulator.
// DetectProjectID can be passed as the projectID argument to instruct
// NewClient to detect the project ID from the credentials.
// Call (*Client).Close() when done with the client.
func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	client, err := NewClientWithDatabase(ctx, projectID, DefaultDatabaseID, opts...)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewClientWithDatabase creates a new Client for given dataset and database.
// If the project ID is empty, it is derived from the DATASTORE_PROJECT_ID environment variable.
// If the DATASTORE_EMULATOR_HOST environment variable is set, client will use
// its value to connect to a locally-running datastore emulator.
// DetectProjectID can be passed as the projectID argument to instruct
// NewClientWithDatabase to detect the project ID from the credentials.
// Call (*Client).Close() when done with the client.
func NewClientWithDatabase(ctx context.Context, projectID, databaseID string, opts ...option.ClientOption) (*Client, error) {
	var o []option.ClientOption
	// Environment variables for gcd emulator:
	// https://cloud.google.com/datastore/docs/tools/datastore-emulator
	// If the emulator is available, dial it without passing any credentials.
	if addr := os.Getenv("DATASTORE_EMULATOR_HOST"); addr != "" {
		o = []option.ClientOption{
			option.WithEndpoint(addr),
			option.WithoutAuthentication(),
			option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		}
		if projectID == DetectProjectID {
			projectID, _ = detectProjectIDFn(ctx, opts...)
			if projectID == "" {
				projectID = "dummy-emulator-datastore-project"
			}
		}
	} else {
		o = []option.ClientOption{
			internaloption.WithDefaultEndpointTemplate(prodEndpointTemplate),
			internaloption.WithDefaultUniverseDomain(prodUniverseDomain),
			internaloption.WithDefaultMTLSEndpoint(prodMtlsAddr),
			option.WithScopes(ScopeDatastore),
			option.WithUserAgent(userAgent),
		}
	}
	// Warn if we see the legacy emulator environment variables.
	if os.Getenv("DATASTORE_HOST") != "" && os.Getenv("DATASTORE_EMULATOR_HOST") == "" {
		log.Print("WARNING: legacy environment variable DATASTORE_HOST is ignored. Use DATASTORE_EMULATOR_HOST instead.")
	}
	if os.Getenv("DATASTORE_DATASET") != "" && os.Getenv("DATASTORE_PROJECT_ID") == "" {
		log.Print("WARNING: legacy environment variable DATASTORE_DATASET is ignored. Use DATASTORE_PROJECT_ID instead.")
	}
	if projectID == "" {
		projectID = os.Getenv("DATASTORE_PROJECT_ID")
	}

	o = append(o, opts...)

	if projectID == DetectProjectID {
		detected, err := detectProjectIDFn(ctx, opts...)
		if err != nil {
			return nil, err
		}
		projectID = detected
	}

	if projectID == "" {
		return nil, errors.New("datastore: missing project/dataset id")
	}
	connPool, err := gtransportDialPoolFn(ctx, o...)
	if err != nil {
		return nil, fmt.Errorf("dialing: %w", err)
	}

	config := newDatastoreConfig(o...)
	return &Client{
		connPool:     connPool,
		client:       newDatastoreClient(connPool, projectID, databaseID),
		dataset:      projectID,
		readSettings: &readSettings{},
		databaseID:   databaseID,
		config:       &config,
	}, nil
}

func detectProjectID(ctx context.Context, opts ...option.ClientOption) (string, error) {
	creds, err := transport.Creds(ctx, opts...)
	if err != nil {
		return "", fmt.Errorf("fetching creds: %w", err)
	}
	if creds.ProjectID == "" {
		return "", errors.New("datastore: see the docs on DetectProjectID")
	}
	return creds.ProjectID, nil
}

var (
	// ErrInvalidEntityType is returned when functions like Get or Next are
	// passed a dst or src argument of invalid type.
	ErrInvalidEntityType = errors.New("datastore: invalid entity type")
	// ErrInvalidKey is returned when an invalid key is presented.
	ErrInvalidKey = errors.New("datastore: invalid key")
	// ErrNoSuchEntity is returned when no entity was found for a given key.
	ErrNoSuchEntity = errors.New("datastore: no such entity")
	// ErrDifferentKeyAndDstLength is returned when the length of dst and key are different.
	ErrDifferentKeyAndDstLength = errors.New("datastore: keys and dst slices have different length")
)

type multiArgType int

const (
	multiArgTypeInvalid multiArgType = iota
	multiArgTypePropertyLoadSaver
	multiArgTypeStruct
	multiArgTypeStructPtr
	multiArgTypeInterface
)

// ErrFieldMismatch is returned when a field is to be loaded into a different
// type than the one it was stored from, or when a field is missing or
// unexported in the destination struct.
// StructType is the type of the struct pointed to by the destination argument
// passed to Get or to Iterator.Next.
type ErrFieldMismatch struct {
	StructType reflect.Type
	FieldName  string
	Reason     string
}

func (e *ErrFieldMismatch) Error() string {
	return fmt.Sprintf("datastore: cannot load field %q into a %q: %s",
		e.FieldName, e.StructType, e.Reason)
}

// GeoPoint represents a location as latitude/longitude in degrees.
type GeoPoint struct {
	Lat, Lng float64
}

// Valid returns whether a GeoPoint is within [-90, 90] latitude and [-180, 180] longitude.
func (g GeoPoint) Valid() bool {
	return -90 <= g.Lat && g.Lat <= 90 && -180 <= g.Lng && g.Lng <= 180
}

func keyToProto(k *Key) *pb.Key {
	if k == nil {
		return nil
	}

	var path []*pb.Key_PathElement
	for {
		el := &pb.Key_PathElement{Kind: k.Kind}
		if k.ID != 0 {
			el.IdType = &pb.Key_PathElement_Id{Id: k.ID}
		} else if k.Name != "" {
			el.IdType = &pb.Key_PathElement_Name{Name: k.Name}
		}
		path = append(path, el)
		if k.Parent == nil {
			break
		}
		k = k.Parent
	}

	// The path should be in order [grandparent, parent, child]
	// We did it backward above, so reverse back.
	for i := 0; i < len(path)/2; i++ {
		path[i], path[len(path)-i-1] = path[len(path)-i-1], path[i]
	}

	key := &pb.Key{Path: path}
	if k.Namespace != "" {
		key.PartitionId = &pb.PartitionId{
			NamespaceId: k.Namespace,
		}
	}
	return key
}

// protoToKey decodes a protocol buffer representation of a key into an
// equivalent *Key object. If the key is invalid, protoToKey will return the
// invalid key along with ErrInvalidKey.
func protoToKey(p *pb.Key) (*Key, error) {
	var key *Key
	var namespace string
	if partition := p.PartitionId; partition != nil {
		namespace = partition.NamespaceId
	}
	for _, el := range p.Path {
		key = &Key{
			Namespace: namespace,
			Kind:      el.Kind,
			ID:        el.GetId(),
			Name:      el.GetName(),
			Parent:    key,
		}
	}
	if !key.valid() { // Also detects key == nil.
		return key, ErrInvalidKey
	}
	return key, nil
}

// multiKeyToProto is a batch version of keyToProto.
func multiKeyToProto(keys []*Key) []*pb.Key {
	ret := make([]*pb.Key, len(keys))
	for i, k := range keys {
		ret[i] = keyToProto(k)
	}
	return ret
}

// multiKeyToProto is a batch version of keyToProto.
func multiProtoToKey(keys []*pb.Key) ([]*Key, error) {
	hasErr := false
	ret := make([]*Key, len(keys))
	err := make(MultiError, len(keys))
	for i, k := range keys {
		ret[i], err[i] = protoToKey(k)
		if err[i] != nil {
			hasErr = true
		}
	}
	if hasErr {
		return nil, err
	}
	return ret, nil
}

// multiValid is a batch version of Key.valid. It returns an error, not a
// []bool.
func multiValid(key []*Key) error {
	invalid := false
	for _, k := range key {
		if !k.valid() {
			invalid = true
			break
		}
	}
	if !invalid {
		return nil
	}
	err := make(MultiError, len(key))
	for i, k := range key {
		if !k.valid() {
			err[i] = ErrInvalidKey
		}
	}
	return err
}

// checkMultiArg checks that v has type []S, []*S, []I, or []P, for some struct
// type S, for some interface type I, or some non-interface non-pointer type P
// such that P or *P implements PropertyLoadSaver.
//
// It returns what category the slice's elements are, and the reflect.Type
// that represents S, I or P.
//
// As a special case, PropertyList is an invalid type for v.
func checkMultiArg(v reflect.Value) (m multiArgType, elemType reflect.Type) {
	// TODO(djd): multiArg is very confusing. Fold this logic into the
	// relevant Put/Get methods to make the logic less opaque.
	if v.Kind() != reflect.Slice {
		return multiArgTypeInvalid, nil
	}
	if v.Type() == typeOfPropertyList {
		return multiArgTypeInvalid, nil
	}
	elemType = v.Type().Elem()
	if reflect.PtrTo(elemType).Implements(typeOfPropertyLoadSaver) {
		return multiArgTypePropertyLoadSaver, elemType
	}
	switch elemType.Kind() {
	case reflect.Struct:
		return multiArgTypeStruct, elemType
	case reflect.Interface:
		return multiArgTypeInterface, elemType
	case reflect.Ptr:
		elemType = elemType.Elem()
		if elemType.Kind() == reflect.Struct {
			return multiArgTypeStructPtr, elemType
		}
	}
	return multiArgTypeInvalid, nil
}

// processFieldMismatchError ignore FieldMismatchErr if WithIgnoreFieldMismatch client option is provided by user
func (c *Client) processFieldMismatchError(err error) error {
	if c.config == nil || !c.config.ignoreFieldMismatchErrors {
		return err
	}
	return ignoreFieldMismatchErrs(err)
}

func ignoreFieldMismatchErrs(err error) error {
	if err == nil {
		return err
	}

	multiErr, isMultiErr := err.(MultiError)
	if isMultiErr {
		foundErr := false
		for i, e := range multiErr {
			multiErr[i] = ignoreFieldMismatchErr(e)
			if multiErr[i] != nil {
				foundErr = true
			}
		}
		if !foundErr {
			return nil
		}
		return multiErr
	}

	return ignoreFieldMismatchErr(err)
}

func ignoreFieldMismatchErr(err error) error {
	if err == nil {
		return err
	}
	_, isFieldMismatchErr := err.(*ErrFieldMismatch)
	if isFieldMismatchErr {
		return nil
	}
	return err
}

// Close closes the Client. Call Close to clean up resources when done with the
// Client.
func (c *Client) Close() error {
	return c.connPool.Close()
}

// Get loads the entity stored for key into dst, which must be a struct pointer
// or implement PropertyLoadSaver. If there is no such entity for the key, Get
// returns ErrNoSuchEntity.
//
// The values of dst's unmatched struct fields are not modified, and matching
// slice-typed fields are not reset before appending to them. In particular, it
// is recommended to pass a pointer to a zero valued struct on each Get call.
//
// ErrFieldMismatch is returned when a field is to be loaded into a different
// type than the one it was stored from, or when a field is missing or
// unexported in the destination struct. ErrFieldMismatch is only returned if
// dst is a struct pointer.
func (c *Client) Get(ctx context.Context, key *Key, dst interface{}) (err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/datastore.Get")
	defer func() { trace.EndSpan(ctx, err) }()

	if dst == nil { // get catches nil interfaces; we need to catch nil ptr here
		return fmt.Errorf("%w: dst cannot be nil", ErrInvalidEntityType)
	}

	var opts *pb.ReadOptions
	if c.readSettings.readTimeExists() {
		opts = &pb.ReadOptions{
			ConsistencyType: &pb.ReadOptions_ReadTime{
				// Timestamp cannot be less than microseconds accuracy. See #6938
				ReadTime: &timestamppb.Timestamp{Seconds: c.readSettings.readTime.Unix()},
			},
		}
	}

	// Since opts does not contain Transaction option, 'get' call below will return nil
	// as transaction id which can be ignored
	_, err = c.get(ctx, []*Key{key}, []interface{}{dst}, opts)
	if me, ok := err.(MultiError); ok {
		return c.processFieldMismatchError(me[0])
	}
	return c.processFieldMismatchError(err)
}

// GetMulti is a batch version of Get.
//
// dst must be a []S, []*S, []I or []P, for some struct type S, some interface
// type I, or some non-interface non-pointer type P such that P or *P
// implements PropertyLoadSaver. If an []I, each element must be a valid dst
// for Get: it must be a struct pointer or implement PropertyLoadSaver.
//
// As a special case, PropertyList is an invalid type for dst, even though a
// PropertyList is a slice of structs. It is treated as invalid to avoid being
// mistakenly passed when []PropertyList was intended.
//
// err may be a MultiError. See ExampleMultiError to check it.
func (c *Client) GetMulti(ctx context.Context, keys []*Key, dst interface{}) (err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/datastore.GetMulti")
	defer func() { trace.EndSpan(ctx, err) }()

	var opts *pb.ReadOptions
	if c.readSettings.readTimeExists() {
		opts = &pb.ReadOptions{
			ConsistencyType: &pb.ReadOptions_ReadTime{
				// Timestamp cannot be less than microseconds accuracy. See #6938
				ReadTime: &timestamppb.Timestamp{Seconds: c.readSettings.readTime.Unix()},
			},
		}
	}

	// Since opts does not contain Transaction option, 'get' call below will return nil
	// as transaction id which can be ignored
	_, err = c.get(ctx, keys, dst, opts)
	return c.processFieldMismatchError(err)
}

func (c *Client) get(ctx context.Context, keys []*Key, dst interface{}, opts *pb.ReadOptions) ([]byte, error) {
	v := reflect.ValueOf(dst)

	var multiArgType multiArgType

	// If kind is of type slice, return error
	if kind := v.Kind(); kind != reflect.Slice {
		return nil, fmt.Errorf("%w: dst: expected slice got %v", ErrInvalidEntityType, kind.String())
	}

	// if type is a type which implements PropertyList, return error
	if argType := v.Type(); argType == typeOfPropertyList {
		return nil, fmt.Errorf("%w: dst: cannot be PropertyListType", ErrInvalidEntityType)
	}

	elemType := v.Type().Elem()
	if reflect.PtrTo(elemType).Implements(typeOfPropertyLoadSaver) {
		multiArgType = multiArgTypePropertyLoadSaver
	}

	switch elemType.Kind() {
	case reflect.Struct:
		multiArgType = multiArgTypeStruct
	case reflect.Interface:
		multiArgType = multiArgTypeInterface
	case reflect.Ptr:
		elemType = elemType.Elem()
		if elemType.Kind() == reflect.Struct {
			multiArgType = multiArgTypeStructPtr
		}
	}

	dstLen := v.Len()

	if keysLen := len(keys); keysLen != dstLen {
		return nil, fmt.Errorf("%w: key length = %d, dst length = %d", ErrDifferentKeyAndDstLength, keysLen, v.Len())
	}

	if len(keys) == 0 {
		return nil, nil
	}

	// Go through keys, validate them, serialize then, and create a dict mapping them to their indices.
	// Equal keys are deduped.
	multiErr, any := make(MultiError, len(keys)), false
	keyMap := make(map[string][]int, len(keys))
	pbKeys := make([]*pb.Key, 0, len(keys))
	for i, k := range keys {
		if !k.valid() {
			multiErr[i] = ErrInvalidKey
			any = true
		} else if k.Incomplete() {
			multiErr[i] = fmt.Errorf("datastore: can't get the incomplete key: %v", k)
			any = true
		} else {
			ks := k.stringInternal()
			if _, ok := keyMap[ks]; !ok {
				pbKeys = append(pbKeys, keyToProto(k))
			}
			keyMap[ks] = append(keyMap[ks], i)
		}
	}
	if any {
		return nil, multiErr
	}
	req := &pb.LookupRequest{
		ProjectId:   c.dataset,
		DatabaseId:  c.databaseID,
		Keys:        pbKeys,
		ReadOptions: opts,
	}
	resp, err := c.client.Lookup(ctx, req)

	if err != nil {
		return nil, err
	}
	txnID := resp.Transaction
	found := resp.Found
	missing := resp.Missing
	// Upper bound 1000 iterations to prevent infinite loop. This matches the max
	// number of Entities you can request from Datastore.
	// Note that if ctx has a deadline, the deadline will probably
	// be hit before we reach 1000 iterations.
	for i := 0; len(resp.Deferred) > 0 && i < 1000; i++ {
		req.Keys = resp.Deferred
		resp, err = c.client.Lookup(ctx, req)
		if err != nil {
			return txnID, err
		}
		found = append(found, resp.Found...)
		missing = append(missing, resp.Missing...)
	}

	filled := 0
	for _, e := range found {
		k, err := protoToKey(e.Entity.Key)
		if err != nil {
			return txnID, errors.New("datastore: internal error: server returned an invalid key")
		}
		filled += len(keyMap[k.stringInternal()])
		for _, index := range keyMap[k.stringInternal()] {
			elem := v.Index(index)
			if multiArgType == multiArgTypePropertyLoadSaver || multiArgType == multiArgTypeStruct {
				elem = elem.Addr()
			}
			if multiArgType == multiArgTypeStructPtr && elem.IsNil() {
				elem.Set(reflect.New(elem.Type().Elem()))
			}
			if err := loadEntityProto(elem.Interface(), e.Entity); err != nil {
				multiErr[index] = err
				any = true
			}
		}
	}
	for _, e := range missing {
		k, err := protoToKey(e.Entity.Key)
		if err != nil {
			return txnID, errors.New("datastore: internal error: server returned an invalid key")
		}
		filled += len(keyMap[k.stringInternal()])
		for _, index := range keyMap[k.stringInternal()] {
			multiErr[index] = ErrNoSuchEntity
		}
		any = true
	}

	if filled != len(keys) {
		return txnID, errors.New("datastore: internal error: server returned the wrong number of entities")
	}

	if any {
		return txnID, multiErr
	}
	return txnID, nil
}

// Put saves the entity src into the datastore with the given key. src must be
// a struct pointer or implement PropertyLoadSaver; if the struct pointer has
// any unexported fields they will be skipped. If the key is incomplete, the
// returned key will be a unique key generated by the datastore.
func (c *Client) Put(ctx context.Context, key *Key, src interface{}) (*Key, error) {
	k, err := c.PutMulti(ctx, []*Key{key}, []interface{}{src})
	if err != nil {
		if me, ok := err.(MultiError); ok {
			return nil, me[0]
		}
		return nil, err
	}
	return k[0], nil
}

// PutMulti is a batch version of Put.
//
// src must satisfy the same conditions as the dst argument to GetMulti.
// err may be a MultiError. See ExampleMultiError to check it.
func (c *Client) PutMulti(ctx context.Context, keys []*Key, src interface{}) (ret []*Key, err error) {
	// TODO(jba): rewrite in terms of Mutate.
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/datastore.PutMulti")
	defer func() { trace.EndSpan(ctx, err) }()

	mutations, err := putMutations(keys, src)
	if err != nil {
		return nil, err
	}

	// Make the request.
	req := &pb.CommitRequest{
		ProjectId:  c.dataset,
		DatabaseId: c.databaseID,
		Mutations:  mutations,
		Mode:       pb.CommitRequest_NON_TRANSACTIONAL,
	}
	resp, err := c.client.Commit(ctx, req)
	if err != nil {
		return nil, err
	}

	// Copy any newly minted keys into the returned keys.
	ret = make([]*Key, len(keys))
	for i, key := range keys {
		if key.Incomplete() {
			// This key is in the mutation results.
			ret[i], err = protoToKey(resp.MutationResults[i].Key)
			if err != nil {
				return nil, errors.New("datastore: internal error: server returned an invalid key")
			}
		} else {
			ret[i] = key
		}
	}
	return ret, nil
}

func putMutations(keys []*Key, src interface{}) ([]*pb.Mutation, error) {
	v := reflect.ValueOf(src)
	var multiArgType multiArgType

	// If kind is of type slice, return error
	if kind := v.Kind(); kind != reflect.Slice {
		return nil, fmt.Errorf("%w: dst: expected slice got %v", ErrInvalidEntityType, kind.String())
	}

	// if type is a type which implements PropertyList, return error
	if argType := v.Type(); argType == typeOfPropertyList {
		return nil, fmt.Errorf("%w: dst: cannot be PropertyListType", ErrInvalidEntityType)
	}

	elemType := v.Type().Elem()
	if reflect.PtrTo(elemType).Implements(typeOfPropertyLoadSaver) {
		multiArgType = multiArgTypePropertyLoadSaver
	}

	switch elemType.Kind() {
	case reflect.Struct:
		multiArgType = multiArgTypeStruct
	case reflect.Interface:
		multiArgType = multiArgTypeInterface
	case reflect.Ptr:
		elemType = elemType.Elem()
		if elemType.Kind() == reflect.Struct {
			multiArgType = multiArgTypeStructPtr
		}
	}
	if multiArgType == multiArgTypeInvalid {
		return nil, fmt.Errorf("datastore: src has invalid type")
	}
	dstLen := v.Len()

	if keysLen := len(keys); keysLen != dstLen {
		return nil, fmt.Errorf("%w: key length = %d, dst length = %d", ErrDifferentKeyAndDstLength, keysLen, v.Len())
	}

	if len(keys) == 0 {
		return nil, nil
	}
	if err := multiValid(keys); err != nil {
		return nil, err
	}
	mutations := make([]*pb.Mutation, 0, len(keys))
	multiErr := make(MultiError, len(keys))
	hasErr := false
	for i, k := range keys {
		elem := v.Index(i)
		// Two cases where we need to take the address:
		// 1) multiArgTypePropertyLoadSaver => &elem implements PLS
		// 2) multiArgTypeStruct => saveEntity needs *struct
		if multiArgType == multiArgTypePropertyLoadSaver || multiArgType == multiArgTypeStruct {
			elem = elem.Addr()
		}
		p, err := saveEntity(k, elem.Interface())
		if err != nil {
			multiErr[i] = err
			hasErr = true
		}
		var mut *pb.Mutation
		if k.Incomplete() {
			mut = &pb.Mutation{Operation: &pb.Mutation_Insert{Insert: p}}
		} else {
			mut = &pb.Mutation{Operation: &pb.Mutation_Upsert{Upsert: p}}
		}
		mutations = append(mutations, mut)
	}
	if hasErr {
		return nil, multiErr
	}
	return mutations, nil
}

// Delete deletes the entity for the given key.
func (c *Client) Delete(ctx context.Context, key *Key) error {
	err := c.DeleteMulti(ctx, []*Key{key})
	if me, ok := err.(MultiError); ok {
		return me[0]
	}
	return err
}

// DeleteMulti is a batch version of Delete.
//
// err may be a MultiError. See ExampleMultiError to check it.
func (c *Client) DeleteMulti(ctx context.Context, keys []*Key) (err error) {
	// TODO(jba): rewrite in terms of Mutate.
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/datastore.DeleteMulti")
	defer func() { trace.EndSpan(ctx, err) }()

	mutations, err := deleteMutations(keys)
	if err != nil {
		return err
	}

	req := &pb.CommitRequest{
		ProjectId:  c.dataset,
		DatabaseId: c.databaseID,
		Mutations:  mutations,
		Mode:       pb.CommitRequest_NON_TRANSACTIONAL,
	}
	_, err = c.client.Commit(ctx, req)
	return err
}

func deleteMutations(keys []*Key) ([]*pb.Mutation, error) {
	mutations := make([]*pb.Mutation, 0, len(keys))
	set := make(map[string]bool, len(keys))
	multiErr := make(MultiError, len(keys))
	hasErr := false
	for i, k := range keys {
		if !k.valid() {
			multiErr[i] = ErrInvalidKey
			hasErr = true
		} else if k.Incomplete() {
			multiErr[i] = fmt.Errorf("datastore: can't delete the incomplete key: %v", k)
			hasErr = true
		} else {
			ks := k.stringInternal()
			if !set[ks] {
				mutations = append(mutations, &pb.Mutation{
					Operation: &pb.Mutation_Delete{Delete: keyToProto(k)},
				})
			}
			set[ks] = true
		}
	}
	if hasErr {
		return nil, multiErr
	}
	return mutations, nil
}

// Mutate applies one or more mutations. Mutations are applied in
// non-transactional mode. If you need atomicity, use Transaction.Mutate.
// It returns the keys of the argument Mutations, in the same order.
//
// If any of the mutations are invalid, Mutate returns a MultiError with the errors.
// Mutate returns a MultiError in this case even if there is only one Mutation.
// See ExampleMultiError to check it.
func (c *Client) Mutate(ctx context.Context, muts ...*Mutation) (ret []*Key, err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/datastore.Mutate")
	defer func() { trace.EndSpan(ctx, err) }()

	pmuts, err := mutationProtos(muts)
	if err != nil {
		return nil, err
	}
	req := &pb.CommitRequest{
		ProjectId:  c.dataset,
		DatabaseId: c.databaseID,
		Mutations:  pmuts,
		Mode:       pb.CommitRequest_NON_TRANSACTIONAL,
	}
	resp, err := c.client.Commit(ctx, req)
	if err != nil {
		return nil, err
	}
	// Copy any newly minted keys into the returned keys.
	ret = make([]*Key, len(muts))
	for i, mut := range muts {
		if mut.key.Incomplete() {
			// This key is in the mutation results.
			ret[i], err = protoToKey(resp.MutationResults[i].Key)
			if err != nil {
				return nil, errors.New("datastore: internal error: server returned an invalid key")
			}
		} else {
			ret[i] = mut.key
		}
	}
	return ret, nil
}

// ReadTime specifies a snapshot (time) of the database to read.
func ReadTime(t time.Time) ReadOption {
	return docReadTime(t)
}

type docReadTime time.Time

func (drt docReadTime) apply(rs *readSettings) {
	rs.readTime = time.Time(drt)
}

// ReadOption provides specific instructions for how to access documents in the database.
// Currently, only ReadTime is supported.
type ReadOption interface {
	apply(*readSettings)
}

type readSettings struct {
	readTime time.Time
}

func (rs *readSettings) readTimeExists() bool {
	return rs != nil && !rs.readTime.IsZero()
}

// WithReadOptions specifies constraints for accessing documents from the database,
// e.g. at what time snapshot to read the documents.
// The client uses this value for subsequent reads, unless additional ReadOptions
// are provided.
func (c *Client) WithReadOptions(ro ...ReadOption) *Client {
	for _, r := range ro {
		r.apply(c.readSettings)
	}
	return c
}
