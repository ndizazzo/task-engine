package task_engine_test

import (
	"context"
	"testing"

	engine "github.com/ndizazzo/task-engine"
)

// minimal ResultProvider for tests
type rp struct{ v interface{} }

func (p rp) GetResult() interface{} { return p.v }
func (p rp) GetError() error        { return nil }

func TestParameterResolvers_ResultProviders(t *testing.T) {
	gc := engine.NewGlobalContext()

	// Prepare action result (map) and task result (map)
	gc.StoreActionResult("actR", rp{v: map[string]interface{}{"sum": 10, "name": "demo"}})
	gc.StoreTaskResult("taskR", rp{v: map[string]interface{}{"ok": true, "n": 3}})

	// ActionResultParameter full result
	arp := engine.ActionResult("actR")
	if v, err := arp.Resolve(context.Background(), gc); err != nil {
		t.Fatalf("ActionResult Resolve err: %v", err)
	} else if m, ok := v.(map[string]interface{}); !ok || m["sum"].(int) != 10 {
		t.Fatalf("unexpected action result: %v", v)
	}
	// ActionResultParameter by key
	arpk := engine.ActionResultField("actR", "name")
	if v, err := arpk.Resolve(context.Background(), gc); err != nil || v.(string) != "demo" {
		t.Fatalf("unexpected action result key: v=%v err=%v", v, err)
	}

	// TaskResultParameter full result
	trp := engine.TaskResult("taskR")
	if v, err := trp.Resolve(context.Background(), gc); err != nil {
		t.Fatalf("TaskResult Resolve err: %v", err)
	} else if m, ok := v.(map[string]interface{}); !ok || m["ok"].(bool) != true {
		t.Fatalf("unexpected task result: %v", v)
	}
	// TaskResultParameter by key
	trpk := engine.TaskResultField("taskR", "n")
	if v, err := trpk.Resolve(context.Background(), gc); err != nil || v.(int) != 3 {
		t.Fatalf("unexpected task result key: v=%v err=%v", v, err)
	}
}

func TestEntityOutputParameter_FallbackToResults(t *testing.T) {
	gc := engine.NewGlobalContext()
	// Only a result is present (no output)
	gc.StoreActionResult("A", rp{v: map[string]interface{}{"k": 1}})
	gc.StoreTaskResult("T", rp{v: map[string]interface{}{"s": "ok"}})

	// Action entity, full result
	p1 := engine.EntityOutput("action", "A")
	if v, err := p1.Resolve(context.Background(), gc); err != nil {
		t.Fatalf("EntityOutput(action) err: %v", err)
	} else if m, ok := v.(map[string]interface{}); !ok || m["k"].(int) != 1 {
		t.Fatalf("unexpected value: %v", v)
	}
	// Action entity by key
	p1k := engine.EntityOutputField("action", "A", "k")
	if v, err := p1k.Resolve(context.Background(), gc); err != nil || v.(int) != 1 {
		t.Fatalf("unexpected key value: %v err=%v", v, err)
	}

	// Task entity, full result
	p2 := engine.EntityOutput("task", "T")
	if v, err := p2.Resolve(context.Background(), gc); err != nil {
		t.Fatalf("EntityOutput(task) err: %v", err)
	} else if m, ok := v.(map[string]interface{}); !ok || m["s"].(string) != "ok" {
		t.Fatalf("unexpected value: %v", v)
	}
	// Task entity by key
	p2k := engine.EntityOutputField("task", "T", "s")
	if v, err := p2k.Resolve(context.Background(), gc); err != nil || v.(string) != "ok" {
		t.Fatalf("unexpected key value: %v err=%v", v, err)
	}
}

func TestResolveAs_GenericAdditional(t *testing.T) {
	gc := engine.NewGlobalContext()
	gc.StoreActionOutput("actX", map[string]interface{}{"flag": true, "nums": []string{"a", "b"}})

	b, err := engine.ResolveAs[bool](context.Background(), engine.ActionOutputField("actX", "flag"), gc)
	if err != nil || b != true {
		t.Fatalf("expected true, got %v err=%v", b, err)
	}
	sl, err := engine.ResolveAs[[]string](context.Background(), engine.ActionOutputField("actX", "nums"), gc)
	if err != nil || len(sl) != 2 || sl[0] != "a" {
		t.Fatalf("unexpected slice: %v err=%v", sl, err)
	}
}
