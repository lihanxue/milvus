// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datacoord

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	"github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/schemapb"
	"github.com/milvus-io/milvus/internal/metastore"
	catalogmocks "github.com/milvus-io/milvus/internal/metastore/mocks"
	"github.com/milvus-io/milvus/internal/metastore/model"
	"github.com/milvus-io/milvus/internal/mocks"
	"github.com/milvus-io/milvus/internal/proto/datapb"
	"github.com/milvus-io/milvus/internal/proto/indexpb"
	"github.com/milvus-io/milvus/pkg/common"
	"github.com/milvus-io/milvus/pkg/util/merr"
	"github.com/milvus-io/milvus/pkg/util/paramtable"
)

var (
	collID    = UniqueID(100)
	partID    = UniqueID(200)
	indexID   = UniqueID(300)
	fieldID   = UniqueID(400)
	indexName = "_default_idx"
	segID     = UniqueID(500)
	buildID   = UniqueID(600)
	nodeID    = UniqueID(700)
)

func createIndexMeta(catalog metastore.DataCoordCatalog) *indexMeta {
	return &indexMeta{
		catalog: catalog,
		indexes: map[UniqueID]map[UniqueID]*model.Index{
			collID: {
				indexID: {
					TenantID:     "",
					CollectionID: collID,
					FieldID:      fieldID,
					IndexID:      indexID,
					IndexName:    indexName,
					IsDeleted:    false,
					CreateTime:   1,
					TypeParams: []*commonpb.KeyValuePair{
						{
							Key:   common.DimKey,
							Value: "128",
						},
					},
					IndexParams: []*commonpb.KeyValuePair{
						{
							Key:   common.MetricTypeKey,
							Value: "L2",
						},
					},
				},
			},
		},
		segmentIndexes: map[UniqueID]map[UniqueID]*model.SegmentIndex{
			segID: {
				indexID: {
					SegmentID:     segID,
					CollectionID:  collID,
					PartitionID:   partID,
					NumRows:       1025,
					IndexID:       indexID,
					BuildID:       buildID,
					NodeID:        0,
					IndexVersion:  0,
					IndexState:    commonpb.IndexState_Unissued,
					FailReason:    "",
					IsDeleted:     false,
					CreateTime:    0,
					IndexFileKeys: nil,
					IndexSize:     0,
				},
			},
			segID + 1: {
				indexID: {
					SegmentID:     segID + 1,
					CollectionID:  collID,
					PartitionID:   partID,
					NumRows:       1026,
					IndexID:       indexID,
					BuildID:       buildID + 1,
					NodeID:        nodeID,
					IndexVersion:  1,
					IndexState:    commonpb.IndexState_InProgress,
					FailReason:    "",
					IsDeleted:     false,
					CreateTime:    1111,
					IndexFileKeys: nil,
					IndexSize:     0,
				},
			},
			segID + 2: {
				indexID: {
					SegmentID:     segID + 2,
					CollectionID:  collID,
					PartitionID:   partID,
					NumRows:       1026,
					IndexID:       indexID,
					BuildID:       buildID + 2,
					NodeID:        nodeID,
					IndexVersion:  1,
					IndexState:    commonpb.IndexState_InProgress,
					FailReason:    "",
					IsDeleted:     true,
					CreateTime:    1111,
					IndexFileKeys: nil,
					IndexSize:     0,
				},
			},
			segID + 3: {
				indexID: {
					SegmentID:     segID + 3,
					CollectionID:  collID,
					PartitionID:   partID,
					NumRows:       500,
					IndexID:       indexID,
					BuildID:       buildID + 3,
					NodeID:        0,
					IndexVersion:  0,
					IndexState:    commonpb.IndexState_Unissued,
					FailReason:    "",
					IsDeleted:     false,
					CreateTime:    1111,
					IndexFileKeys: nil,
					IndexSize:     0,
				},
			},
			segID + 4: {
				indexID: {
					SegmentID:     segID + 4,
					CollectionID:  collID,
					PartitionID:   partID,
					NumRows:       1026,
					IndexID:       indexID,
					BuildID:       buildID + 4,
					NodeID:        nodeID,
					IndexVersion:  1,
					IndexState:    commonpb.IndexState_Finished,
					FailReason:    "",
					IsDeleted:     false,
					CreateTime:    1111,
					IndexFileKeys: nil,
					IndexSize:     0,
				},
			},
			segID + 5: {
				indexID: {
					SegmentID:     segID + 5,
					CollectionID:  collID,
					PartitionID:   partID,
					NumRows:       1026,
					IndexID:       indexID,
					BuildID:       buildID + 5,
					NodeID:        0,
					IndexVersion:  1,
					IndexState:    commonpb.IndexState_Finished,
					FailReason:    "",
					IsDeleted:     false,
					CreateTime:    1111,
					IndexFileKeys: nil,
					IndexSize:     0,
				},
			},
			segID + 6: {
				indexID: {
					SegmentID:     segID + 6,
					CollectionID:  collID,
					PartitionID:   partID,
					NumRows:       1026,
					IndexID:       indexID,
					BuildID:       buildID + 6,
					NodeID:        0,
					IndexVersion:  1,
					IndexState:    commonpb.IndexState_Finished,
					FailReason:    "",
					IsDeleted:     false,
					CreateTime:    1111,
					IndexFileKeys: nil,
					IndexSize:     0,
				},
			},
			segID + 7: {
				indexID: {
					SegmentID:     segID + 7,
					CollectionID:  collID,
					PartitionID:   partID,
					NumRows:       1026,
					IndexID:       indexID,
					BuildID:       buildID + 7,
					NodeID:        0,
					IndexVersion:  1,
					IndexState:    commonpb.IndexState_Failed,
					FailReason:    "error",
					IsDeleted:     false,
					CreateTime:    1111,
					IndexFileKeys: nil,
					IndexSize:     0,
				},
			},
			segID + 8: {
				indexID: {
					SegmentID:     segID + 8,
					CollectionID:  collID,
					PartitionID:   partID,
					NumRows:       1026,
					IndexID:       indexID,
					BuildID:       buildID + 8,
					NodeID:        nodeID + 1,
					IndexVersion:  1,
					IndexState:    commonpb.IndexState_InProgress,
					FailReason:    "",
					IsDeleted:     false,
					CreateTime:    1111,
					IndexFileKeys: nil,
					IndexSize:     0,
				},
			},
			segID + 9: {
				indexID: {
					SegmentID:     segID + 9,
					CollectionID:  collID,
					PartitionID:   partID,
					NumRows:       500,
					IndexID:       indexID,
					BuildID:       buildID + 9,
					NodeID:        0,
					IndexVersion:  0,
					IndexState:    commonpb.IndexState_Unissued,
					FailReason:    "",
					IsDeleted:     false,
					CreateTime:    1111,
					IndexFileKeys: nil,
					IndexSize:     0,
				},
			},
			segID + 10: {
				indexID: {
					SegmentID:     segID + 10,
					CollectionID:  collID,
					PartitionID:   partID,
					NumRows:       500,
					IndexID:       indexID,
					BuildID:       buildID + 10,
					NodeID:        nodeID,
					IndexVersion:  0,
					IndexState:    commonpb.IndexState_Unissued,
					FailReason:    "",
					IsDeleted:     false,
					CreateTime:    1111,
					IndexFileKeys: nil,
					IndexSize:     0,
				},
			},
		},
		buildID2SegmentIndex: map[UniqueID]*model.SegmentIndex{
			buildID: {
				SegmentID:     segID,
				CollectionID:  collID,
				PartitionID:   partID,
				NumRows:       1025,
				IndexID:       indexID,
				BuildID:       buildID,
				NodeID:        0,
				IndexVersion:  0,
				IndexState:    commonpb.IndexState_Unissued,
				FailReason:    "",
				IsDeleted:     false,
				CreateTime:    0,
				IndexFileKeys: nil,
				IndexSize:     0,
			},
			buildID + 1: {
				SegmentID:     segID + 1,
				CollectionID:  collID,
				PartitionID:   partID,
				NumRows:       1026,
				IndexID:       indexID,
				BuildID:       buildID + 1,
				NodeID:        nodeID,
				IndexVersion:  1,
				IndexState:    commonpb.IndexState_InProgress,
				FailReason:    "",
				IsDeleted:     false,
				CreateTime:    1111,
				IndexFileKeys: nil,
				IndexSize:     0,
			},
			buildID + 2: {
				SegmentID:     segID + 2,
				CollectionID:  collID,
				PartitionID:   partID,
				NumRows:       1026,
				IndexID:       indexID,
				BuildID:       buildID + 2,
				NodeID:        nodeID,
				IndexVersion:  1,
				IndexState:    commonpb.IndexState_InProgress,
				FailReason:    "",
				IsDeleted:     true,
				CreateTime:    1111,
				IndexFileKeys: nil,
				IndexSize:     0,
			},
			buildID + 3: {
				SegmentID:     segID + 3,
				CollectionID:  collID,
				PartitionID:   partID,
				NumRows:       500,
				IndexID:       indexID,
				BuildID:       buildID + 3,
				NodeID:        0,
				IndexVersion:  0,
				IndexState:    commonpb.IndexState_Unissued,
				FailReason:    "",
				IsDeleted:     false,
				CreateTime:    1111,
				IndexFileKeys: nil,
				IndexSize:     0,
			},
			buildID + 4: {
				SegmentID:     segID + 4,
				CollectionID:  collID,
				PartitionID:   partID,
				NumRows:       1026,
				IndexID:       indexID,
				BuildID:       buildID + 4,
				NodeID:        nodeID,
				IndexVersion:  1,
				IndexState:    commonpb.IndexState_Finished,
				FailReason:    "",
				IsDeleted:     false,
				CreateTime:    1111,
				IndexFileKeys: nil,
				IndexSize:     0,
			},
			buildID + 5: {
				SegmentID:     segID + 5,
				CollectionID:  collID,
				PartitionID:   partID,
				NumRows:       1026,
				IndexID:       indexID,
				BuildID:       buildID + 5,
				NodeID:        0,
				IndexVersion:  1,
				IndexState:    commonpb.IndexState_Finished,
				FailReason:    "",
				IsDeleted:     false,
				CreateTime:    1111,
				IndexFileKeys: nil,
				IndexSize:     0,
			},
			buildID + 6: {
				SegmentID:     segID + 6,
				CollectionID:  collID,
				PartitionID:   partID,
				NumRows:       1026,
				IndexID:       indexID,
				BuildID:       buildID + 6,
				NodeID:        0,
				IndexVersion:  1,
				IndexState:    commonpb.IndexState_Finished,
				FailReason:    "",
				IsDeleted:     false,
				CreateTime:    1111,
				IndexFileKeys: nil,
				IndexSize:     0,
			},
			buildID + 7: {
				SegmentID:     segID + 7,
				CollectionID:  collID,
				PartitionID:   partID,
				NumRows:       1026,
				IndexID:       indexID,
				BuildID:       buildID + 7,
				NodeID:        0,
				IndexVersion:  1,
				IndexState:    commonpb.IndexState_Failed,
				FailReason:    "error",
				IsDeleted:     false,
				CreateTime:    1111,
				IndexFileKeys: nil,
				IndexSize:     0,
			},
			buildID + 8: {
				SegmentID:     segID + 8,
				CollectionID:  collID,
				PartitionID:   partID,
				NumRows:       1026,
				IndexID:       indexID,
				BuildID:       buildID + 8,
				NodeID:        nodeID + 1,
				IndexVersion:  1,
				IndexState:    commonpb.IndexState_InProgress,
				FailReason:    "",
				IsDeleted:     false,
				CreateTime:    1111,
				IndexFileKeys: nil,
				IndexSize:     0,
			},
			buildID + 9: {
				SegmentID:     segID + 9,
				CollectionID:  collID,
				PartitionID:   partID,
				NumRows:       500,
				IndexID:       indexID,
				BuildID:       buildID + 9,
				NodeID:        0,
				IndexVersion:  0,
				IndexState:    commonpb.IndexState_Unissued,
				FailReason:    "",
				IsDeleted:     false,
				CreateTime:    1111,
				IndexFileKeys: nil,
				IndexSize:     0,
			},
			buildID + 10: {
				SegmentID:     segID + 10,
				CollectionID:  collID,
				PartitionID:   partID,
				NumRows:       500,
				IndexID:       indexID,
				BuildID:       buildID + 10,
				NodeID:        nodeID,
				IndexVersion:  0,
				IndexState:    commonpb.IndexState_Unissued,
				FailReason:    "",
				IsDeleted:     false,
				CreateTime:    1111,
				IndexFileKeys: nil,
				IndexSize:     0,
			},
		},
	}
}

