package chainkitdepositeventinbox

import "testing"

func TestStatusValuesMatchSchema(t *testing.T) {
	tests := map[string]struct {
		got  int8
		want int8
	}{
		"orphaned":   {StatusOrphaned, -1},
		"ignored":    {StatusIgnored, -2},
		"pending":    {StatusPending, 0},
		"ready":      {StatusReady, 1},
		"processing": {StatusProcessing, 2},
		"completed":  {StatusCompleted, 3},
		"retry":      {StatusRetry, 4},
	}
	for name, test := range tests {
		if test.got != test.want {
			t.Errorf("%s status = %d, want %d", name, test.got, test.want)
		}
	}
}

func TestSourceValuesAreBitMask(t *testing.T) {
	if SourceWebSocket != 1 || SourceGetLogs != 2 || SourceWebSocket|SourceGetLogs != 3 {
		t.Fatalf("unexpected source values: websocket=%d getLogs=%d", SourceWebSocket, SourceGetLogs)
	}
}

func TestNormalize(t *testing.T) {
	event := &ChainDepositEventInbox{
		ContractAddress: "0xAaBb",
		TxHash:          "0xCcDd",
		BlockHash:       "0xEeFf",
		FromAddress:     "0xAAbb",
		ToAddress:       "0xCCdd",
	}
	event.normalize()

	if event.ContractAddress != "0xaabb" || event.TxHash != "0xccdd" ||
		event.BlockHash != "0xeeff" || event.FromAddress != "0xaabb" ||
		event.ToAddress != "0xccdd" {
		t.Fatalf("event was not normalized: %#v", event)
	}
}

func TestTruncatePreservesUTF8(t *testing.T) {
	if got := truncate("充值事件失败", 4); got != "充值事件" {
		t.Fatalf("truncate() = %q, want %q", got, "充值事件")
	}
}
