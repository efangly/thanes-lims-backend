package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portssample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	portsuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SampleTools implements the Sample-domain MCP tools: getSampleById,
// listSamplesByStatus, listSamplesByCustodianName. Users is needed to
// resolve a Sample's CustodianUserID to a display name, and (for
// listSamplesByCustodianName) to resolve a name back to a user ID.
type SampleTools struct {
	Samples portssample.SampleRepository
	Users   portsuser.UserRepository
}

// SampleOutput is the shape returned by every Sample tool.
type SampleOutput struct {
	ID   string `json:"id" jsonschema:"Sample ID, e.g. SMP-2569-00001."`
	Name string `json:"name" jsonschema:"Human-readable sample name/label."`
	Type string `json:"type" jsonschema:"Sample material type: one of blood, urine, water, tissue, food, serum."`
	// Status enum meaning is spelled out for the LLM since it never sees the
	// Go domain type, only this JSON schema description.
	Status          string  `json:"status" jsonschema:"Lifecycle status: pending (received, awaiting testing), testing (being tested now), completed (testing finished - terminal, never reopened), transferred (moved to another department)."`
	CustodianUserID int64   `json:"custodianUserId" jsonschema:"User ID of the person responsible for this sample."`
	CustodianName   string  `json:"custodianName,omitempty" jsonschema:"Full name of the custodian, when resolvable from the user directory."`
	LocationID      *string `json:"locationId,omitempty" jsonschema:"Storage location ID the sample currently occupies, if it has been put away."`
	ReceivedAt      string  `json:"receivedAt" jsonschema:"RFC3339 timestamp the sample was received into the lab."`
	PendingDays     *int    `json:"pendingDays,omitempty" jsonschema:"Whole days elapsed since receivedAt. Only populated when the query filtered by an age threshold (olderThanDays)."`
	BarcodeID       *string `json:"barcodeId,omitempty" jsonschema:"Optional physical/scan barcode identifier, distinct from id."`
	Description     string  `json:"description,omitempty" jsonschema:"Free-text notes captured when the sample was received."`
}

func (t *SampleTools) resolveCustodianName(ctx context.Context, userID int64) string {
	u, err := t.Users.FindByID(ctx, userID)
	if err != nil {
		return ""
	}
	return u.Name
}

func (t *SampleTools) toOutput(ctx context.Context, s sample.Sample) SampleOutput {
	return SampleOutput{
		ID:              s.ID,
		Name:            s.Name,
		Type:            string(s.Type),
		Status:          string(s.Status),
		CustodianUserID: s.CustodianUserID,
		CustodianName:   t.resolveCustodianName(ctx, s.CustodianUserID),
		LocationID:      s.LocationID,
		ReceivedAt:      s.ReceivedAt.Format(time.RFC3339),
		BarcodeID:       s.BarcodeID,
		Description:     s.Description,
	}
}

// --- getSampleById ---

type GetSampleByIDInput struct {
	ID string `json:"id" jsonschema:"Sample ID to look up, e.g. SMP-2569-00001."`
}

// GetByID implements the getSampleById tool.
func (t *SampleTools) GetByID(ctx context.Context, _ *mcp.CallToolRequest, in GetSampleByIDInput) (*mcp.CallToolResult, SampleOutput, error) {
	if strings.TrimSpace(in.ID) == "" {
		return errorResult[SampleOutput]("id is required")
	}
	s, err := t.Samples.FindByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return errorResult[SampleOutput](fmt.Sprintf("sample %q not found", in.ID))
		}
		return nil, SampleOutput{}, err
	}
	return nil, t.toOutput(ctx, s), nil
}

// --- listSamplesByStatus ---

type ListSamplesByStatusInput struct {
	Status string `json:"status" jsonschema:"Sample status to filter by: pending, testing, completed, or transferred."`
	// OlderThanDays supports "pending samples stuck longer than N days"
	// (docs/chatbot-acceptance-checklist.md scenario 1). The repository's
	// List has no date filter, so this is applied in the handler against
	// ReceivedAt after a broader status-filtered List call, to avoid
	// widening the SampleRepository interface for a single narrow query.
	OlderThanDays *int `json:"olderThanDays,omitempty" jsonschema:"When set, only return samples whose receivedAt is more than this many whole days in the past - e.g. 7 for 'pending longer than a week'. Typically combined with status=pending."`
}

type ListSamplesOutput struct {
	Samples []SampleOutput `json:"samples" jsonschema:"Matching samples, most fields as in getSampleById."`
	Count   int            `json:"count" jsonschema:"Number of matching samples."`
}

// ListByStatus implements the listSamplesByStatus tool.
func (t *SampleTools) ListByStatus(ctx context.Context, _ *mcp.CallToolRequest, in ListSamplesByStatusInput) (*mcp.CallToolResult, ListSamplesOutput, error) {
	status := sample.Status(in.Status)
	switch status {
	case sample.StatusPending, sample.StatusTesting, sample.StatusCompleted, sample.StatusTransferred:
	default:
		return errorResult[ListSamplesOutput](fmt.Sprintf("invalid status %q: must be one of pending, testing, completed, transferred", in.Status))
	}

	samples, err := t.Samples.List(ctx, portssample.ListFilter{Status: &status})
	if err != nil {
		return nil, ListSamplesOutput{}, err
	}

	now := time.Now()
	out := make([]SampleOutput, 0, len(samples))
	for _, s := range samples {
		if in.OlderThanDays != nil {
			days := int(now.Sub(s.ReceivedAt).Hours() / 24)
			if days < *in.OlderThanDays {
				continue
			}
			so := t.toOutput(ctx, s)
			so.PendingDays = &days
			out = append(out, so)
			continue
		}
		out = append(out, t.toOutput(ctx, s))
	}
	return nil, ListSamplesOutput{Samples: out, Count: len(out)}, nil
}

// --- listSamplesByCustodianName ---

type ListSamplesByCustodianNameInput struct {
	Name string `json:"name" jsonschema:"Custodian's full name to search for, e.g. \"วิภา สายใจ\" (case-insensitive exact match against the user's display name)."`
}

// ListByCustodianName implements the listSamplesByCustodianName tool.
//
// internal/ports/user.UserRepository has no find-by-name method, and adding
// one purely for this single chatbot query would widen a core port for a
// low-traffic, read-only convenience lookup. Instead this resolves the name
// by listing all Users and matching case-insensitively in the handler -
// acceptable because the User table is small (lab staff, not customers) and
// this tool is not on any hot path. See docs/mcp-server-tools.md for the
// documented tradeoff.
func (t *SampleTools) ListByCustodianName(ctx context.Context, _ *mcp.CallToolRequest, in ListSamplesByCustodianNameInput) (*mcp.CallToolResult, ListSamplesOutput, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return errorResult[ListSamplesOutput]("name is required")
	}

	users, err := t.Users.List(ctx)
	if err != nil {
		return nil, ListSamplesOutput{}, err
	}
	var userID int64
	found := false
	for _, u := range users {
		if strings.EqualFold(strings.TrimSpace(u.Name), name) {
			userID = u.ID
			found = true
			break
		}
	}
	if !found {
		return errorResult[ListSamplesOutput](fmt.Sprintf("no user found with name %q", in.Name))
	}

	samples, err := t.Samples.List(ctx, portssample.ListFilter{CustodianUserID: &userID})
	if err != nil {
		return nil, ListSamplesOutput{}, err
	}
	out := make([]SampleOutput, 0, len(samples))
	for _, s := range samples {
		out = append(out, t.toOutput(ctx, s))
	}
	return nil, ListSamplesOutput{Samples: out, Count: len(out)}, nil
}