func createMeta(catalog metastore.DataCoordCatalog, am *analyzeMeta, im *indexMeta) *meta {
	return &meta{
		catalog: catalog,
		segments: &SegmentsInfo{
			segments: map[UniqueID]*SegmentInfo{
				1000: {
					SegmentInfo: &datapb.SegmentInfo{
						ID:           1000,
						CollectionID: 10000,
						PartitionID:  10001,
						NumOfRows:    3000,
						State:        commonpb.SegmentState_Flushed,
						Binlogs:      []*datapb.FieldBinlog{{FieldID: 10002, Binlogs: []*datapb.Binlog{{LogID: 1}, {LogID: 2}, {LogID: 3}}}},
					},
				},
				1001: {
					SegmentInfo: &datapb.SegmentInfo{
						ID:           1001,
						CollectionID: 10000,
						PartitionID:  10001,
						NumOfRows:    3000,
						State:        commonpb.SegmentState_Flushed,
						Binlogs:      []*datapb.FieldBinlog{{FieldID: 10002, Binlogs: []*datapb.Binlog{{LogID: 1}, {LogID: 2}, {LogID: 3}}}},
					},
				},
				1002: {
					SegmentInfo: &datapb.SegmentInfo{
						ID:           1002,
						CollectionID: 10000,
						PartitionID:  10001,
						NumOfRows:    3000,
						State:        commonpb.SegmentState_Flushed,
						Binlogs:      []*datapb.FieldBinlog{{FieldID: 10002, Binlogs: []*datapb.Binlog{{LogID: 1}, {LogID: 2}, {LogID: 3}}}},
					},
				},
				segID: {
					SegmentInfo: &datapb.SegmentInfo{
						ID:             segID,
						CollectionID:   collID,
						PartitionID:    partID,
						InsertChannel:  "",
						NumOfRows:      1025,
						State:          commonpb.SegmentState_Flushed,
						MaxRowNum:      65536,
						LastExpireTime: 10,
					},
				},
				segID + 1: {
					SegmentInfo: &datapb.SegmentInfo{
						ID:             segID + 1,
						CollectionID:   collID,
						PartitionID:    partID,
						InsertChannel:  "",
						NumOfRows:      1026,
						State:          commonpb.SegmentState_Flushed,
						MaxRowNum:      65536,
						LastExpireTime: 10,
					},
				},
				segID + 2: {
					SegmentInfo: &datapb.SegmentInfo{
						ID:             segID + 2,
						CollectionID:   collID,
						PartitionID:    partID,
						InsertChannel:  "",
						NumOfRows:      1026,
						State:          commonpb.SegmentState_Flushed,
						MaxRowNum:      65536,
						LastExpireTime: 10,
					},
				},
				segID + 3: {
					SegmentInfo: &datapb.SegmentInfo{
						ID:             segID + 3,
						CollectionID:   collID,
						PartitionID:    partID,
						InsertChannel:  "",
						NumOfRows:      500,
						State:          commonpb.SegmentState_Flushed,
						MaxRowNum:      65536,
						LastExpireTime: 10,
					},
				},
				segID + 4: {
					SegmentInfo: &datapb.SegmentInfo{
						ID:             segID + 4,
						CollectionID:   collID,
						PartitionID:    partID,
						InsertChannel:  "",
						NumOfRows:      1026,
						State:          commonpb.SegmentState_Flushed,
						MaxRowNum:      65536,
						LastExpireTime: 10,
					},
				},
				segID + 5: {
					SegmentInfo: &datapb.SegmentInfo{
						ID:             segID + 5,
						CollectionID:   collID,
						PartitionID:    partID,
						InsertChannel:  "",
						NumOfRows:      1026,
						State:          commonpb.SegmentState_Flushed,
						MaxRowNum:      65536,
						LastExpireTime: 10,
					},
				},
				segID + 6: {
					SegmentInfo: &datapb.SegmentInfo{
						ID:             segID + 6,
						CollectionID:   collID,
						PartitionID:    partID,
						InsertChannel:  "",
						NumOfRows:      1026,
						State:          commonpb.SegmentState_Flushed,
						MaxRowNum:      65536,
						LastExpireTime: 10,
					},
				},
				segID + 7: {
					SegmentInfo: &datapb.SegmentInfo{
						ID:             segID + 7,
						CollectionID:   collID,
						PartitionID:    partID,
						InsertChannel:  "",
						NumOfRows:      1026,
						State:          commonpb.SegmentState_Flushed,
						MaxRowNum:      65536,
						LastExpireTime: 10,
					},
				},
				segID + 8: {
					SegmentInfo: &datapb.SegmentInfo{
						ID:             segID + 8,
						CollectionID:   collID,
						PartitionID:    partID,
						InsertChannel:  "",
						NumOfRows:      1026,
						State:          commonpb.SegmentState_Flushed,
						MaxRowNum:      65536,
						LastExpireTime: 10,
					},
				},
				segID + 9: {
					SegmentInfo: &datapb.SegmentInfo{
						ID:             segID + 9,
						CollectionID:   collID,
						PartitionID:    partID,
						InsertChannel:  "",
						NumOfRows:      500,
						State:          commonpb.SegmentState_Flushed,
						MaxRowNum:      65536,
						LastExpireTime: 10,
					},
				},
				segID + 10: {
					SegmentInfo: &datapb.SegmentInfo{
						ID:             segID + 10,
						CollectionID:   collID,
						PartitionID:    partID,
						InsertChannel:  "",
						NumOfRows:      500,
						State:          commonpb.SegmentState_Flushed,
						MaxRowNum:      65536,
						LastExpireTime: 10,
					},
				},
			},
		},
		analyzeMeta: am,
		indexMeta:   im,
	}
}

