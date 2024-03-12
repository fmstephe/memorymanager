package fuzzutil

import (
	"fmt"
)

type TestRun struct {
	steps []Step
}

func NewTestRun(bytes []byte, stepMaker func(*ByteConsumer) Step) *TestRun {
	tr := &TestRun{
		steps: make([]Step, 0),
	}
	byteConsumer := NewByteConsumer(bytes)

	for byteConsumer.Len() > 0 {
		step := stepMaker(byteConsumer)
		tr.steps = append(tr.steps, step)
	}
	return tr
}

func (t *TestRun) Run() {
	fmt.Printf("\nTesting Run with %d steps\n", len(t.steps))
	for _, step := range t.steps {
		step.DoStep()
	}
}

type Step interface {
	DoStep()
}
