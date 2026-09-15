// Package partnergrpc implements portenvironment.PartnerAPIClient against
// SMtrack's Partner API gRPC service (docs/partner-api-guide.md,
// proto/partner/partner.proto). Auth is an "x-api-key" gRPC metadata value
// sent per call (not a message field, not a static header) - see
// CONTEXT.md#environment, ADR 0011 and ADR 0012.
package partnergrpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/efangly/thanes-lims-backend/internal/adapters/partnergrpc/pb"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
)

type Client struct {
	conn    *grpc.ClientConn
	client  pb.PartnerServiceClient
	apiKey  string
	timeout time.Duration
}

// New dials the SMtrack Partner API gRPC service. addr is a bare host:port
// (e.g. "siamatic.thddns.net:50051") - an accidental "http://"/"https://"
// scheme prefix (an easy mistake, since operators are used to REST URLs) is
// stripped defensively rather than failing at boot.
//
// grpc.NewClient does not eagerly connect - a connectivity problem doesn't
// surface here, it surfaces as codes.Unavailable on the first RPC, which
// isRetryable already treats as a stale-cache-fallback case. So a non-nil
// error from New is limited to a malformed target/dial option, not "SMtrack
// is unreachable".
//
// TODO(security): SMtrack does not yet offer TLS for this endpoint, so this
// dials in plaintext (insecure.NewCredentials()) - a known, accepted gap
// (see ADR 0012), not a deliberate choice.
func New(addr, apiKey string, timeout time.Duration) (*Client, error) {
	addr = strings.TrimPrefix(strings.TrimPrefix(addr, "https://"), "http://")

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("partnergrpc: dial %s: %w", addr, err)
	}
	return &Client{
		conn:    conn,
		client:  pb.NewPartnerServiceClient(conn),
		apiKey:  apiKey,
		timeout: timeout,
	}, nil
}

// NewFromConn builds a Client around an already-established *grpc.ClientConn
// - used by tests to inject a bufconn-backed connection instead of dialing a
// real address.
func NewFromConn(conn *grpc.ClientConn, apiKey string, timeout time.Duration) *Client {
	return &Client{conn: conn, client: pb.NewPartnerServiceClient(conn), apiKey: apiKey, timeout: timeout}
}

// Close releases the underlying gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) withAuth(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-api-key", c.apiKey)
}

// RPCError is returned when the Partner API responds with a non-OK gRPC
// status. Code lets callers (the poller) distinguish a transient failure
// (rate limited, unavailable, or a deadline) worth a stale-cache fallback
// from a permanent one (unauthenticated/not found/invalid argument) that
// should surface as a real error - see ADR 0011.
type RPCError struct {
	Code    codes.Code
	Message string
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("partner api: %s: %s", e.Code, e.Message)
}

// Retryable reports whether the failure is worth a stale-cache fallback
// (rate limited, unavailable, or timed out) rather than a hard error.
func (e *RPCError) Retryable() bool {
	switch e.Code {
	case codes.ResourceExhausted, codes.Unavailable, codes.DeadlineExceeded:
		return true
	case codes.Unauthenticated, codes.NotFound, codes.InvalidArgument:
		return false
	default:
		return true
	}
}

// mapErr wraps a gRPC call error as *RPCError when it carries a gRPC status;
// a non-status error (e.g. a raw network error before the status machinery
// applies) is returned unwrapped, so the poller's isRetryable defaults it to
// retryable - matching the old REST client's behavior for bare network
// errors.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	return &RPCError{Code: st.Code(), Message: st.Message()}
}

func toMetadata(d *pb.DeviceMetadata) portenvironment.PartnerDeviceMetadata {
	return portenvironment.PartnerDeviceMetadata{
		Serial:   d.GetSerial(),
		Name:     d.GetName(),
		Status:   d.GetStatus(),
		Firmware: d.GetFirmware(),
		Online:   d.GetOnline(),
	}
}

// toReading maps a TelemetryPoint's send_time (ISO 8601 UTC) and the two
// display fields the port actually consumes. A send_time parse failure is a
// permanent adapter/protocol bug, not a network condition, so it is returned
// as an error rather than silently zeroing the time.
func toReading(serial string, p *pb.TelemetryPoint) (portenvironment.PartnerDeviceReading, error) {
	sendTime, err := time.Parse(time.RFC3339, p.GetSendTime())
	if err != nil {
		return portenvironment.PartnerDeviceReading{}, fmt.Errorf("partner api: parse send_time %q: %w", p.GetSendTime(), err)
	}
	return portenvironment.PartnerDeviceReading{
		Serial:          serial,
		SendTime:        sendTime,
		TempDisplay:     p.GetTempDisplay(),
		HumidityDisplay: p.GetHumidityDisplay(),
	}, nil
}

// FetchSnapshot returns the device's metadata plus its most recent reading
// in the trailing 1h window (timeseries[0], newest first) - found is false
// (with a nil error) when the device has no reading in that window.
func (c *Client) FetchSnapshot(ctx context.Context, serial string) (portenvironment.PartnerDeviceMetadata, portenvironment.PartnerDeviceReading, bool, error) {
	ctx, cancel := context.WithTimeout(c.withAuth(ctx), c.timeout)
	defer cancel()

	resp, err := c.client.GetDeviceSnapshot(ctx, &pb.GetDeviceSnapshotRequest{Serial: serial})
	if err != nil {
		return portenvironment.PartnerDeviceMetadata{}, portenvironment.PartnerDeviceReading{}, false, mapErr(err)
	}

	meta := toMetadata(resp.GetDevice())
	if len(resp.GetTimeseries()) == 0 {
		return meta, portenvironment.PartnerDeviceReading{}, false, nil
	}

	reading, err := toReading(serial, resp.GetTimeseries()[0])
	if err != nil {
		return portenvironment.PartnerDeviceMetadata{}, portenvironment.PartnerDeviceReading{}, false, err
	}
	return meta, reading, true, nil
}

// ListDevicesByWard returns every device SMtrack reports in ward (that the
// configured API key is scoped to), each with its single latest reading if
// it has one.
func (c *Client) ListDevicesByWard(ctx context.Context, ward string, page, limit int) (portenvironment.PartnerDeviceListing, error) {
	ctx, cancel := context.WithTimeout(c.withAuth(ctx), c.timeout)
	defer cancel()

	resp, err := c.client.ListDevicesByWard(ctx, &pb.ListDevicesByWardRequest{
		Ward:  ward,
		Page:  int32(page),
		Limit: int32(limit),
	})
	if err != nil {
		return portenvironment.PartnerDeviceListing{}, mapErr(err)
	}

	devices := make([]portenvironment.PartnerDeviceWithReading, 0, len(resp.GetDevices()))
	for _, d := range resp.GetDevices() {
		meta := toMetadata(d.GetDevice())
		item := portenvironment.PartnerDeviceWithReading{Device: meta}
		if latest := d.GetLatest(); latest != nil {
			reading, err := toReading(meta.Serial, latest)
			if err != nil {
				return portenvironment.PartnerDeviceListing{}, err
			}
			item.Latest = reading
			item.HasLatest = true
		}
		devices = append(devices, item)
	}

	return portenvironment.PartnerDeviceListing{
		Devices: devices,
		Total:   int(resp.GetTotal()),
		Page:    int(resp.GetPage()),
		Limit:   int(resp.GetLimit()),
	}, nil
}
