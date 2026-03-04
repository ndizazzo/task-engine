package task_engine_test

import (
	"context"
	"fmt"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
)

func TestParameterPassingSystem(t *testing.T) {
	t.Run("StaticParameter", func(t *testing.T) {
		staticParam := task_engine.StaticParameter{Value: "test value"}
		globalContext := task_engine.NewGlobalContext()

		result, err := staticParam.Resolve(context.Background(), globalContext)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != "test value" {
			t.Fatalf("Expected 'test value', got %v", result)
		}
	})
	t.Run("ActionOutputParameter", func(t *testing.T) {
		globalContext := task_engine.NewGlobalContext()
		globalContext.StoreActionOutput("test-action", map[string]interface{}{
			"content": "file content",
			"size":    12,
		})
		param := task_engine.ActionOutputParameter{ActionID: "test-action"}
		result, err := param.Resolve(context.Background(), globalContext)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		expected := map[string]interface{}{
			"content": "file content",
			"size":    12,
		}
		if result == nil {
			t.Fatalf("Expected non-nil result, got nil")
		}
		m, ok := result.(map[string]interface{})
		if !ok || m["content"] != expected["content"] || m["size"] != expected["size"] {
			t.Fatalf("Expected %v, got %v", expected, result)
		}
		paramWithKey := task_engine.ActionOutputParameter{ActionID: "test-action", OutputKey: "content"}
		result, err = paramWithKey.Resolve(context.Background(), globalContext)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != "file content" {
			t.Fatalf("Expected 'file content', got %v", result)
		}
	})
	t.Run("TaskOutputParameter", func(t *testing.T) {
		globalContext := task_engine.NewGlobalContext()
		globalContext.StoreTaskOutput("test-task", map[string]interface{}{
			"result": "task result",
			"status": "completed",
		})

		param := task_engine.TaskOutputParameter{TaskID: "test-task", OutputKey: "result"}
		result, err := param.Resolve(context.Background(), globalContext)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != "task result" {
			t.Fatalf("Expected 'task result', got %v", result)
		}
	})
	t.Run("EntityOutputParameter", func(t *testing.T) {
		globalContext := task_engine.NewGlobalContext()
		globalContext.StoreActionOutput("test-action", "action output")
		globalContext.StoreTaskOutput("test-task", "task output")
		actionParam := task_engine.EntityOutputParameter{EntityType: "action", EntityID: "test-action"}
		result, err := actionParam.Resolve(context.Background(), globalContext)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != "action output" {
			t.Fatalf("Expected 'action output', got %v", result)
		}
		taskParam := task_engine.EntityOutputParameter{EntityType: "task", EntityID: "test-task"}
		result, err = taskParam.Resolve(context.Background(), globalContext)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != "task output" {
			t.Fatalf("Expected 'task output', got %v", result)
		}
	})
	t.Run("HelperFunctions", func(t *testing.T) {
		param1 := task_engine.ActionOutput("test-action")
		if param1.ActionID != "test-action" {
			t.Fatalf("Expected ActionID 'test-action', got %s", param1.ActionID)
		}
		if param1.OutputKey != "" {
			t.Fatalf("Expected empty OutputKey, got %s", param1.OutputKey)
		}
		param2 := task_engine.ActionOutputField("test-action", "content")
		if param2.ActionID != "test-action" {
			t.Fatalf("Expected ActionID 'test-action', got %s", param2.ActionID)
		}
		if param2.OutputKey != "content" {
			t.Fatalf("Expected OutputKey 'content', got %s", param2.OutputKey)
		}
		param3 := task_engine.TaskOutput("test-task")
		if param3.TaskID != "test-task" {
			t.Fatalf("Expected TaskID 'test-task', got %s", param3.TaskID)
		}
		if param3.OutputKey != "" {
			t.Fatalf("Expected empty OutputKey, got %s", param3.OutputKey)
		}
	})
}

