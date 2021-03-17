package systemtree

import (
	"encoding/json"
	"fmt"

	"log"
	"math"
	"strings"
)

//=============系统节点信息============
type System struct {
	SysID       uint32      `json:"sys_id"`
	ParentSysID uint32      `json:"parent_sys_id"`
	UsrData     interface{} `json:"usr_data"`
}

//nolint:gosimple
func (s *System) isValid() bool {
	if s.SysID == s.ParentSysID {
		return false
	}
	return true
}

func (s *System) isEqualTo(other *System) bool {
	if s.SysID != other.SysID || s.ParentSysID != other.ParentSysID || s.UsrData != other.UsrData {
		return false
	}
	return true
}

func (s *System) Marshal() ([]byte, error) {
	return json.Marshal(s)
}

func (s *System) Unmarshal(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *System) String() string {
	return fmt.Sprintf("{%d,%d,%v}", s.SysID, s.ParentSysID, s.UsrData)
}

func NewSystemFromString(from string) *System {
	s := &System{}
	err := s.Unmarshal([]byte(from))
	if err != nil {
		return nil
	}
	return s
}

//=============系统节点树============
type SystemTree struct {
	Systems map[uint32]*System            // {SysID: *System}
	Tree    map[uint32]map[uint32]*System //{ ParentSysID: {SysID: *System} }
}

func (st *SystemTree) String() string {
	return fmt.Sprintf("systems:%v tree:%v", st.Systems, st.Tree)
}

func (st *SystemTree) buildTree() error {
	for id, s := range st.Systems {
		if st.Tree[s.ParentSysID] == nil {
			st.Tree[s.ParentSysID] = make(map[uint32]*System)
		}
		st.Tree[s.ParentSysID][id] = s
	}
	return st.checkCircle()
}

//环检测O(n)
func (st *SystemTree) checkCircle() error {
	done := make(map[uint32]bool) //已完成节点记录
	for id, s := range st.Systems {
		if done[id] {
			continue
		}
		if !s.isValid() { //自身循环检测
			return fmt.Errorf("invalid system node:%v", s)
		}
		history := make(map[uint32]bool)
		history[id] = true
		pid := s.ParentSysID
		if done[pid] {
			done[id] = true
			continue
		}
		//向上查找
		history[pid] = true
		parent := st.Systems[pid]
		for parent != nil {
			pid = parent.ParentSysID
			if done[pid] {
				break
			}
			if history[pid] { //出现环
				return fmt.Errorf("system tree %s has circle sysid:%d, pid:%d", st, id, pid)
			}
			history[pid] = true
			parent = st.Systems[pid]
		}
		//确认无环才加入
		for hID := range history {
			done[hID] = true
		}
	}
	return nil
}

func (st *SystemTree) AddSystem(s *System) error {
	old := st.Systems[s.SysID]
	if old != nil {
		if old.isEqualTo(s) {
			log.Fatalf("re-add system %s", s)
			return nil
		}
		st.DelSystem(s.SysID) //nolint:errcheck
	}
	//加入时,检测环
	pid := s.ParentSysID
	parent := st.Systems[pid]
	history := make(map[uint32]bool)
	history[pid] = true
	for parent != nil {
		pid = parent.ParentSysID
		if pid == s.SysID || history[pid] { //出现环
			return fmt.Errorf("system tree %s has circle sysid:%d, pid:%d", st, s.SysID, pid)
		}
		history[pid] = true
		parent = st.Systems[pid]
	}

	st.Systems[s.SysID] = s
	if st.Tree[s.ParentSysID] == nil {
		st.Tree[s.ParentSysID] = make(map[uint32]*System)
	}
	st.Tree[s.ParentSysID][s.SysID] = s
	return nil
}

