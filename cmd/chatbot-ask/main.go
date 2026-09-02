// Command chatbot-ask is a manual smoke test for the chatbot POC: it wires
// the Oracle SQL runner + Claude assistant directly (no HTTP, no JWT) and
// asks one question passed on the command line.
//
// macOS: export DYLD_LIBRARY_PATH=<instantclient dir> before running.
//
//	go run ./cmd/chatbot-ask "มี sample อะไรบ้างที่ค้าง pending เกิน 7 วัน"
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	anthropicchatbot "github.com/efangly/thanes-lims-backend/internal/adapters/anthropic/chatbot"
	oraclechatbot "github.com/efangly/thanes-lims-backend/internal/adapters/oracle/chatbot"
	oracledb "github.com/efangly/thanes-lims-backend/internal/adapters/oracle/db"
	applicationchatbot "github.com/efangly/thanes-lims-backend/internal/application/chatbot"
	"github.com/efangly/thanes-lims-backend/internal/config"
)

func main() {
	question := strings.Join(os.Args[1:], " ")
	if question == "" {
		question = "มี sample อะไรบ้างที่ยังค้างสถานะ pending เกิน 7 วัน"
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if !cfg.OracleEnabled || cfg.OracleDSN == "" {
		log.Fatal("set ORACLE_ENABLED=true and ORACLE_DSN in .env")
	}

	dsn := cfg.OracleChatbotDSN
	if dsn == "" {
		dsn = cfg.OracleDSN
		log.Println("note: ORACLE_CHATBOT_DSN unset, using ORACLE_DSN (CHATBOT_APP)")
	}

	sdb, err := oracledb.New(dsn, cfg.OracleTNSAdmin)
	if err != nil {
		log.Fatalf("oracle: %v", err)
	}
	defer sdb.Close()

	assistant := anthropicchatbot.New(cfg.AnthropicAPIKey, cfg.ChatbotModel, oraclechatbot.New(sdb))
	uc := applicationchatbot.NewAskUseCase(assistant)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	fmt.Printf("Q: %s\n\n", question)
	ans, err := uc.Execute(ctx, applicationchatbot.AskInput{Question: question})
	if err != nil {
		log.Fatalf("ask: %v", err)
	}

	for i, q := range ans.SQLQueries {
		fmt.Printf("SQL[%d]: %s\n", i+1, q)
	}
	fmt.Printf("\nrows=%d  elapsed=%dms  cache_read=%d  cache_write=%d\n\nA: %s\n",
		ans.Rows, ans.ElapsedMS, ans.CacheReadTokens, ans.CacheWriteTokens, ans.Text)
}
