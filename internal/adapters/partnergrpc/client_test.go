package partnergrpc_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/efangly/thanes-lims-backend/internal/adapters/partnergrpc"
	"github.com/efangly/thanes-lims-backend/internal/adapters/partnergrpc/pb"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
)

// fakeServer is a configurable, test-local implementation of
// pb.PartnerServiceServer used to exercise the client without a real
// SMtrack connection.
type fakeServer struct {
	pb.UnimplementedPartnerServiceServer

	snapshotResp *pb.GetDeviceSnapshotResponse
	snapshotErr  error
	listResp     *pb.ListDevicesByWardResponse
	listErr      error

	gotAPIKey string
}

func (s *fakeServer) GetDeviceSnapshot(ctx context.Context, req *pb.GetDeviceSnapshotRequest) (*pb.GetDeviceSnapshotResponse, error) {
	s.captureAPIKey(ctx)
	if s.snapshotErr != nil {
		return nil, s.snapshotErr
	}
	return s.snapshotResp, nil
}

func (s *fakeServer) ListDevicesByWard(ctx context.Context, req *pb.ListDevicesByWardRequest) (*pb.ListDevicesByWardResponse, error) {
	s.captureAPIKey(ctx)
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.listResp, nil
}

func (s *fakeServer) captureAPIKey(ctx context.Context) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-api-key"); len(vals) > 0 {
			s.gotAPIKey = vals[0]
		}
	}
}

// newTestClient starts an in-process gRPC server (bufconn - avoids
// port-binding flakiness) backed by srv, and returns a *partnergrpc.Client
// dialed against it.
func newTestClient(t *testing.T, srv *fakeServer) *partnergrpc.Client {
	t.Helper()

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	pb.RegisterPartnerServiceServer(s, srv)
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(s.Stop)

	// partnergrpc.New only accepts a real dial target, so build the client
	// directly against a bufconn-backed *grpc.ClientConn the same way New
	// does, minus the addr/credentials plumbing.
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return partnergrpc.NewFromConn(conn, "test-key", 5*time.Second)
}

func TestFetchSnapshot_HappyPath(t *testing.T) {
	srv := &fakeServer{snapshotResp: &pb.GetDeviceSnapshotResponse{
		Device: &pb.DeviceMetadata{Serial: "SN-00042", Name: "Fridge", Status: true, Firmware: "1.2.3", Online: true},
		Timeseries: []*pb.TelemetryPoint{
			{SendTime: "2026-09-15T10:00:00Z", TempDisplay: 4.8, HumidityDisplay: 55.5},
			{SendTime: "2026-09-15T09:00:00Z", TempDisplay: 4.5, HumidityDisplay: 54.0},
		},
	}}
	client := newTestClient(t, srv)

	meta, reading, found, err := client.FetchSnapshot(context.Background(), "SN-00042")

	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, portenvironment.PartnerDeviceMetadata{Serial: "SN-00042", Name: "Fridge", Status: true, Firmware: "1.2.3", Online: true}, meta)
	assert.Equal(t, 4.8, reading.TempDisplay)
	assert.Equal(t, 55.5, reading.HumidityDisplay)
	assert.Equal(t, "SN-00042", reading.Serial)
	assert.Equal(t, "test-key", srv.gotAPIKey)
}

func TestFetchTimeseries_ReturnsEveryPoint(t *testing.T) {
	srv := &fakeServer{snapshotResp: &pb.GetDeviceSnapshotResponse{
		Device: &pb.DeviceMetadata{Serial: "SN-00042", Name: "Fridge"},
		Timeseries: []*pb.TelemetryPoint{
			{SendTime: "2026-09-15T10:00:00Z", TempDisplay: 4.8, HumidityDisplay: 55.5},
			{SendTime: "2026-09-15T09:55:00Z", TempDisplay: 4.7, HumidityDisplay: 55.2},
			{SendTime: "2026-09-15T09:50:00Z", TempDisplay: 4.6, HumidityDisplay: 55.0},
		},
	}}
	client := newTestClient(t, srv)

	readings, err := client.FetchTimeseries(context.Background(), "SN-00042")

	require.NoError(t, err)
	require.Len(t, readings, 3)
	assert.Equal(t, 4.8, readings[0].TempDisplay)
	assert.Equal(t, 4.6, readings[2].TempDisplay)
	for _, r := range readings {
		assert.Equal(t, "SN-00042", r.Serial)
	}
}