func TestGlobalContext(t *testing.T) {
	t.Run("GlobalContextOperations", func(t *testing.T) {
		gc := task_engine.NewGlobalContext()
		gc.StoreActionOutput("action1", "output1")
		if gc.ActionOutputs["action1"] != "output1" {
			t.Fatalf("Expected 'output1', got %v", gc.ActionOutputs["action1"])
		}
		gc.StoreTaskOutput("task1", "output1")
		if gc.TaskOutputs["task1"] != "output1" {
			t.Fatalf("Expected 'output1', got %v", gc.TaskOutputs["task1"])
		}
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(id int) {
				gc.StoreActionOutput(fmt.Sprintf("action%d", id), fmt.Sprintf("output%d", id))
				done <- true
			}(i)
		}

		for i := 0; i < 10; i++ {
			<-done
		}
		for i := 0; i < 10; i++ {
			expected := fmt.Sprintf("output%d", i)
			actual := gc.ActionOutputs[fmt.Sprintf("action%d", i)]
			if actual != expected {
				t.Fatalf("Expected %s, got %v", expected, actual)
			}
		}
	})
}

func TestTypedGlobalContextHelpers(t *testing.T) {
	gc := task_engine.NewGlobalContext()

	gc.StoreActionOutput("act1", map[string]interface{}{"k": 123, "s": "abc"})
	gc.StoreActionResult("actRes", testResultProvider{v: map[string]interface{}{"sum": 7}})
	gc.StoreTaskOutput("task1", map[string]interface{}{"ok": true, "n": 9})
	gc.StoreTaskResult("taskRes", testResultProvider{v: "done"})

	// ActionOutputFieldAs
	vInt, err := task_engine.ActionOutputFieldAs[int](gc, "act1", "k")
	if err != nil || vInt != 123 {
		t.Fatalf("expected 123, got %v, err=%v", vInt, err)
	}
	vStr, err := task_engine.ActionOutputFieldAs[string](gc, "act1", "s")
	if err != nil || vStr != "abc" {
		t.Fatalf("expected 'abc', got %v, err=%v", vStr, err)
	}

	// TaskOutputFieldAs
	vBool, err := task_engine.TaskOutputFieldAs[bool](gc, "task1", "ok")
	if err != nil || vBool != true {
		t.Fatalf("expected true, got %v, err=%v", vBool, err)
	}
	vNum, err := task_engine.TaskOutputFieldAs[int](gc, "task1", "n")
	if err != nil || vNum != 9 {
		t.Fatalf("expected 9, got %v, err=%v", vNum, err)
	}

	// ActionResultAs / TaskResultAs
	rmap, ok := task_engine.ActionResultAs[map[string]interface{}](gc, "actRes")
	if !ok || rmap["sum"].(int) != 7 {
		t.Fatalf("expected action result sum=7, got %v", rmap)
	}
	rstr, ok := task_engine.TaskResultAs[string](gc, "taskRes")
	if !ok || rstr != "done" {
		t.Fatalf("expected task result 'done', got %v", rstr)
	}

	// EntityValue / EntityValueAs
	if v, err := task_engine.EntityValue(gc, "action", "act1", "k"); err != nil || v.(int) != 123 {
		t.Fatalf("EntityValue action k expected 123, got %v, err=%v", v, err)
	}
	if v, err := task_engine.EntityValue(gc, "task", "task1", "ok"); err != nil || v.(bool) != true {
		t.Fatalf("EntityValue task ok expected true, got %v, err=%v", v, err)
	}
	if v, err := task_engine.EntityValue(gc, "action", "actRes", ""); err != nil {
		t.Fatalf("EntityValue action result expected no error, got err=%v", err)
	} else {
		if vm, ok := v.(map[string]interface{}); !ok || vm["sum"].(int) != 7 {
			t.Fatalf("EntityValue action result expected map with sum=7, got %v", v)
		}
	}
	if s, err := task_engine.EntityValueAs[string](gc, "task", "taskRes", ""); err != nil || s != "done" {
		t.Fatalf("EntityValueAs task result expected 'done', got %v, err=%v", s, err)
	}
}

