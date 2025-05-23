// Copyright 2017 Google LLC
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

package firestore

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	pb "cloud.google.com/go/firestore/apiv1/firestorepb"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/googleapis/type/latlng"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	writeResultForSet    = &WriteResult{UpdateTime: aTime}
	commitResponseForSet = &pb.CommitResponse{
		WriteResults: []*pb.WriteResult{{UpdateTime: aTimestamp}},
	}
)

func TestDocGet(t *testing.T) {
	ctx := context.Background()
	c, srv, cleanup := newMock(t)
	defer cleanup()

	path := "projects/projectID/databases/(default)/documents/C/a"
	pdoc := &pb.Document{
		Name:       path,
		CreateTime: aTimestamp,
		UpdateTime: aTimestamp,
		Fields:     map[string]*pb.Value{"f": intval(1)},
	}
	srv.addRPC(&pb.BatchGetDocumentsRequest{
		Database:  c.path(),
		Documents: []string{path},
	}, []interface{}{
		&pb.BatchGetDocumentsResponse{
			Result:   &pb.BatchGetDocumentsResponse_Found{Found: pdoc},
			ReadTime: aTimestamp2,
		},
	})
	ref := c.Collection("C").Doc("a")
	gotDoc, err := ref.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	wantDoc := &DocumentSnapshot{
		Ref:        ref,
		CreateTime: aTime,
		UpdateTime: aTime,
		ReadTime:   aTime2,
		proto:      pdoc,
		c:          c,
	}
	if !testEqual(gotDoc, wantDoc) {
		t.Fatalf("\ngot  %+v\nwant %+v", gotDoc, wantDoc)
	}

	path2 := "projects/projectID/databases/(default)/documents/C/b"
	srv.addRPC(
		&pb.BatchGetDocumentsRequest{
			Database:  c.path(),
			Documents: []string{path2},
		}, []interface{}{
			&pb.BatchGetDocumentsResponse{
				Result:   &pb.BatchGetDocumentsResponse_Missing{Missing: path2},
				ReadTime: aTimestamp3,
			},
		})
	_, err = c.Collection("C").Doc("b").Get(ctx)
	if status.Code(err) != codes.NotFound {
		t.Errorf("got %v, want NotFound", err)
	}

	// Invalid UTF-8 characters
	if _, gotErr := c.Collection("C").Doc("Mayag\xcfez").Get(ctx); !errorsMatch(gotErr, errInvalidUtf8DocRef) {
		t.Errorf("got: %v, want: %v", gotErr, errInvalidUtf8DocRef)
	}

	// nil DocRef
	var nilDocRef *DocumentRef
	if _, gotErr := nilDocRef.Get(ctx); !errorsMatch(gotErr, errNilDocRef) {
		t.Errorf("got: %v, want: %v", gotErr, errInvalidUtf8DocRef)
	}
}

func TestDocSet(t *testing.T) {
	// Most tests for Set are in the conformance tests.
	ctx := context.Background()
	c, srv, cleanup := newMock(t)
	defer cleanup()

	doc := c.Collection("C").Doc("d")
	// Merge with a struct and FieldPaths.
	srv.addRPC(&pb.CommitRequest{
		Database: "projects/projectID/databases/(default)",
		Writes: []*pb.Write{
			{
				Operation: &pb.Write_Update{
					Update: &pb.Document{
						Name: "projects/projectID/databases/(default)/documents/C/d",
						Fields: map[string]*pb.Value{
							"*": mapval(map[string]*pb.Value{
								"~": boolval(true),
							}),
						},
					},
				},
				UpdateMask: &pb.DocumentMask{FieldPaths: []string{"`*`.`~`"}},
			},
		},
	}, commitResponseForSet)
	data := struct {
		A map[string]bool `firestore:"*"`
	}{A: map[string]bool{"~": true}}
	wr, err := doc.Set(ctx, data, Merge([]string{"*", "~"}))
	if err != nil {
		t.Fatal(err)
	}
	if !testEqual(wr, writeResultForSet) {
		t.Errorf("got %v, want %v", wr, writeResultForSet)
	}

	// MergeAll cannot be used with structs.
	_, err = doc.Set(ctx, data, MergeAll)
	if err == nil {
		t.Errorf("got nil, want error")
	}

	// Invalid UTF-8 characters
	if _, gotErr := c.Collection("C").Doc("Mayag\xcfez").
		Set(ctx, data, Merge([]string{"*", "~"})); !errorsMatch(gotErr, errInvalidUtf8DocRef) {
		t.Errorf("got: %v, want: %v", gotErr, errInvalidUtf8DocRef)
	}

	// nil DocRef
	var nilDocRef *DocumentRef
	if _, gotErr := nilDocRef.Set(ctx, data, Merge([]string{"*", "~"})); !errorsMatch(gotErr, errNilDocRef) {
		t.Errorf("got: %v, want: %v", gotErr, errInvalidUtf8DocRef)
	}
}