type taskSchedulerSuite struct {
	suite.Suite

	collectionID int64
	partitionID  int64
	fieldID      int64
	segmentIDs   []int64
	nodeID       int64
	duration     time.Duration
}

func (s *taskSchedulerSuite) initParams() {
	s.collectionID = 10000
	s.partitionID = 10001
	s.fieldID = 10002
	s.nodeID = 10003
	s.segmentIDs = []int64{1000, 1001, 1002}
	s.duration = time.Millisecond * 100
}

func (s *taskSchedulerSuite) createAnalyzeMeta(catalog metastore.DataCoordCatalog) *analyzeMeta {
	return &analyzeMeta{
		ctx:     context.Background(),
		catalog: catalog,
		tasks: map[int64]*model.AnalyzeTask{
			1: {
				TenantID:     "",
				CollectionID: s.collectionID,
				PartitionID:  s.partitionID,
				FieldID:      s.fieldID,
				SegmentIDs:   s.segmentIDs,
				TaskID:       1,
				State:        indexpb.JobState_JobStateInit,
			},
			2: {
				TenantID:     "",
				CollectionID: s.collectionID,
				PartitionID:  s.partitionID,
				FieldID:      s.fieldID,
				SegmentIDs:   s.segmentIDs,
				TaskID:       2,
				NodeID:       s.nodeID,
				State:        indexpb.JobState_JobStateInProgress,
			},
			3: {
				TenantID:     "",
				CollectionID: s.collectionID,
				PartitionID:  s.partitionID,
				FieldID:      s.fieldID,
				SegmentIDs:   s.segmentIDs,
				TaskID:       3,
				NodeID:       s.nodeID,
				State:        indexpb.JobState_JobStateFinished,
			},
			4: {
				TenantID:     "",
				CollectionID: s.collectionID,
				PartitionID:  s.partitionID,
				FieldID:      s.fieldID,
				SegmentIDs:   s.segmentIDs,
				TaskID:       4,
				NodeID:       s.nodeID,
				State:        indexpb.JobState_JobStateFailed,
			},
			5: {
				TenantID:     "",
				CollectionID: s.collectionID,
				PartitionID:  s.partitionID,
				FieldID:      s.fieldID,
				SegmentIDs:   []int64{1001, 1002},
				TaskID:       5,
				NodeID:       s.nodeID,
				State:        indexpb.JobState_JobStateRetry,
			},
		},
	}
}