func TestResolveAsGeneric(t *testing.T) {
	gc := task_engine.NewGlobalContext()
	gc.StoreActionOutput("act", map[string]interface{}{"name": "demo", "count": 5})

	name, err := task_engine.ResolveAs[string](context.Background(), task_engine.ActionOutputField("act", "name"), gc)
	if err != nil || name != "demo" {
		t.Fatalf("expected 'demo', got %v, err=%v", name, err)
	}
	count, err := task_engine.ResolveAs[int](context.Background(), task_engine.ActionOutputField("act", "count"), gc)
	if err != nil || count != 5 {
		t.Fatalf("expected 5, got %v, err=%v", count, err)
	}
}

func TestEntityValueNegativePaths(t *testing.T) {
	gc := task_engine.NewGlobalContext()

	if _, err := task_engine.EntityValue(gc, "invalid", "id", ""); err == nil {
		t.Fatalf("expected error for invalid entity type")
	}
	if _, err := task_engine.EntityValue(gc, "action", "missing", ""); err == nil {
		t.Fatalf("expected error for missing action")
	}
	gc.StoreActionOutput("a1", map[string]interface{}{"k": 1})
	if _, err := task_engine.ActionOutputFieldAs[string](gc, "a1", "k"); err == nil {
		t.Fatalf("expected type error for wrong cast")
	}
}

func TestResolveAsNegative(t *testing.T) {
	gc := task_engine.NewGlobalContext()
	gc.StoreActionOutput("a", map[string]interface{}{"x": "str"})
	if _, err := task_engine.ResolveAs[int](context.Background(), task_engine.ActionOutputField("a", "x"), gc); err == nil {
		t.Fatalf("expected type error for ResolveAs")
	}
}

func TestTaskOutputFieldHelper(t *testing.T) {
	p := task_engine.TaskOutputField("my-task", "result")
	if p.TaskID != "my-task" {
		t.Fatalf("expected TaskID 'my-task', got %q", p.TaskID)
	}
	if p.OutputKey != "result" {
		t.Fatalf("expected OutputKey 'result', got %q", p.OutputKey)
	}
}

func TestTaskResultAndActionResultHelpers(t *testing.T) {
	p := task_engine.TaskResult("t1")
	if p.TaskID != "t1" {
		t.Fatalf("expected TaskID 't1', got %q", p.TaskID)
	}
	if p.ResultKey != "" {
		t.Fatalf("expected empty ResultKey, got %q", p.ResultKey)
	}

	pf := task_engine.TaskResultField("t1", "key1")
	if pf.TaskID != "t1" || pf.ResultKey != "key1" {
		t.Fatalf("unexpected TaskResultField: %+v", pf)
	}

	ap := task_engine.ActionResult("a1")
	if ap.ActionID != "a1" {
		t.Fatalf("expected ActionID 'a1', got %q", ap.ActionID)
	}

	apf := task_engine.ActionResultField("a1", "f1")
	if apf.ActionID != "a1" || apf.ResultKey != "f1" {
		t.Fatalf("unexpected ActionResultField: %+v", apf)
	}
}

func TestEntityOutputHelpers(t *testing.T) {
	p := task_engine.EntityOutput("action", "a1")
	if p.EntityType != "action" || p.EntityID != "a1" || p.OutputKey != "" {
		t.Fatalf("unexpected EntityOutput: %+v", p)
	}

	pf := task_engine.EntityOutputField("task", "t1", "k1")
	if pf.EntityType != "task" || pf.EntityID != "t1" || pf.OutputKey != "k1" {
		t.Fatalf("unexpected EntityOutputField: %+v", pf)
	}
}

