package llm

// price is per-1M-token cost in USD micros (1 micro = $0.000001).
// Keep this table small and approximate; it drives cost visibility, not billing.
// Update as provider prices change. Unknown models cost 0 (still metered on tokens).
type price struct {
	inMicrosPerMTok  int64
	outMicrosPerMTok int64
}

var priceTable = map[string]price{
	"gemini-2.0-flash": {inMicrosPerMTok: 100, outMicrosPerMTok: 400},
	"gemini-1.5-flash": {inMicrosPerMTok: 75, outMicrosPerMTok: 300},
	"gpt-4o-mini":      {inMicrosPerMTok: 150, outMicrosPerMTok: 600},
	"claude-3-5-haiku": {inMicrosPerMTok: 800, outMicrosPerMTok: 4000},
}

// CostMicros returns the approximate USD-micro cost of a call. Unknown models
// return 0 so metering never blocks on an unpriced model.
func CostMicros(model string, promptTokens, completionTokens int) int64 {
	p, ok := priceTable[model]
	if !ok {
		return 0
	}
	in := (int64(promptTokens) * p.inMicrosPerMTok) / 1_000_000
	out := (int64(completionTokens) * p.outMicrosPerMTok) / 1_000_000
	return in + out
}