func (s *taskSchedulerSuite) SetupTest() {
	paramtable.Init()
	s.initParams()
}

func (s *taskSchedulerSuite) scheduler(handler Handler) {
	ctx := context.Background()

	catalog := catalogmocks.NewDataCoordCatalog(s.T())
	catalog.EXPECT().SaveAnalyzeTask(mock.Anything, mock.Anything).Return(nil)
	catalog.EXPECT().AlterSegmentIndexes(mock.Anything, mock.Anything).Return(nil)

	in := mocks.NewMockIndexNodeClient(s.T())
	in.EXPECT().CreateJobV2(mock.Anything, mock.Anything).Return(merr.Success(), nil)
	in.EXPECT().QueryJobsV2(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, request *indexpb.QueryJobsV2Request, option ...grpc.CallOption) (*indexpb.QueryJobsV2Response, error) {
			switch request.GetJobType() {
			case indexpb.JobType_JobTypeIndexJob:
				results := make([]*indexpb.IndexTaskInfo, 0)
				for _, buildID := range request.GetTaskIDs() {
					results = append(results, &indexpb.IndexTaskInfo{
						BuildID:             buildID,
						State:               commonpb.IndexState_Finished,
						IndexFileKeys:       []string{"file1", "file2", "file3"},
						SerializedSize:      1024,
						FailReason:          "",
						CurrentIndexVersion: 1,
						IndexStoreVersion:   1,
					})
				}
				return &indexpb.QueryJobsV2Response{
					Status:    merr.Success(),
					ClusterID: request.GetClusterID(),
					Result: &indexpb.QueryJobsV2Response_IndexJobResults{
						IndexJobResults: &indexpb.IndexJobResults{
							Results: results,
						},
					},
				}, nil
			case indexpb.JobType_JobTypeAnalyzeJob:
				results := make([]*indexpb.AnalyzeResult, 0)
				for _, taskID := range request.GetTaskIDs() {
					results = append(results, &indexpb.AnalyzeResult{
						TaskID: taskID,
						State:  indexpb.JobState_JobStateFinished,
						//CentroidsFile: fmt.Sprintf("%d/stats_file", taskID),
						//SegmentOffsetMappingFiles: map[int64]string{
						//	1000: "1000/offset_mapping",
						//	1001: "1001/offset_mapping",
						//	1002: "1002/offset_mapping",
						//},
						FailReason: "",
					})
				}
				return &indexpb.QueryJobsV2Response{
					Status:    merr.Success(),
					ClusterID: request.GetClusterID(),
					Result: &indexpb.QueryJobsV2Response_AnalyzeJobResults{
						AnalyzeJobResults: &indexpb.AnalyzeResults{
							Results: results,
						},
					},
				}, nil
			default:
				return &indexpb.QueryJobsV2Response{
					Status:    merr.Status(errors.New("unknown job type")),
					ClusterID: request.GetClusterID(),
				}, nil
			}
		})
	in.EXPECT().DropJobsV2(mock.Anything, mock.Anything).Return(merr.Success(), nil)

	workerManager := NewMockWorkerManager(s.T())
	workerManager.EXPECT().PickClient().Return(s.nodeID, in)
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true)

	mt := createMeta(catalog, s.createAnalyzeMeta(catalog), createIndexMeta(catalog))

	cm := mocks.NewChunkManager(s.T())
	cm.EXPECT().RootPath().Return("root")

	scheduler := newTaskScheduler(ctx, mt, workerManager, cm, newIndexEngineVersionManager(), handler)
	s.Equal(9, len(scheduler.tasks))
	s.Equal(indexpb.JobState_JobStateInit, scheduler.tasks[1].GetState())
	s.Equal(indexpb.JobState_JobStateInProgress, scheduler.tasks[2].GetState())
	s.Equal(indexpb.JobState_JobStateRetry, scheduler.tasks[5].GetState())
	s.Equal(indexpb.JobState_JobStateInit, scheduler.tasks[buildID].GetState())
	s.Equal(indexpb.JobState_JobStateInProgress, scheduler.tasks[buildID+1].GetState())
	s.Equal(indexpb.JobState_JobStateInit, scheduler.tasks[buildID+3].GetState())
	s.Equal(indexpb.JobState_JobStateInProgress, scheduler.tasks[buildID+8].GetState())
	s.Equal(indexpb.JobState_JobStateInit, scheduler.tasks[buildID+9].GetState())
	s.Equal(indexpb.JobState_JobStateInit, scheduler.tasks[buildID+10].GetState())

	mt.segments.DropSegment(segID + 9)

	scheduler.scheduleDuration = time.Millisecond * 500
	scheduler.Start()

	s.Run("enqueue", func() {
		taskID := int64(6)
		newTask := &model.AnalyzeTask{
			CollectionID: s.collectionID,
			PartitionID:  s.partitionID,
			FieldID:      s.fieldID,
			SegmentIDs:   s.segmentIDs,
			TaskID:       taskID,
		}
		err := scheduler.meta.analyzeMeta.AddAnalyzeTask(newTask)
		s.NoError(err)
		t := &analyzeTask{
			taskID: taskID,
			taskInfo: &indexpb.AnalyzeResult{
				TaskID:     taskID,
				State:      indexpb.JobState_JobStateInit,
				FailReason: "",
			},
		}
		scheduler.enqueue(t)
	})

	for {
		scheduler.RLock()
		taskNum := len(scheduler.tasks)
		scheduler.RUnlock()

		if taskNum == 0 {
			break
		}
		time.Sleep(time.Second)
	}

	scheduler.Stop()
}