func TestActionOutputParameterErrors(t *testing.T) {
	ctx := context.Background()
	gc := task_engine.NewGlobalContext()

	t.Run("nil globalContext", func(t *testing.T) {
		p := task_engine.ActionOutputParameter{ActionID: "a1"}
		_, err := p.Resolve(ctx, nil)
		if err == nil {
			t.Fatal("expected error for nil globalContext")
		}
	})
	t.Run("empty ActionID", func(t *testing.T) {
		p := task_engine.ActionOutputParameter{ActionID: ""}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for empty ActionID")
		}
	})
	t.Run("missing action", func(t *testing.T) {
		p := task_engine.ActionOutputParameter{ActionID: "missing"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for missing action")
		}
	})
	t.Run("output not a map with key", func(t *testing.T) {
		gc.StoreActionOutput("scalar", "just a string")
		p := task_engine.ActionOutputParameter{ActionID: "scalar", OutputKey: "field"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error when output is not a map and key is requested")
		}
	})
	t.Run("key not found in map", func(t *testing.T) {
		gc.StoreActionOutput("mapped", map[string]interface{}{"a": 1})
		p := task_engine.ActionOutputParameter{ActionID: "mapped", OutputKey: "missing_key"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for missing key in map")
		}
	})
}

func TestActionResultParameterErrors(t *testing.T) {
	ctx := context.Background()
	gc := task_engine.NewGlobalContext()

	t.Run("nil globalContext", func(t *testing.T) {
		p := task_engine.ActionResultParameter{ActionID: "a1"}
		_, err := p.Resolve(ctx, nil)
		if err == nil {
			t.Fatal("expected error for nil globalContext")
		}
	})
	t.Run("empty ActionID", func(t *testing.T) {
		p := task_engine.ActionResultParameter{ActionID: ""}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for empty ActionID")
		}
	})
	t.Run("missing action result", func(t *testing.T) {
		p := task_engine.ActionResultParameter{ActionID: "missing"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for missing action result")
		}
	})
	t.Run("result not a map with key", func(t *testing.T) {
		gc.StoreActionResult("scalar-res", testResultProvider{v: "just a string"})
		p := task_engine.ActionResultParameter{ActionID: "scalar-res", ResultKey: "field"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error when result is not a map and key is requested")
		}
	})
	t.Run("key not found in result map", func(t *testing.T) {
		gc.StoreActionResult("mapped-res", testResultProvider{v: map[string]interface{}{"a": 1}})
		p := task_engine.ActionResultParameter{ActionID: "mapped-res", ResultKey: "nope"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for missing key in result map")
		}
	})
}

func TestTaskResultParameterErrors(t *testing.T) {
	ctx := context.Background()
	gc := task_engine.NewGlobalContext()

	t.Run("nil globalContext", func(t *testing.T) {
		p := task_engine.TaskResultParameter{TaskID: "t1"}
		_, err := p.Resolve(ctx, nil)
		if err == nil {
			t.Fatal("expected error for nil globalContext")
		}
	})
	t.Run("empty TaskID", func(t *testing.T) {
		p := task_engine.TaskResultParameter{TaskID: ""}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for empty TaskID")
		}
	})
	t.Run("missing task result", func(t *testing.T) {
		p := task_engine.TaskResultParameter{TaskID: "missing"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for missing task result")
		}
	})
	t.Run("result not a map with key", func(t *testing.T) {
		gc.StoreTaskResult("scalar-tr", testResultProvider{v: 42})
		p := task_engine.TaskResultParameter{TaskID: "scalar-tr", ResultKey: "field"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error when task result is not a map and key is requested")
		}
	})
	t.Run("key not found in result map", func(t *testing.T) {
		gc.StoreTaskResult("mapped-tr", testResultProvider{v: map[string]interface{}{"x": 1}})
		p := task_engine.TaskResultParameter{TaskID: "mapped-tr", ResultKey: "missing"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for missing key in task result map")
		}
	})
}

