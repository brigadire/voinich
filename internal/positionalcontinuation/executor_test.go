package positionalcontinuation

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// executorFixture builds a small SAiinOccurrence/AiinOccurrence set across
// 3 blocks with varied line/block positional categories and predecessors,
// giving every one of the 5 distributable batteries real (if tiny) work.
func executorFixture() ([]SAiinOccurrence, []AiinOccurrence) {
	var sAiin []SAiinOccurrence
	var aiin []AiinOccurrence
	cats := []string{"LINE_START", "LINE_MIDDLE", "LINE_END"}
	blockCats := []string{"BLOCK_START", "BLOCK_MIDDLE", "BLOCK_END"}
	for _, b := range []string{"B1", "B2", "B3"} {
		for i := range 6 {
			x := "chey"
			if i%2 == 0 {
				x = "other"
			}
			lc := cats[i%len(cats)]
			bc := blockCats[i%len(blockCats)]
			sAiin = append(sAiin, SAiinOccurrence{
				Block: b, X: x, LineCategory: lc, BlockBinCoarse: bc,
				TokensFromLineStart: i, TokensToLineEnd: 10 - i,
				TokensFromBlockStart: i, TokensToBlockEnd: 20 - i,
			})
			aiin = append(aiin, AiinOccurrence{
				Block: b, X: x, LineCategory: lc, BlockBinCoarse: bc,
				HasPredecessor: true, PredecessorIsS: i%3 == 0,
			})
		}
	}
	return sAiin, aiin
}

func TestComputeBatteryDeterministicPerBattery(t *testing.T) {
	sAiin, aiin := executorFixture()
	for _, battery := range batteryNames {
		r1, err := ComputeBattery(sAiin, aiin, battery, 50, 1)
		if err != nil {
			t.Fatalf("battery %q: %v", battery, err)
		}
		r2, err := ComputeBattery(sAiin, aiin, battery, 50, 1)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(r1, r2) {
			t.Fatalf("battery %q not deterministic\nr1=%#v\nr2=%#v", battery, r1, r2)
		}
	}
}

func TestComputeBatteryUnknownBattery(t *testing.T) {
	sAiin, aiin := executorFixture()
	if _, err := ComputeBattery(sAiin, aiin, "no-such-battery", 10, 1); err == nil {
		t.Fatal("expected an error for an unknown battery name")
	}
}

func TestComputeBatteryPostestMatchesRunPositionalTests(t *testing.T) {
	sAiin, aiin := executorFixture()
	_ = aiin
	got, err := ComputeBattery(sAiin, aiin, "postest_line", 50, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := runPositionalTests(sAiin, "line_position", lineCategories, 50, seedFor(1, "line_position"))
	if !reflect.DeepEqual(got.Dependence, want.Dependence) || !reflect.DeepEqual(got.Entropy, want.Entropy) || !reflect.DeepEqual(got.CheyEffect, want.CheyEffect) {
		t.Fatalf("postest_line battery diverged from direct runPositionalTests call\ngot=%#v\nwant=%#v", got, want)
	}
}

func TestDefaultBatteryExecutorMatchesComputeBattery(t *testing.T) {
	sAiin, aiin := executorFixture()
	ex := newDefaultBatteryExecutor(sAiin, aiin, 50, 1)
	for _, battery := range batteryNames {
		got, err := ex.Run(context.Background(), battery)
		if err != nil {
			t.Fatal(err)
		}
		want, err := ComputeBattery(sAiin, aiin, battery, 50, 1)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("battery %q: default executor diverged from ComputeBattery\ngot=%#v\nwant=%#v", battery, got, want)
		}
	}
}

func TestRunBatteryDispatchRestoresCanonicalOrder(t *testing.T) {
	sAiin, aiin := executorFixture()
	ex := newDefaultBatteryExecutor(sAiin, aiin, 50, 1)

	var orderSeen []int
	err := runBatteryDispatch(context.Background(), ex, batteryNames, len(batteryNames), func(i int, battery string, r BatteryResult) error {
		orderSeen = append(orderSeen, i)
		if battery != batteryNames[i] {
			return fmt.Errorf("onReady got mismatched battery: %q at index %d", battery, i)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range orderSeen {
		if v != i {
			t.Fatalf("onReady was not called in ascending canonical order: %v", orderSeen)
		}
	}
}

func TestRunBatteryDispatchPropagatesWorkerError(t *testing.T) {
	stub := stubBatteryExecutor{fail: map[string]bool{"bad": true}}
	err := runBatteryDispatch(context.Background(), stub, []string{"postest_line", "bad"}, 2, func(int, string, BatteryResult) error { return nil })
	if err == nil {
		t.Fatal("expected an error to propagate from a failing battery job")
	}
}

func TestRunBatteryDispatchPropagatesOnReadyError(t *testing.T) {
	stub := stubBatteryExecutor{}
	boom := errors.New("boom")
	err := runBatteryDispatch(context.Background(), stub, []string{"postest_line"}, 1, func(int, string, BatteryResult) error { return boom })
	if !errors.Is(err, boom) {
		t.Fatalf("expected onReady's error to propagate, got %v", err)
	}
}

// stubBatteryExecutor is a minimal BatteryExecutor for testing
// runBatteryDispatch's dispatch/ordering/error-propagation in isolation
// from the real per-battery computation.
type stubBatteryExecutor struct {
	fail map[string]bool
}

func (s stubBatteryExecutor) Run(_ context.Context, battery string) (BatteryResult, error) {
	if s.fail[battery] {
		return BatteryResult{}, fmt.Errorf("stub failure for %q", battery)
	}
	return BatteryResult{Stratified: StratifiedPredecessorRow{PositionVariable: battery}}, nil
}