func TestDocCreate(t *testing.T) {
	// Verify creation with structs. In particular, make sure zero values
	// are handled well.
	// Other tests for Create are handled by the conformance tests.
	ctx := context.Background()
	c, srv, cleanup := newMock(t)
	defer cleanup()

	type create struct {
		Time  time.Time
		Bytes []byte
		Geo   *latlng.LatLng
	}
	srv.addRPC(
		&pb.CommitRequest{
			Database: "projects/projectID/databases/(default)",
			Writes: []*pb.Write{
				{
					Operation: &pb.Write_Update{
						Update: &pb.Document{
							Name: "projects/projectID/databases/(default)/documents/C/d",
							Fields: map[string]*pb.Value{
								"Time":  tsval(time.Time{}),
								"Bytes": bytesval(nil),
								"Geo":   nullValue,
							},
						},
					},
					CurrentDocument: &pb.Precondition{
						ConditionType: &pb.Precondition_Exists{Exists: false},
					},
				},
			},
		},
		commitResponseForSet,
	)
	_, err := c.Collection("C").Doc("d").Create(ctx, &create{})
	if err != nil {
		t.Fatal(err)
	}

	// Invalid UTF-8 characters
	if _, gotErr := c.Collection("C").Doc("Mayag\xcfez").
		Create(ctx, &create{}); !errorsMatch(gotErr, errInvalidUtf8DocRef) {
		t.Errorf("got: %v, want: %v", gotErr, errInvalidUtf8DocRef)
	}

	// nil DocRef
	var nilDocRef *DocumentRef
	if _, gotErr := nilDocRef.Create(ctx, &create{}); !errorsMatch(gotErr, errNilDocRef) {
		t.Errorf("got: %v, want: %v", gotErr, errInvalidUtf8DocRef)
	}
}

func TestDocDelete(t *testing.T) {
	ctx := context.Background()
	c, srv, cleanup := newMock(t)
	defer cleanup()

	srv.addRPC(
		&pb.CommitRequest{
			Database: "projects/projectID/databases/(default)",
			Writes: []*pb.Write{
				{Operation: &pb.Write_Delete{Delete: "projects/projectID/databases/(default)/documents/C/d"}},
			},
		},
		&pb.CommitResponse{
			WriteResults: []*pb.WriteResult{{}},
		})
	wr, err := c.Collection("C").Doc("d").Delete(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !testEqual(wr, &WriteResult{}) {
		t.Errorf("got %+v, want %+v", wr, writeResultForSet)
	}

	// Invalid UTF-8 characters
	if _, gotErr := c.Collection("C").Doc("Mayag\xcfez").
		Delete(ctx); !errorsMatch(gotErr, errInvalidUtf8DocRef) {
		t.Errorf("got: %v, want: %v", gotErr, errInvalidUtf8DocRef)
	}

	// nil DocRef
	var nilDocRef *DocumentRef
	if _, gotErr := nilDocRef.Delete(ctx); !errorsMatch(gotErr, errNilDocRef) {
		t.Errorf("got: %v, want: %v", gotErr, errInvalidUtf8DocRef)
	}
}

var (
	testData   = map[string]interface{}{"a": 1}
	testFields = map[string]*pb.Value{"a": intval(1)}
)

// Update is tested by the conformance tests.

func TestFPVsFromData(t *testing.T) {
	type S struct{ X int }

	for _, test := range []struct {
		in   interface{}
		want []fpv
	}{
		{
			in:   nil,
			want: []fpv{{nil, nil}},
		},
		{
			in:   map[string]interface{}{"a": nil},
			want: []fpv{{[]string{"a"}, nil}},
		},
		{
			in:   map[string]interface{}{"a": 1},
			want: []fpv{{[]string{"a"}, 1}},
		},
		{
			in: map[string]interface{}{
				"a": 1,
				"b": map[string]interface{}{"c": 2},
			},
			want: []fpv{{[]string{"a"}, 1}, {[]string{"b", "c"}, 2}},
		},
		{
			in:   map[string]interface{}{"s": &S{X: 3}},
			want: []fpv{{[]string{"s"}, &S{X: 3}}},
		},
	} {
		var got []fpv
		fpvsFromData(reflect.ValueOf(test.in), nil, &got)
		sort.Sort(byFieldPath(got))
		if !testEqual(got, test.want) {
			t.Errorf("%+v: got %v, want %v", test.in, got, test.want)
		}
	}
}

type byFieldPath []fpv

func (b byFieldPath) Len() int           { return len(b) }
func (b byFieldPath) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byFieldPath) Less(i, j int) bool { return b[i].fieldPath.less(b[j].fieldPath) }

func commitRequestForSet() *pb.CommitRequest {
	return &pb.CommitRequest{
		Database: "projects/projectID/databases/(default)",
		Writes: []*pb.Write{
			{
				Operation: &pb.Write_Update{
					Update: &pb.Document{
						Name:   "projects/projectID/databases/(default)/documents/C/d",
						Fields: testFields,
					},
				},
			},
		},
	}
}