func TestTaskOutputParameterErrors(t *testing.T) {
	ctx := context.Background()
	gc := task_engine.NewGlobalContext()

	t.Run("nil globalContext", func(t *testing.T) {
		p := task_engine.TaskOutputParameter{TaskID: "t1"}
		_, err := p.Resolve(ctx, nil)
		if err == nil {
			t.Fatal("expected error for nil globalContext")
		}
	})
	t.Run("empty TaskID", func(t *testing.T) {
		p := task_engine.TaskOutputParameter{TaskID: ""}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for empty TaskID")
		}
	})
	t.Run("missing task output", func(t *testing.T) {
		p := task_engine.TaskOutputParameter{TaskID: "missing"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for missing task output")
		}
	})
	t.Run("output not a map with key", func(t *testing.T) {
		gc.StoreTaskOutput("scalar-to", "not a map")
		p := task_engine.TaskOutputParameter{TaskID: "scalar-to", OutputKey: "field"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error when task output is not a map and key is requested")
		}
	})
	t.Run("key not found in output map", func(t *testing.T) {
		gc.StoreTaskOutput("mapped-to", map[string]interface{}{"a": 1})
		p := task_engine.TaskOutputParameter{TaskID: "mapped-to", OutputKey: "nope"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for missing key in task output map")
		}
	})
}

func TestEntityOutputParameterErrors(t *testing.T) {
	ctx := context.Background()
	gc := task_engine.NewGlobalContext()

	t.Run("nil globalContext", func(t *testing.T) {
		p := task_engine.EntityOutputParameter{EntityType: "action", EntityID: "a1"}
		_, err := p.Resolve(ctx, nil)
		if err == nil {
			t.Fatal("expected error for nil globalContext")
		}
	})
	t.Run("empty EntityType and EntityID", func(t *testing.T) {
		p := task_engine.EntityOutputParameter{EntityType: "", EntityID: ""}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for empty EntityType/EntityID")
		}
	})
	t.Run("invalid entity type", func(t *testing.T) {
		p := task_engine.EntityOutputParameter{EntityType: "bogus", EntityID: "x"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for invalid entity type")
		}
	})
	t.Run("action not found at all", func(t *testing.T) {
		p := task_engine.EntityOutputParameter{EntityType: "action", EntityID: "nope"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for missing action")
		}
	})
	t.Run("task not found at all", func(t *testing.T) {
		p := task_engine.EntityOutputParameter{EntityType: "task", EntityID: "nope"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for missing task")
		}
	})
	t.Run("action output not a map with key", func(t *testing.T) {
		gc.StoreActionOutput("entity-scalar", "string_val")
		p := task_engine.EntityOutputParameter{EntityType: "action", EntityID: "entity-scalar", OutputKey: "k"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error when entity action output is not a map and key is requested")
		}
	})
	t.Run("action result fallback with key not a map", func(t *testing.T) {
		gc.StoreActionResult("entity-res-scalar", testResultProvider{v: 99})
		p := task_engine.EntityOutputParameter{EntityType: "action", EntityID: "entity-res-scalar", OutputKey: "k"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error when action result is not a map and key is requested")
		}
	})
	t.Run("task output not a map with key", func(t *testing.T) {
		gc.StoreTaskOutput("entity-task-scalar", "string_val")
		p := task_engine.EntityOutputParameter{EntityType: "task", EntityID: "entity-task-scalar", OutputKey: "k"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error when entity task output is not a map and key is requested")
		}
	})
	t.Run("task result fallback with key not a map", func(t *testing.T) {
		gc.StoreTaskResult("entity-tres-scalar", testResultProvider{v: 99})
		p := task_engine.EntityOutputParameter{EntityType: "task", EntityID: "entity-tres-scalar", OutputKey: "k"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error when task result is not a map and key is requested")
		}
	})
	t.Run("action result fallback key not found in map", func(t *testing.T) {
		gc.StoreActionResult("entity-res-map", testResultProvider{v: map[string]interface{}{"x": 1}})
		p := task_engine.EntityOutputParameter{EntityType: "action", EntityID: "entity-res-map", OutputKey: "nope"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for missing key in action result map")
		}
	})
	t.Run("task result fallback key not found in map", func(t *testing.T) {
		gc.StoreTaskResult("entity-tres-map", testResultProvider{v: map[string]interface{}{"x": 1}})
		p := task_engine.EntityOutputParameter{EntityType: "task", EntityID: "entity-tres-map", OutputKey: "nope"}
		_, err := p.Resolve(ctx, gc)
		if err == nil {
			t.Fatal("expected error for missing key in task result map")
		}
	})
}