//删除节点实体, 相当于从某节点的子节点中卸载自身, 其对应子树变为游离态
func (st *SystemTree) DelSystem(sysID uint32) (*System, error) {
	s := st.Systems[sysID]
	if s == nil {
		return nil, fmt.Errorf("re-delete system")
	}
	delete(st.Systems, sysID)
	delete(st.Tree[s.ParentSysID], sysID)
	return s, nil
}

func (st *SystemTree) GetSystem(sysID uint32) *System {
	return st.Systems[sysID]
}

//bfs search
func (st *SystemTree) getChilds(sysID uint32, maxFloor int) (map[uint32]*System, int) {
	if maxFloor <= 0 {
		maxFloor = math.MaxInt32
	}
	queue := make([]uint32, 1)
	queue[0] = sysID
	history := make(map[uint32]*System)
	curFloor := 0
	head, tail := 0, len(queue)
	for curFloor < maxFloor {
		for head < tail {
			parentID := queue[head]
			for cid, child := range st.Tree[parentID] {
				if cid == sysID || history[cid] != nil { //出现环...
					log.Fatalf("system tree %s has circle sysid:%d, pid:%d cid:%d", st, sysID, parentID, cid)
					return history, curFloor
				}
				queue = append(queue, cid)
				history[cid] = child
			}
			head++
		}
		tail = len(queue)
		if head < tail {
			curFloor++
		} else {
			break
		}
	}
	return history, curFloor
}

func (st *SystemTree) GetChildSystems(sysID uint32) map[uint32]*System {
	childs, _ := st.getChilds(sysID, 1)
	return childs
}

func (st *SystemTree) getNodeNum() int {
	return len(st.Systems)
}

//从根到节点的唯一路径长(根的深度为0)，-1表示节点不存在
func (st *SystemTree) getNodeDeep(sysID uint32) int {
	s := st.Systems[sysID]
	if s == nil {
		if st.Tree[sysID] != nil {
			return 0
		}
		return -1
	}
	history := make(map[uint32]bool)
	history[s.SysID] = true
	deep := 1
	parent := st.Systems[s.ParentSysID]
	for parent != nil {
		deep++
		if history[parent.SysID] {
			log.Fatalf("suffer dead circle when getNodeDeep node:%d parent:%d", sysID, parent.SysID)
			break
		}
		history[parent.SysID] = true
		parent = st.Systems[parent.ParentSysID]
	}
	return deep
}

//从节点到一片树叶的最长路径长(所有树叶的高度为0)，-1表示节点不存在
func (st *SystemTree) getNodeHeight(sysID uint32) int {
	if st.Systems[sysID] == nil && st.Tree[sysID] == nil {
		return -1
	}
	_, height := st.getChilds(sysID, 0)
	return height
}

func (st *SystemTree) describeTree() {
	trees := []uint32{}
	for pid := range st.Tree {
		if st.Systems[pid] == nil {
			trees = append(trees, pid)
		}
	}
	log.Printf("========Tree Describe:=====\n")
	for _, rootID := range trees {
		queue := make([]uint32, len(st.Tree[rootID]))
		queue = queue[:0]
		history := make(map[uint32]*System)
		for cid, child := range st.Tree[rootID] {
			queue = append(queue, cid)
			history[cid] = child
		}
		log.Printf("%d(root):%v\n", rootID, queue)
		head, tail := 0, len(queue)
		floor := 1
		for head < tail {
			log.Printf(strings.Repeat("___ ", floor))
			for head < tail {
				parentID := queue[head]
				subque := make([]uint32, len(st.Tree[parentID]))
				subque = subque[:0]
				for cid, child := range st.Tree[parentID] {
					if cid == rootID || history[cid] != nil { //出现环...
						log.Printf("suffer dead circle root(%d) cid(%d) systems:%v tree:%v!!", rootID, cid, st.Systems, st.Tree)
						return
					}
					queue = append(queue, cid)
					subque = append(subque, cid)
					history[cid] = child
				}
				log.Printf("%d:%v ", parentID, subque)
				head++
			}
			floor++
			log.Printf("\n")
			tail = len(queue)
		}
	}
}
