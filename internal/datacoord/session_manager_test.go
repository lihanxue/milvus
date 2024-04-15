package datacoord

import (
	"context"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/milvus-io/milvus/internal/mocks"
	"github.com/milvus-io/milvus/internal/proto/datapb"
	"github.com/milvus-io/milvus/internal/types"
	"github.com/milvus-io/milvus/pkg/metrics"
	"github.com/milvus-io/milvus/pkg/util/merr"
	"github.com/milvus-io/milvus/pkg/util/testutils"
)

func TestSessionManagerSuite(t *testing.T) {
	suite.Run(t, new(SessionManagerSuite))
}

type SessionManagerSuite struct {
	testutils.PromMetricsSuite

	dn *mocks.MockDataNodeClient

	m *SessionManagerImpl
}

func (s *SessionManagerSuite) SetupTest() {
	s.dn = mocks.NewMockDataNodeClient(s.T())

	s.m = NewSessionManagerImpl(withSessionCreator(func(ctx context.Context, addr string, nodeID int64) (types.DataNodeClient, error) {
		return s.dn, nil
	}))

	s.m.AddSession(&NodeInfo{1000, "addr-1"})
	s.MetricsEqual(metrics.DataCoordNumDataNodes, 1)
}

func (s *SessionManagerSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *SessionManagerSuite) TestExecFlush() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := &datapb.FlushSegmentsRequest{
		CollectionID: 1,
		SegmentIDs:   []int64{100, 200},
		ChannelName:  "ch-1",
	}

	s.Run("no node", func() {
		s.m.execFlush(ctx, 100, req)
	})

	s.Run("fail", func() {
		s.dn.EXPECT().FlushSegments(mock.Anything, mock.Anything).Return(nil, errors.New("mock")).Once()
		s.m.execFlush(ctx, 1000, req)
	})

	s.Run("normal", func() {
		s.dn.EXPECT().FlushSegments(mock.Anything, mock.Anything).Return(merr.Status(nil), nil).Once()
		s.m.execFlush(ctx, 1000, req)
	})
}

func (s *SessionManagerSuite) TestNotifyChannelOperation() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	info := &datapb.ChannelWatchInfo{
		Vchan: &datapb.VchannelInfo{},
		State: datapb.ChannelWatchState_ToWatch,
		OpID:  1,
	}

	req := &datapb.ChannelOperationsRequest{
		Infos: []*datapb.ChannelWatchInfo{info},
	}
	s.Run("no node", func() {
		err := s.m.NotifyChannelOperation(ctx, 100, req)
		s.Error(err)
	})

	s.Run("fail", func() {
		s.dn.EXPECT().NotifyChannelOperation(mock.Anything, mock.Anything).Return(nil, errors.New("mock")).Once()

		err := s.m.NotifyChannelOperation(ctx, 1000, req)
		s.Error(err)
	})

	s.Run("normal", func() {
		s.dn.EXPECT().NotifyChannelOperation(mock.Anything, mock.Anything).Return(merr.Status(nil), nil).Once()

		err := s.m.NotifyChannelOperation(ctx, 1000, req)
		s.NoError(err)
	})
}

func (s *SessionManagerSuite) TestCheckCHannelOperationProgress() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	info := &datapb.ChannelWatchInfo{
		Vchan: &datapb.VchannelInfo{},
		State: datapb.ChannelWatchState_ToWatch,
		OpID:  1,
	}

	s.Run("no node", func() {
		resp, err := s.m.CheckChannelOperationProgress(ctx, 100, info)
		s.Error(err)
		s.Nil(resp)
	})

	s.Run("fail", func() {
		s.dn.EXPECT().CheckChannelOperationProgress(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("mock")).Once()

		resp, err := s.m.CheckChannelOperationProgress(ctx, 1000, info)
		s.Error(err)
		s.Nil(resp)
	})

	s.Run("normal", func() {
		s.dn.EXPECT().CheckChannelOperationProgress(mock.Anything, mock.Anything, mock.Anything).
			Return(&datapb.ChannelOperationProgressResponse{
				Status:   merr.Status(nil),
				OpID:     info.OpID,
				State:    info.State,
				Progress: 100,
			}, nil).Once()

		resp, err := s.m.CheckChannelOperationProgress(ctx, 1000, info)
		s.NoError(err)
		s.Equal(resp.GetState(), info.State)
		s.Equal(resp.OpID, info.OpID)
		s.EqualValues(100, resp.Progress)
	})
}

func (s *SessionManagerSuite) TestImportV2() {
	mockErr := errors.New("mock error")

	s.Run("PreImport", func() {
		err := s.m.PreImport(0, &datapb.PreImportRequest{})
		s.Error(err)

		s.SetupTest()
		s.dn.EXPECT().PreImport(mock.Anything, mock.Anything).Return(merr.Success(), nil)
		err = s.m.PreImport(1000, &datapb.PreImportRequest{})
		s.NoError(err)
	})

	s.Run("ImportV2", func() {
		err := s.m.ImportV2(0, &datapb.ImportRequest{})
		s.Error(err)

		s.SetupTest()
		s.dn.EXPECT().ImportV2(mock.Anything, mock.Anything).Return(merr.Success(), nil)
		err = s.m.ImportV2(1000, &datapb.ImportRequest{})
		s.NoError(err)
	})

	s.Run("QueryPreImport", func() {
		_, err := s.m.QueryPreImport(0, &datapb.QueryPreImportRequest{})
		s.Error(err)

		s.SetupTest()
		s.dn.EXPECT().QueryPreImport(mock.Anything, mock.Anything).Return(&datapb.QueryPreImportResponse{
			Status: merr.Status(mockErr),
		}, nil)
		_, err = s.m.QueryPreImport(1000, &datapb.QueryPreImportRequest{})
		s.Error(err)
	})

	s.Run("QueryImport", func() {
		_, err := s.m.QueryImport(0, &datapb.QueryImportRequest{})
		s.Error(err)

		s.SetupTest()
		s.dn.EXPECT().QueryImport(mock.Anything, mock.Anything).Return(&datapb.QueryImportResponse{
			Status: merr.Status(mockErr),
		}, nil)
		_, err = s.m.QueryImport(1000, &datapb.QueryImportRequest{})
		s.Error(err)
	})

	s.Run("DropImport", func() {
		err := s.m.DropImport(0, &datapb.DropImportRequest{})
		s.Error(err)

		s.SetupTest()
		s.dn.EXPECT().DropImport(mock.Anything, mock.Anything).Return(merr.Success(), nil)
		err = s.m.DropImport(1000, &datapb.DropImportRequest{})
		s.NoError(err)
	})
}
