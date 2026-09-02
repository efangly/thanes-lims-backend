// Package chatbot is the Claude API adapter for the chatbot POC. It runs a
// manual tool-use loop: Claude is given the ADB schema and a single run_sql
// tool, proposes SELECT queries, the adapter executes them read-only via the
// SQLRunner port, and Claude narrates a final Thai answer from the rows.
//
// This replaces the original Oracle Select AI approach (see
// docs/chatbot-poc-plan.md, 2026-09-02 pivot): NL->SQL now happens in Go
// against the Claude API instead of inside the Oracle DB.
package chatbot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	domainchatbot "github.com/efangly/thanes-lims-backend/internal/domain/chatbot"
	portchatbot "github.com/efangly/thanes-lims-backend/internal/ports/chatbot"
)

const (
	maxTurns  = 5
	maxTokens = 4096
)

// systemPrompt gives Claude the ADB schema (mirrors scripts/oracle/001_schema.sql)
// and the answering rules. Kept in sync with that file by hand - the 4 tables
// are stable POC scope.
const systemPrompt = `คุณเป็นผู้ช่วยตอบคำถามข้อมูลห้องปฏิบัติการ (LIMS) จากฐานข้อมูล Oracle

ใช้ tool "run_sql" เพื่อ query ข้อมูล โดยเขียนได้เฉพาะ SELECT statement เดียว (Oracle SQL dialect, ห้ามมี ; ปิดท้าย, ห้าม DML/DDL)
วันปัจจุบันใน DB คือ SYSDATE เสมอ เช่นคำถาม "ค้างเกิน 7 วัน" ให้ใช้เงื่อนไขแบบ received_at < SYSDATE - 7
(received_at เป็น TIMESTAMP อย่าใช้ TRUNC/CAST กับผลลบ interval — ถ้าต้องแสดงจำนวนวันให้ใช้ FLOOR(SYSDATE - CAST(received_at AS DATE)))
เรียก run_sql ซ้ำได้ถ้า query ผิดพลาดหรือต้องดึงข้อมูลเพิ่ม (ไม่เกิน 5 ครั้ง)
เมื่อได้ข้อมูลครบแล้ว ตอบเป็นภาษาไทย กระชับ อ้างอิงรหัส/ตัวเลขจริงจากผลลัพธ์ ถ้าข้อมูลไม่พอให้บอกตรงๆ อย่าเดา

Schema (ตารางทั้งหมดอยู่ใน schema ปัจจุบัน เรียกชื่อตารางได้ตรงๆ):

samples — ตัวอย่างที่ห้องแล็บรับเข้ามาตรวจวิเคราะห์
  id VARCHAR2 (เช่น SMP-2569-00001), name, sample_type (blood|urine|water|tissue|food|serum),
  custodian (ชื่อผู้ดูแล), location, status (pending=รอดำเนินการ|testing=กำลังทดสอบ|completed=เสร็จ|transferred=ส่งต่อ),
  received_at TIMESTAMP (วันเวลาที่รับเข้าระบบ)

test_results — ผลการทดสอบของแต่ละตัวอย่าง (1 ตัวอย่างมีได้หลายผล)
  id VARCHAR2, sample_id -> samples.id, test_name, analyst, result,
  flag (hi=สูงกว่าปกติ|lo=ต่ำกว่าปกติ|ok=ปกติ), ref_range,
  status (analyzing|pending_verification|approved)

inventory_items — วัสดุ/สารเคมี/อุปกรณ์คงคลัง
  id VARCHAR2 (เช่น INV-0001), name, category, quantity (คงเหลือ), unit,
  min_qty (ต่ำกว่านี้ถือว่าสต็อกต่ำ ต้องสั่งซื้อ), max_qty, default_vendor

purchase_orders — ใบสั่งซื้อเติมสต็อก
  id VARCHAR2 (เช่น PO-2569-0001), item_id -> inventory_items.id, quantity, vendor,
  order_date DATE, status (pending_approval=รออนุมัติ|sent_to_vendor=ส่งผู้ขายแล้ว|received=รับของแล้ว|cancelled=ยกเลิก)`

// Assistant implements ports/chatbot.Assistant via the Claude Messages API.
type Assistant struct {
	client anthropic.Client
	model  string
	runner portchatbot.SQLRunner
}