func TestResolveStringEdgeCases(t *testing.T) {
	gc := task_engine.NewGlobalContext()
	ctx := context.Background()

	t.Run("nil parameter", func(t *testing.T) {
		v, err := task_engine.ResolveString(ctx, nil, gc)
		if err != nil || v != "" {
			t.Fatalf("expected empty string, got %q, err=%v", v, err)
		}
	})
	t.Run("byte slice", func(t *testing.T) {
		gc.StoreActionOutput("bytes", []byte("hello"))
		p := task_engine.ActionOutput("bytes")
		v, err := task_engine.ResolveString(ctx, p, gc)
		if err != nil || v != "hello" {
			t.Fatalf("expected 'hello', got %q, err=%v", v, err)
		}
	})
	t.Run("int value", func(t *testing.T) {
		gc.StoreActionOutput("intval", 42)
		p := task_engine.ActionOutput("intval")
		v, err := task_engine.ResolveString(ctx, p, gc)
		if err != nil || v != "42" {
			t.Fatalf("expected '42', got %q, err=%v", v, err)
		}
	})
	t.Run("non-stringable value", func(t *testing.T) {
		gc.StoreActionOutput("struct", struct{ X int }{X: 1})
		p := task_engine.ActionOutput("struct")
		_, err := task_engine.ResolveString(ctx, p, gc)
		if err == nil {
			t.Fatal("expected error for non-stringable type")
		}
	})
}

func TestResolveBoolEdgeCases(t *testing.T) {
	gc := task_engine.NewGlobalContext()
	ctx := context.Background()

	t.Run("nil parameter", func(t *testing.T) {
		v, err := task_engine.ResolveBool(ctx, nil, gc)
		if err != nil || v != false {
			t.Fatalf("expected false, got %v, err=%v", v, err)
		}
	})
	t.Run("string yes", func(t *testing.T) {
		gc.StoreActionOutput("yes", "yes")
		v, err := task_engine.ResolveBool(ctx, task_engine.ActionOutput("yes"), gc)
		if err != nil || v != true {
			t.Fatalf("expected true, got %v, err=%v", v, err)
		}
	})
	t.Run("string no", func(t *testing.T) {
		gc.StoreActionOutput("no", "no")
		v, err := task_engine.ResolveBool(ctx, task_engine.ActionOutput("no"), gc)
		if err != nil || v != false {
			t.Fatalf("expected false, got %v, err=%v", v, err)
		}
	})
	t.Run("string invalid", func(t *testing.T) {
		gc.StoreActionOutput("maybe", "maybe")
		_, err := task_engine.ResolveBool(ctx, task_engine.ActionOutput("maybe"), gc)
		if err == nil {
			t.Fatal("expected error for invalid bool string")
		}
	})
	t.Run("int nonzero", func(t *testing.T) {
		gc.StoreActionOutput("one", 1)
		v, err := task_engine.ResolveBool(ctx, task_engine.ActionOutput("one"), gc)
		if err != nil || v != true {
			t.Fatalf("expected true, got %v, err=%v", v, err)
		}
	})
	t.Run("unsupported type", func(t *testing.T) {
		gc.StoreActionOutput("slice", []string{"a"})
		_, err := task_engine.ResolveBool(ctx, task_engine.ActionOutput("slice"), gc)
		if err == nil {
			t.Fatal("expected error for unsupported bool type")
		}
	})
}