func (s *taskSchedulerSuite) Test_scheduler() {
	s.Run("test scheduler with indexBuilderV1", func() {
		s.scheduler(nil)
	})

	s.Run("test scheduler with indexBuilderV2", func() {
		paramtable.Get().CommonCfg.EnableStorageV2.SwapTempValue("true")
		defer paramtable.Get().CommonCfg.EnableStorageV2.SwapTempValue("false")

		handler := NewNMockHandler(s.T())
		handler.EXPECT().GetCollection(mock.Anything, mock.Anything).Return(&collectionInfo{
			ID: collID,
			Schema: &schemapb.CollectionSchema{
				Fields: []*schemapb.FieldSchema{
					{FieldID: fieldID, Name: "vec", TypeParams: []*commonpb.KeyValuePair{{Key: "dim", Value: "10"}}},
				},
			},
		}, nil)

		s.scheduler(handler)
	})
}

func (s *taskSchedulerSuite) Test_analyzeTaskFailCase() {
	ctx := context.Background()

	catalog := catalogmocks.NewDataCoordCatalog(s.T())
	catalog.EXPECT().DropAnalyzeTask(mock.Anything, mock.Anything).Return(nil)

	in := mocks.NewMockIndexNodeClient(s.T())

	workerManager := NewMockWorkerManager(s.T())

	mt := createMeta(catalog, s.createAnalyzeMeta(catalog), &indexMeta{
		RWMutex: sync.RWMutex{},
		ctx:     ctx,
		catalog: catalog,
	})

	scheduler := newTaskScheduler(ctx, mt, workerManager, nil, nil, nil)

	// remove task in meta
	err := scheduler.meta.analyzeMeta.DropAnalyzeTask(2)
	s.NoError(err)

	mt.segments.DropSegment(1000)
	scheduler.scheduleDuration = s.duration
	scheduler.Start()

	// taskID 1 peek client success, update version success. AssignTask failed --> state: None --> remove
	workerManager.EXPECT().PickClient().Return(s.nodeID, in).Once()
	catalog.EXPECT().SaveAnalyzeTask(mock.Anything, mock.Anything).Return(nil).Once()

	// taskID 5 state retry, drop task on worker --> state: Init
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().DropJobsV2(mock.Anything, mock.Anything).Return(merr.Success(), nil).Once()

	// pick client fail --> state: init
	workerManager.EXPECT().PickClient().Return(0, nil).Once()

	// update version failed --> state: init
	workerManager.EXPECT().PickClient().Return(s.nodeID, in)
	catalog.EXPECT().SaveAnalyzeTask(mock.Anything, mock.Anything).Return(errors.New("catalog update version error")).Once()

	// assign task to indexNode fail --> state: retry
	catalog.EXPECT().SaveAnalyzeTask(mock.Anything, mock.Anything).Return(nil).Once()
	in.EXPECT().CreateJobV2(mock.Anything, mock.Anything).Return(&commonpb.Status{
		Code:      65535,
		Retriable: false,
		Detail:    "",
		ExtraInfo: nil,
		Reason:    "mock error",
	}, nil).Once()

	// drop task failed --> state: retry
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().DropJobsV2(mock.Anything, mock.Anything).Return(merr.Status(errors.New("drop job failed")), nil).Once()

	// retry --> state: init
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().DropJobsV2(mock.Anything, mock.Anything).Return(merr.Success(), nil).Once()

	// update state to building failed --> state: retry
	catalog.EXPECT().SaveAnalyzeTask(mock.Anything, mock.Anything).Return(nil).Once()
	in.EXPECT().CreateJobV2(mock.Anything, mock.Anything).Return(merr.Success(), nil).Once()
	catalog.EXPECT().SaveAnalyzeTask(mock.Anything, mock.Anything).Return(errors.New("catalog update building state error")).Once()

	// retry --> state: init
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().DropJobsV2(mock.Anything, mock.Anything).Return(merr.Success(), nil).Once()

	// assign success --> state: InProgress
	catalog.EXPECT().SaveAnalyzeTask(mock.Anything, mock.Anything).Return(nil).Twice()
	in.EXPECT().CreateJobV2(mock.Anything, mock.Anything).Return(merr.Success(), nil).Once()

	// query result InProgress --> state: InProgress
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().QueryJobsV2(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, request *indexpb.QueryJobsV2Request, option ...grpc.CallOption) (*indexpb.QueryJobsV2Response, error) {
			results := make([]*indexpb.AnalyzeResult, 0)
			for _, taskID := range request.GetTaskIDs() {
				results = append(results, &indexpb.AnalyzeResult{
					TaskID: taskID,
					State:  indexpb.JobState_JobStateInProgress,
				})
			}
			return &indexpb.QueryJobsV2Response{
				Status:    merr.Success(),
				ClusterID: request.GetClusterID(),
				Result: &indexpb.QueryJobsV2Response_AnalyzeJobResults{
					AnalyzeJobResults: &indexpb.AnalyzeResults{
						Results: results,
					},
				},
			}, nil
		}).Once()

	// query result Retry --> state: retry
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().QueryJobsV2(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, request *indexpb.QueryJobsV2Request, option ...grpc.CallOption) (*indexpb.QueryJobsV2Response, error) {
			results := make([]*indexpb.AnalyzeResult, 0)
			for _, taskID := range request.GetTaskIDs() {
				results = append(results, &indexpb.AnalyzeResult{
					TaskID:     taskID,
					State:      indexpb.JobState_JobStateRetry,
					FailReason: "node analyze data failed",
				})
			}
			return &indexpb.QueryJobsV2Response{
				Status:    merr.Success(),
				ClusterID: request.GetClusterID(),
				Result: &indexpb.QueryJobsV2Response_AnalyzeJobResults{
					AnalyzeJobResults: &indexpb.AnalyzeResults{
						Results: results,
					},
				},
			}, nil
		}).Once()

	// retry --> state: init
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().DropJobsV2(mock.Anything, mock.Anything).Return(merr.Success(), nil).Once()

	// init --> state: InProgress
	catalog.EXPECT().SaveAnalyzeTask(mock.Anything, mock.Anything).Return(nil).Twice()
	in.EXPECT().CreateJobV2(mock.Anything, mock.Anything).Return(merr.Success(), nil).Once()

	// query result failed --> state: retry
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().QueryJobsV2(mock.Anything, mock.Anything).Return(&indexpb.QueryJobsV2Response{
		Status: merr.Status(errors.New("query job failed")),
	}, nil).Once()

	// retry --> state: init
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().DropJobsV2(mock.Anything, mock.Anything).Return(merr.Success(), nil).Once()

	// init --> state: InProgress
	catalog.EXPECT().SaveAnalyzeTask(mock.Anything, mock.Anything).Return(nil).Twice()
	in.EXPECT().CreateJobV2(mock.Anything, mock.Anything).Return(merr.Success(), nil).Once()

	// query result not exists --> state: retry
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().QueryJobsV2(mock.Anything, mock.Anything).Return(&indexpb.QueryJobsV2Response{
		Status:    merr.Success(),
		ClusterID: "",
		Result:    &indexpb.QueryJobsV2Response_AnalyzeJobResults{},
	}, nil).Once()

	// retry --> state: init
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().DropJobsV2(mock.Anything, mock.Anything).Return(merr.Success(), nil).Once()

	// init --> state: InProgress
	catalog.EXPECT().SaveAnalyzeTask(mock.Anything, mock.Anything).Return(nil).Twice()
	in.EXPECT().CreateJobV2(mock.Anything, mock.Anything).Return(merr.Success(), nil).Once()

	// node not exist --> state: retry
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(nil, false).Once()

	// retry --> state: init
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().DropJobsV2(mock.Anything, mock.Anything).Return(merr.Success(), nil).Once()

	// init --> state: InProgress
	catalog.EXPECT().SaveAnalyzeTask(mock.Anything, mock.Anything).Return(nil).Twice()
	in.EXPECT().CreateJobV2(mock.Anything, mock.Anything).Return(merr.Success(), nil).Once()

	// query result success --> state: finished
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().QueryJobsV2(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, request *indexpb.QueryJobsV2Request, option ...grpc.CallOption) (*indexpb.QueryJobsV2Response, error) {
			results := make([]*indexpb.AnalyzeResult, 0)
			for _, taskID := range request.GetTaskIDs() {
				results = append(results, &indexpb.AnalyzeResult{
					TaskID: taskID,
					State:  indexpb.JobState_JobStateFinished,
					//CentroidsFile: fmt.Sprintf("%d/stats_file", taskID),
					//SegmentOffsetMappingFiles: map[int64]string{
					//	1000: "1000/offset_mapping",
					//	1001: "1001/offset_mapping",
					//	1002: "1002/offset_mapping",
					//},
					FailReason: "",
				})
			}
			return &indexpb.QueryJobsV2Response{
				Status:    merr.Success(),
				ClusterID: request.GetClusterID(),
				Result: &indexpb.QueryJobsV2Response_AnalyzeJobResults{
					AnalyzeJobResults: &indexpb.AnalyzeResults{
						Results: results,
					},
				},
			}, nil
		}).Once()
	// set job info failed --> state: Finished
	catalog.EXPECT().SaveAnalyzeTask(mock.Anything, mock.Anything).Return(errors.New("set job info failed")).Once()

	// set job success, drop job on task failed --> state: Finished
	catalog.EXPECT().SaveAnalyzeTask(mock.Anything, mock.Anything).Return(nil).Once()
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().DropJobsV2(mock.Anything, mock.Anything).Return(merr.Status(errors.New("drop job failed")), nil).Once()

	// drop job success --> no task
	catalog.EXPECT().SaveAnalyzeTask(mock.Anything, mock.Anything).Return(nil).Once()
	workerManager.EXPECT().GetClientByID(mock.Anything).Return(in, true).Once()
	in.EXPECT().DropJobsV2(mock.Anything, mock.Anything).Return(merr.Success(), nil).Once()

	for {
		scheduler.RLock()
		taskNum := len(scheduler.tasks)
		scheduler.RUnlock()

		if taskNum == 0 {
			break
		}
		time.Sleep(time.Second)
	}

	scheduler.Stop()
}

func Test_taskSchedulerSuite(t *testing.T) {
	suite.Run(t, new(taskSchedulerSuite))
}