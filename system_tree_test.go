package systemtree

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystemTree(t *testing.T) {
	st := &SystemTree{
		Systems: make(map[uint32]*System),
		Tree:    make(map[uint32]map[uint32]*System),
	}
	tests := []System{
		{
			22,
			11,
			"aaa",
		},
		{
			55,
			11,
			"cccc",
		},
		{
			33,
			22,
			"bbb",
		},
		{
			222,
			111,
			"xxxx",
		},
		{
			333,
			222,
			"yyyy",
		},
		{
			44,
			22,
			"dddd",
		},
		{
			11,
			0,
			"ffff",
		},
	}
	for i := range tests {
		err := st.AddSystem(&tests[i])
		assert.Equal(t, nil, err)
	}
	err := st.AddSystem(&System{
		SysID:       0,
		ParentSysID: 11,
		UsrData:     "rpc",
	})
	//出现环
	assert.NotNil(t, err)

	//覆盖
	err = st.AddSystem(&System{
		SysID:       55,
		ParentSysID: 22,
		UsrData:     "rpc",
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(st.GetChildSystems(11)))
	assert.Equal(t, 3, len(st.GetChildSystems(22)))
	fmt.Printf("tree init:")
	st.describeTree()
	//删除
	st.DelSystem(33)
	st.DelSystem(222)
	assert.Equal(t, 2, len(st.GetChildSystems(22)))
	fmt.Printf("tree after delete node 33, 222:")
	st.describeTree()

	//reset
	st = &SystemTree{
		Systems: make(map[uint32]*System),
		Tree:    make(map[uint32]map[uint32]*System),
	}
	tests = []System{
		{
			22,
			11,
			"aaa",
		},
		{
			55,
			11,
			"cccc",
		},
		{
			33,
			22,
			"bbb",
		},
		{
			222,
			111,
			"xxxx",
		},
		{
			333,
			222,
			"yyyy",
		},
		{
			44,
			22,
			"dddd",
		},
		{
			11,
			33,
			"dddd",
		},
	}
	for i, v := range tests {
		st.Systems[v.SysID] = &tests[i]
	}
	err = st.buildTree()
	//出现环
	assert.NotNil(t, err)
}