func TestResolveStringSliceEdgeCases(t *testing.T) {
	gc := task_engine.NewGlobalContext()
	ctx := context.Background()

	t.Run("nil parameter", func(t *testing.T) {
		v, err := task_engine.ResolveStringSlice(ctx, nil, gc)
		if err != nil || v != nil {
			t.Fatalf("expected nil, got %v, err=%v", v, err)
		}
	})
	t.Run("string slice direct", func(t *testing.T) {
		gc.StoreActionOutput("ss", []string{"a", "b"})
		v, err := task_engine.ResolveStringSlice(ctx, task_engine.ActionOutput("ss"), gc)
		if err != nil || len(v) != 2 {
			t.Fatalf("expected [a b], got %v, err=%v", v, err)
		}
	})
	t.Run("comma separated", func(t *testing.T) {
		gc.StoreActionOutput("csv", "a, b, c")
		v, err := task_engine.ResolveStringSlice(ctx, task_engine.ActionOutput("csv"), gc)
		if err != nil || len(v) != 3 || v[0] != "a" {
			t.Fatalf("expected [a b c], got %v, err=%v", v, err)
		}
	})
	t.Run("space separated", func(t *testing.T) {
		gc.StoreActionOutput("spaces", "a b c")
		v, err := task_engine.ResolveStringSlice(ctx, task_engine.ActionOutput("spaces"), gc)
		if err != nil || len(v) != 3 {
			t.Fatalf("expected [a b c], got %v, err=%v", v, err)
		}
	})
	t.Run("empty string", func(t *testing.T) {
		gc.StoreActionOutput("empty", "")
		v, err := task_engine.ResolveStringSlice(ctx, task_engine.ActionOutput("empty"), gc)
		if err != nil || len(v) != 0 {
			t.Fatalf("expected empty slice, got %v, err=%v", v, err)
		}
	})
	t.Run("unsupported type", func(t *testing.T) {
		gc.StoreActionOutput("int-val", 42)
		_, err := task_engine.ResolveStringSlice(ctx, task_engine.ActionOutput("int-val"), gc)
		if err == nil {
			t.Fatal("expected error for non-slice/string type")
		}
	})
}

func TestActionResultAsMissingAndNil(t *testing.T) {
	gc := task_engine.NewGlobalContext()

	_, ok := task_engine.ActionResultAs[string](gc, "missing")
	if ok {
		t.Fatal("expected ok=false for missing action result")
	}

	_, ok = task_engine.TaskResultAs[string](gc, "missing")
	if ok {
		t.Fatal("expected ok=false for missing task result")
	}
}

func TestTypedHelperNoFallbackForTaskOutputFieldAs(t *testing.T) {
	gc := task_engine.NewGlobalContext()
	// Only set task result, no task output
	gc.StoreTaskResult("t1", testResultProvider{v: map[string]interface{}{"v": 1}})
	if _, err := task_engine.TaskOutputFieldAs[int](gc, "t1", "v"); err == nil {
		t.Fatalf("expected error since TaskOutputFieldAs should not fallback to results")
	}
	// But EntityValue should fallback to results and succeed (full result)
	if v, err := task_engine.EntityValue(gc, "task", "t1", ""); err != nil {
		t.Fatalf("expected EntityValue to return fallback result, err=%v", err)
	} else {
		if m, ok := v.(map[string]interface{}); !ok || m["v"].(int) != 1 {
			t.Fatalf("unexpected result fallback: %v", v)
		}
	}
	// And with a key, EntityValue should read from result map
	if v, err := task_engine.EntityValue(gc, "task", "t1", "v"); err != nil || v.(int) != 1 {
		t.Fatalf("expected EntityValue with key to read from result map, got %v, err=%v", v, err)
	}
}

// --- From parameters_resolve_test.go ---

