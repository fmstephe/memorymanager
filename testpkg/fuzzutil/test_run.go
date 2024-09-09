// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file.

package fuzzutil

type TestRun struct {
	steps   []Step
	cleanup func()
}

func NewTestRun(bytes []byte, stepMaker func(*ByteConsumer) Step, cleanup func()) *TestRun {
	tr := &TestRun{
		steps:   make([]Step, 0),
		cleanup: cleanup,
	}
	byteConsumer := NewByteConsumer(bytes)

	for byteConsumer.Len() > 0 {
		step := stepMaker(byteConsumer)
		tr.steps = append(tr.steps, step)
	}
	return tr
}

func (t *TestRun) Run() {
	//fmt.Printf("\nTesting Run with %d steps\n", len(t.steps))
	defer t.cleanup()
	for _, step := range t.steps {
		step.DoStep()
	}
}

type Step interface {
	DoStep()
}
