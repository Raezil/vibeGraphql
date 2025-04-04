package vibeGraphql

import "testing"

// dummyResolver remains unchanged
func dummyResolvers(source interface{}, args map[string]interface{}) (interface{}, error) {
	return "dummy success", nil
}

func TestRegisterQueryResolver(t *testing.T) {
	field := "testQuery"
	RegisterQueryResolver(field, dummyResolvers)

	resolver, exists := QueryResolvers[field]
	if !exists {
		t.Fatalf("expected resolver for field %q to be registered", field)
	}

	result, err := resolver(nil, nil)
	if err != nil {
		t.Errorf("expected no error from resolver, got %v", err)
	}
	if result != "dummy success" {
		t.Errorf("expected result 'dummy success', got %v", result)
	}
}

func TestRegisterMutationResolver(t *testing.T) {
	field := "testMutation"
	RegisterMutationResolver(field, dummyResolvers)

	resolver, exists := MutationResolvers[field]
	if !exists {
		t.Fatalf("expected resolver for field %q to be registered", field)
	}

	result, err := resolver(nil, nil)
	if err != nil {
		t.Errorf("expected no error from resolver, got %v", err)
	}
	if result != "dummy success" {
		t.Errorf("expected result 'dummy success', got %v", result)
	}
}

func TestRegisterSubscriptionResolver(t *testing.T) {
	field := "testSubscription"
	RegisterSubscriptionResolver(field, dummyResolvers)

	resolver, exists := SubscriptionResolvers[field]
	if !exists {
		t.Fatalf("expected resolver for field %q to be registered", field)
	}

	result, err := resolver(nil, nil)
	if err != nil {
		t.Errorf("expected no error from resolver, got %v", err)
	}
	if result != "dummy success" {
		t.Errorf("expected result 'dummy success', got %v", result)
	}
}