func TestParameterResolvers_ResultProviders(t *testing.T) {
	gc := task_engine.NewGlobalContext()

	gc.StoreActionResult("actR", testResultProvider{v: map[string]interface{}{"sum": 10, "name": "demo"}})
	gc.StoreTaskResult("taskR", testResultProvider{v: map[string]interface{}{"ok": true, "n": 3}})

	// ActionResultParameter full result
	arp := task_engine.ActionResult("actR")
	if v, err := arp.Resolve(context.Background(), gc); err != nil {
		t.Fatalf("ActionResult Resolve err: %v", err)
	} else if m, ok := v.(map[string]interface{}); !ok || m["sum"].(int) != 10 {
		t.Fatalf("unexpected action result: %v", v)
	}
	// ActionResultParameter by key
	arpk := task_engine.ActionResultField("actR", "name")
	if v, err := arpk.Resolve(context.Background(), gc); err != nil || v.(string) != "demo" {
		t.Fatalf("unexpected action result key: v=%v err=%v", v, err)
	}

	// TaskResultParameter full result
	trp := task_engine.TaskResult("taskR")
	if v, err := trp.Resolve(context.Background(), gc); err != nil {
		t.Fatalf("TaskResult Resolve err: %v", err)
	} else if m, ok := v.(map[string]interface{}); !ok || m["ok"].(bool) != true {
		t.Fatalf("unexpected task result: %v", v)
	}
	// TaskResultParameter by key
	trpk := task_engine.TaskResultField("taskR", "n")
	if v, err := trpk.Resolve(context.Background(), gc); err != nil || v.(int) != 3 {
		t.Fatalf("unexpected task result key: v=%v err=%v", v, err)
	}
}

func TestEntityOutputParameter_FallbackToResults(t *testing.T) {
	gc := task_engine.NewGlobalContext()
	// Only a result is present (no output)
	gc.StoreActionResult("A", testResultProvider{v: map[string]interface{}{"k": 1}})
	gc.StoreTaskResult("T", testResultProvider{v: map[string]interface{}{"s": "ok"}})

	// Action entity, full result
	p1 := task_engine.EntityOutput("action", "A")
	if v, err := p1.Resolve(context.Background(), gc); err != nil {
		t.Fatalf("EntityOutput(action) err: %v", err)
	} else if m, ok := v.(map[string]interface{}); !ok || m["k"].(int) != 1 {
		t.Fatalf("unexpected value: %v", v)
	}
	// Action entity by key
	p1k := task_engine.EntityOutputField("action", "A", "k")
	if v, err := p1k.Resolve(context.Background(), gc); err != nil || v.(int) != 1 {
		t.Fatalf("unexpected key value: %v err=%v", v, err)
	}

	// Task entity, full result
	p2 := task_engine.EntityOutput("task", "T")
	if v, err := p2.Resolve(context.Background(), gc); err != nil {
		t.Fatalf("EntityOutput(task) err: %v", err)
	} else if m, ok := v.(map[string]interface{}); !ok || m["s"].(string) != "ok" {
		t.Fatalf("unexpected value: %v", v)
	}
	// Task entity by key
	p2k := task_engine.EntityOutputField("task", "T", "s")
	if v, err := p2k.Resolve(context.Background(), gc); err != nil || v.(string) != "ok" {
		t.Fatalf("unexpected key value: %v err=%v", v, err)
	}
}

func TestResolveAs_GenericAdditional(t *testing.T) {
	gc := task_engine.NewGlobalContext()
	gc.StoreActionOutput("actX", map[string]interface{}{"flag": true, "nums": []string{"a", "b"}})

	b, err := task_engine.ResolveAs[bool](context.Background(), task_engine.ActionOutputField("actX", "flag"), gc)
	if err != nil || b != true {
		t.Fatalf("expected true, got %v err=%v", b, err)
	}
	sl, err := task_engine.ResolveAs[[]string](context.Background(), task_engine.ActionOutputField("actX", "nums"), gc)
	if err != nil || len(sl) != 2 || sl[0] != "a" {
		t.Fatalf("unexpected slice: %v err=%v", sl, err)
	}
}
