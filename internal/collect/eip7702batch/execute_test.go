package eip7702batch

import (
	"math"
	"math/big"
	"testing"

	"github.com/jiajia556/chainkit/pkg/contracts/batchsweepexecutor"
)

func TestAddCollectExecutionBudget(t *testing.T) {
	items := []batchsweepexecutor.BatchSweepExecutorCollectItem{
		{CallGasLimit: big.NewInt(100_000)},
	}

	got, err := addCollectExecutionBudget(115_924, items)
	if err != nil {
		t.Fatalf("add execution budget: %v", err)
	}
	if want := uint64(245_924); got != want {
		t.Fatalf("gas limit = %d, want %d", got, want)
	}
}

func TestAddCollectExecutionBudgetClampsLikeExecutor(t *testing.T) {
	items := []batchsweepexecutor.BatchSweepExecutorCollectItem{
		{CallGasLimit: big.NewInt(1)},
		{CallGasLimit: big.NewInt(600_000)},
	}

	got, err := addCollectExecutionBudget(100, items)
	if err != nil {
		t.Fatalf("add execution budget: %v", err)
	}
	if want := uint64(600_100); got != want {
		t.Fatalf("gas limit = %d, want %d", got, want)
	}
}

func TestAddCollectExecutionBudgetRejectsInvalidAndOverflow(t *testing.T) {
	tests := []struct {
		name      string
		estimated uint64
		callGas   *big.Int
	}{
		{name: "nil", callGas: nil},
		{name: "negative", callGas: big.NewInt(-1)},
		{name: "too large", callGas: new(big.Int).Lsh(big.NewInt(1), 65)},
		{name: "overflow", estimated: math.MaxUint64, callGas: big.NewInt(100_000)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			items := []batchsweepexecutor.BatchSweepExecutorCollectItem{{CallGasLimit: test.callGas}}
			if _, err := addCollectExecutionBudget(test.estimated, items); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