// New builds the adapter. If apiKey is empty the SDK resolves credentials the
// usual way (ANTHROPIC_API_KEY / `ant auth login` profile).
func New(apiKey, model string, runner portchatbot.SQLRunner) *Assistant {
	var opts []option.RequestOption
	if apiKey != "" {
		opts = append(opts, option.WithAPIKey(apiKey))
	}
	return &Assistant{
		client: anthropic.NewClient(opts...),
		model:  model,
		runner: runner,
	}
}

func (a *Assistant) Ask(ctx context.Context, q domainchatbot.Question) (domainchatbot.Answer, error) {
	start := time.Now()

	tool := anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{
		Properties: map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Oracle SQL SELECT statement เดียว ไม่มี ; ปิดท้าย",
			},
		},
		Required: []string{"query"},
	}, "run_sql")
	tool.OfTool.Description = anthropic.String("รัน SELECT query กับฐานข้อมูล Oracle แล้วคืนผลลัพธ์เป็นตาราง (อ่านอย่างเดียว)")

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(a.model),
		MaxTokens: maxTokens,
		// Prompt caching: the system prompt (schema + rules) and the tool
		// definition are byte-identical on every ask, so a breakpoint on the
		// last system block caches tools+system together (render order is
		// tools -> system -> messages). 5-minute ephemeral TTL - within one
		// ask the tool-use turns are seconds apart, and back-to-back asks
		// keep the entry warm. Note: Haiku 4.5's minimum cacheable prefix is
		// 4096 tokens; below that this is silently a no-op (usage fields
		// on the Answer show whether it took effect).
		System: []anthropic.TextBlockParam{{
			Text:         systemPrompt,
			CacheControl: anthropic.NewCacheControlEphemeralParam(),
		}},
		Tools: []anthropic.ToolUnionParam{tool},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(q.Text)),
		},
	}

	ans := domainchatbot.Answer{}

	for turn := 0; turn < maxTurns; turn++ {
		resp, err := a.client.Messages.New(ctx, params)
		if err != nil {
			return domainchatbot.Answer{}, fmt.Errorf("chatbot: llm call: %w", err)
		}
		ans.CacheReadTokens += resp.Usage.CacheReadInputTokens
		ans.CacheWriteTokens += resp.Usage.CacheCreationInputTokens

		var textParts []string
		var toolResults []anthropic.ContentBlockParamUnion

		for _, block := range resp.Content {
			switch v := block.AsAny().(type) {
			case anthropic.TextBlock:
				if s := strings.TrimSpace(v.Text); s != "" {
					textParts = append(textParts, s)
				}
			case anthropic.ToolUseBlock:
				if v.Name != "run_sql" {
					toolResults = append(toolResults, anthropic.NewToolResultBlock(v.ID, "unknown tool", true))
					continue
				}
				var in struct {
					Query string `json:"query"`
				}
				if err := json.Unmarshal(v.Input, &in); err != nil {
					toolResults = append(toolResults, anthropic.NewToolResultBlock(v.ID, "invalid tool input: "+err.Error(), true))
					continue
				}
				ans.SQLQueries = append(ans.SQLQueries, in.Query)
				cols, rows, runErr := a.runner.RunSelect(ctx, in.Query)
				if runErr != nil {
					toolResults = append(toolResults, anthropic.NewToolResultBlock(v.ID, "SQL error: "+runErr.Error(), true))
					continue
				}
				ans.Rows += len(rows)
				toolResults = append(toolResults, anthropic.NewToolResultBlock(v.ID, formatRows(cols, rows), false))
			}
		}

		if resp.StopReason != anthropic.StopReasonToolUse {
			ans.Text = strings.Join(textParts, "\n\n")
			ans.ElapsedMS = time.Since(start).Milliseconds()
			return ans, nil
		}

		params.Messages = append(params.Messages, resp.ToParam())
		params.Messages = append(params.Messages, anthropic.NewUserMessage(toolResults...))
	}

	return domainchatbot.Answer{}, fmt.Errorf("chatbot: exceeded %d tool-use turns without a final answer", maxTurns)
}

// formatRows renders a result set as a compact pipe table for the LLM.
func formatRows(cols []string, rows [][]string) string {
	if len(rows) == 0 {
		return "(0 rows)"
	}
	var b strings.Builder
	b.WriteString(strings.Join(cols, " | "))
	b.WriteString("\n")
	for _, r := range rows {
		b.WriteString(strings.Join(r, " | "))
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "(%d rows)", len(rows))
	return b.String()
}
