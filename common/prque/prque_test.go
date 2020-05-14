package prque

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrque_PopItem(t *testing.T) {
	var triegc *Prque
	triegc = New(nil)
	triegc.Push(31243124, 3)
	triegc.Push(12132321, 1)
	assert.Equal(t, triegc.PopItem(), 31243124)
	triegc.Push(45678899, 2)
	assert.Equal(t, triegc.PopItem(), 45678899)
}

func TestPrque_Pop(t *testing.T) {
	var triegc *Prque
	triegc = New(nil)
	triegc.Push(31243124, 3)
	triegc.Push(12132321, 1)
	res1, _ := triegc.Pop()
	assert.Equal(t, res1, 31243124)
	triegc.Push(45678899, 2)
	res2, _ := triegc.Pop()
	assert.Equal(t, res2, 45678899)
}