func TestUpdateProcess(t *testing.T) {
	for _, test := range []struct {
		desc    string
		in      Update
		want    fpv
		wantErr bool
		wantStr string
	}{
		{
			desc:    "Integer value",
			in:      Update{Path: "a", Value: 1},
			want:    fpv{fieldPath: []string{"a"}, value: 1},
			wantStr: "{Path:a FieldPath:[] Value:%!s(int=1)}",
		},
		{
			desc:    "Delete transform",
			in:      Update{Path: "c.d", Value: Delete},
			want:    fpv{fieldPath: []string{"c", "d"}, value: Delete},
			wantStr: "{Path:c.d FieldPath:[] Value:Delete}",
		},
		{
			desc: "Increment transform",
			in:   Update{Path: "c.d", Value: Increment(8)},
			want: fpv{fieldPath: []string{"c", "d"}, value: transform{
				t: &pb.DocumentTransform_FieldTransform{
					TransformType: &pb.DocumentTransform_FieldTransform_Increment{
						Increment: &pb.Value{
							ValueType: &pb.Value_IntegerValue{IntegerValue: 8},
						},
					},
				},
				err: nil,
			}},
			wantStr: "{Path:c.d FieldPath:[] Value:{t:increment:{integer_value:8}}}",
		},
		{
			desc:    "ServerTimestamp transform",
			in:      Update{FieldPath: []string{"*", "~"}, Value: ServerTimestamp},
			want:    fpv{fieldPath: []string{"*", "~"}, value: ServerTimestamp},
			wantStr: "{Path: FieldPath:[* ~] Value:ServerTimestamp}",
		},
		{
			desc:    "bad rune in path",
			in:      Update{Path: "*"},
			wantErr: true, // bad rune in path
			wantStr: "{Path:* FieldPath:[] Value:%!s(<nil>)}",
		},
		{
			desc:    "both Path and FieldPath",
			in:      Update{Path: "a", FieldPath: []string{"b"}},
			wantErr: true, // both Path and FieldPath
			wantStr: "{Path:a FieldPath:[b] Value:%!s(<nil>)}",
		},
		{
			desc:    "neither Path nor FieldPath",
			in:      Update{Value: 1},
			wantErr: true, // neither Path nor FieldPath
			wantStr: "{Path: FieldPath:[] Value:%!s(int=1)}",
		},
		{
			desc:    "empty FieldPath component",
			in:      Update{FieldPath: []string{"", "a"}},
			wantErr: true, // empty FieldPath component
			wantStr: "{Path: FieldPath:[ a] Value:%!s(<nil>)}",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			got, err := test.in.process()
			if test.wantErr {
				if err == nil {
					t.Errorf("%+v: got nil, want error", test.in)
				}
			} else if err != nil {
				t.Errorf("%+v: got error %v, want nil", test.in, err)
			} else if !testEqual(got, test.want) {
				t.Errorf("%+v: got %+v, want %+v", test.in, got, test.want)
			}

			gotStr := test.in.String()
			if gotStr != test.wantStr {
				t.Errorf("%+v: got %q, want %q", test.in, gotStr, test.wantStr)
			}
		})
	}
}

func TestDocRef_WithReadOptions(t *testing.T) {
	ctx := context.Background()
	c, srv, cleanup := newMock(t)
	defer cleanup()

	const dbPath = "projects/projectID/databases/(default)"
	const docPath = dbPath + "/documents/C/a"
	tm := time.Date(2021, time.February, 20, 0, 0, 0, 0, time.UTC)

	srv.addRPC(&pb.ListDocumentsRequest{
		Parent:       dbPath + "/documents",
		CollectionId: "myCollection",
		Mask:         &pb.DocumentMask{},
		ShowMissing:  true,
	}, []interface{}{
		&pb.ListDocumentsResponse{
			Documents: []*pb.Document{
				{
					Name:       dbPath + "/documents/C/a",
					CreateTime: &timestamppb.Timestamp{Seconds: 10},
					UpdateTime: &timestamppb.Timestamp{Seconds: 20},
					Fields:     map[string]*pb.Value{"a": intval(1)},
				},
			},
		},
	})
	srv.addRPC(&pb.BatchGetDocumentsRequest{
		Database:  dbPath,
		Documents: []string{docPath},
		ConsistencySelector: &pb.BatchGetDocumentsRequest_ReadTime{
			ReadTime: &timestamppb.Timestamp{Seconds: tm.Unix()},
		},
	}, []interface{}{
		&pb.BatchGetDocumentsResponse{
			ReadTime: &timestamppb.Timestamp{Seconds: tm.Unix()},
			Result: &pb.BatchGetDocumentsResponse_Found{
				Found: &pb.Document{
					Name:       dbPath + "/documents/C/a",
					CreateTime: &timestamppb.Timestamp{Seconds: 10},
					UpdateTime: &timestamppb.Timestamp{Seconds: 20},
					Fields:     map[string]*pb.Value{"a": intval(1)},
				},
			},
		},
	})

	it := c.Collection("myCollection").DocumentRefs(ctx)

	for {
		dr, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		_, err = dr.WithReadOptions(ReadTime(tm)).Get(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}

}