func TestFetchTimeseries_EmptyWhenNoReadings(t *testing.T) {
	srv := &fakeServer{snapshotResp: &pb.GetDeviceSnapshotResponse{
		Device:     &pb.DeviceMetadata{Serial: "SN-00042"},
		Timeseries: nil,
	}}
	client := newTestClient(t, srv)

	readings, err := client.FetchTimeseries(context.Background(), "SN-00042")

	require.NoError(t, err)
	assert.Empty(t, readings)
}

func TestFetchTimeseries_ErrorMapped(t *testing.T) {
	srv := &fakeServer{snapshotErr: status.Error(codes.NotFound, "unknown serial")}
	client := newTestClient(t, srv)

	_, err := client.FetchTimeseries(context.Background(), "SN-UNKNOWN")

	require.Error(t, err)
	var re portenvironment.RetryableError
	require.True(t, errors.As(err, &re))
	assert.False(t, re.Retryable())
}

func TestFetchSnapshot_EmptyTimeseriesNotFound(t *testing.T) {
	srv := &fakeServer{snapshotResp: &pb.GetDeviceSnapshotResponse{
		Device:     &pb.DeviceMetadata{Serial: "SN-00042"},
		Timeseries: nil,
	}}
	client := newTestClient(t, srv)

	_, _, found, err := client.FetchSnapshot(context.Background(), "SN-00042")

	require.NoError(t, err)
	assert.False(t, found)
}

func TestFetchSnapshot_StatusCodeRetryability(t *testing.T) {
	cases := []struct {
		code      codes.Code
		retryable bool
	}{
		{codes.Unauthenticated, false},
		{codes.NotFound, false},
		{codes.InvalidArgument, false},
		{codes.ResourceExhausted, true},
		{codes.Unavailable, true},
		{codes.DeadlineExceeded, true},
	}
	for _, tc := range cases {
		t.Run(tc.code.String(), func(t *testing.T) {
			srv := &fakeServer{snapshotErr: status.Error(tc.code, "boom")}
			client := newTestClient(t, srv)

			_, _, _, err := client.FetchSnapshot(context.Background(), "SN-00042")

			require.Error(t, err)
			var re portenvironment.RetryableError
			require.True(t, errors.As(err, &re))
			assert.Equal(t, tc.retryable, re.Retryable())
		})
	}
}

func TestListDevicesByWard_HappyPath(t *testing.T) {
	srv := &fakeServer{listResp: &pb.ListDevicesByWardResponse{
		Devices: []*pb.DeviceWithLatestReading{
			{
				Device: &pb.DeviceMetadata{Serial: "SN-00042", Name: "Fridge"},
				Latest: &pb.TelemetryPoint{SendTime: "2026-09-15T10:00:00Z", TempDisplay: 4.8, HumidityDisplay: 55.5},
			},
			{
				Device: &pb.DeviceMetadata{Serial: "SN-00099", Name: "Freezer"},
				Latest: nil,
			},
		},
		Total: 2, Page: 1, Limit: 20,
	}}
	client := newTestClient(t, srv)

	listing, err := client.ListDevicesByWard(context.Background(), "ICU", 1, 20)

	require.NoError(t, err)
	assert.Equal(t, 2, listing.Total)
	require.Len(t, listing.Devices, 2)
	assert.True(t, listing.Devices[0].HasLatest)
	assert.Equal(t, 4.8, listing.Devices[0].Latest.TempDisplay)
	assert.False(t, listing.Devices[1].HasLatest)
}

func TestListDevicesByWard_NotFound(t *testing.T) {
	srv := &fakeServer{listErr: status.Error(codes.NotFound, "unknown ward")}
	client := newTestClient(t, srv)

	_, err := client.ListDevicesByWard(context.Background(), "UNKNOWN", 1, 20)

	require.Error(t, err)
	var re portenvironment.RetryableError
	require.True(t, errors.As(err, &re))
	assert.False(t, re.Retryable())
}
