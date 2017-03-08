package context_test

import (
	"testing"
	"time"

	"github.com/drborges/rivers/context"
	. "github.com/drborges/rivers/context/matchers"
	"github.com/drborges/rivers/expectations"
	. "github.com/drborges/rivers/expectations/matchers"
)

func TestClose(t *testing.T) {
	expect := expectations.New()

	ctx := context.New()

	if err := expect(ctx).ToNot(BeClosed()); err != nil {
		t.Error(err)
	}

	ctx.Close()

	if err := expect(ctx).To(BeClosed()); err != nil {
		t.Error(err)
	}
}

func TestParentContextCannotBeClosedWhenChildrenAreStillOpened(t *testing.T) {
	expect := expectations.New()

	parent := context.New()
	child1 := parent.NewChild()
	child2 := parent.NewChild()

	if err := expect(child1).ToNot(BeClosed()); err != nil {
		t.Error(err)
	}

	if err := expect(child2).ToNot(BeClosed()); err != nil {
		t.Error(err)
	}

	parent.Close()

	if err := expect(child1).ToNot(BeClosed()); err != nil {
		t.Error(err)
	}

	if err := expect(child2).ToNot(BeClosed()); err != nil {
		t.Error(err)
	}
}

func TestParentContextIsClosedWhenAllChildrenAreClosed(t *testing.T) {
	expect := expectations.New()

	parent := context.New()
	child1 := parent.NewChild()
	child2 := parent.NewChild()
	grandchild := child2.NewChild()

	child1.Close()

	if err := expect(child1).To(BeClosed()); err != nil {
		t.Error(err)
	}

	if err := expect(parent).ToNot(BeClosed()); err != nil {
		t.Error(err)
	}

	if err := expect(child2).ToNot(BeClosed()); err != nil {
		t.Error(err)
	}

	if err := expect(grandchild).ToNot(BeClosed()); err != nil {
		t.Error(err)
	}

	child2.Close()

	if err := expect(child2).ToNot(BeClosed()); err != nil {
		t.Error(err)
	}

	if err := expect(parent).ToNot(BeClosed()); err != nil {
		t.Error(err)
	}

	grandchild.Close()

	if err := expect(grandchild).To(BeClosed()); err != nil {
		t.Error(err)
	}

	if err := expect(child2).To(BeClosed()); err != nil {
		t.Error(err)
	}

	if err := expect(parent).To(BeClosed()); err != nil {
		t.Error(err)
	}
}

func TestConfigPropagation(t *testing.T) {
	expect := expectations.New()

	parent := context.New()
	child := parent.NewChild()

	if err := expect(parent.Config()).To(Be(child.Config())); err != nil {
		t.Error(err)
	}
}

func TestNewContextWithConfig(t *testing.T) {
	expect := expectations.New()

	config := context.Config{
		Timeout:    1 * time.Second,
		BufferSize: 2,
	}

	ctx := context.WithConfig(context.New(), config)

	if err := expect(ctx.Config()).To(Be(config)); err != nil {
		t.Error(err)
	}

	deadline, ok := ctx.Deadline()
	if err := expect(ok).To(Be(true)); err != nil {
		t.Error(err)
	}

	if err := expect(deadline).ToNot(Be(nil)); err != nil {
		t.Error(err)
	}
}
