package merge

import (
	"context"
	"testing"
)

var testCases = []struct {
	Name     string
	In       [][]int
	Expected []int
}{
	{
		Name:     "TestEmpty",
		In:       [][]int{},
		Expected: []int{},
	},
	{
		Name: "TestSimple",
		In: [][]int{
			[]int{0, 1, 2, 3, 4},
			[]int{0, 1, 2, 3, 4},
			[]int{0, 1, 2, 3, 4},
		},
		Expected: []int{0, 0, 0, 1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4},
	},
}

func TestMergeSlices(t *testing.T) {
	for _, testCase := range testCases {
		pass := t.Run(testCase.Name+"Slice", func(t *testing.T) {
			out := MergeSlices(func(i, j int) bool { return i < j }, testCase.In...)
			if len(testCase.Expected) != len(out) {
				t.Fail()
			} else {
				var i int
				for i < len(testCase.Expected) && !t.Failed() {
					if testCase.Expected[i] != out[i] {
						t.Fail()
					}
					i++
				}
			}
			if t.Failed() {
				t.Log("Expected value differs from output")
				t.Logf("Expected: %v", testCase.Expected)
				t.Logf("Output: %v", out)
			}
		})
		if !pass {
			t.Fail()
		}
	}
}

func TestMergeChannels(t *testing.T) {
	for _, testCase := range testCases {
		pass := t.Run(testCase.Name+"Channel", func(t *testing.T) {
			channels := make([]chan int, len(testCase.In))
			for i, _ := range channels {
				channels[i] = make(chan int, len(testCase.In[i]))
				for j, _ := range testCase.In[i] {
					channels[i] <- testCase.In[i][j]
				}
				close(channels[i])
			}
			outChan := MergeChannels(context.Background(), func(i, j int) bool { return i < j }, channels...)
			var out []int
			for v := range outChan {
				out = append(out, v)
			}
			if len(testCase.Expected) != len(out) {
				t.Fail()
			} else {
				var i int
				for i < len(testCase.Expected) && !t.Failed() {
					if testCase.Expected[i] != out[i] {
						t.Fail()
					}
					i++
				}
			}
			if t.Failed() {
				t.Log("Expected value differs from output")
				t.Logf("Expected: %v", testCase.Expected)
				t.Logf("Output: %v", out)
				t.Fail()
			}
		})
		if !pass {
			t.Fail()
		}
	}
}
