package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMenuNode(t *testing.T) {
	menuList := []*Menu{
		{ID: 1, ParentID: 4, Name: "菜单1-1", Type: "", Sort: 0},
		{ID: 2, ParentID: 5, Name: "菜单2-1", Type: "", Sort: 1},
		{ID: 3, ParentID: 5, Name: "菜单2-2", Type: "", Sort: 0},
		{ID: 4, ParentID: 0, Name: "菜单1", Type: "", Sort: 1},
		{ID: 5, ParentID: 0, Name: "菜单2", Type: "", Sort: 0},
		{ID: 6, ParentID: 2, Name: "菜单2-1-1", Type: "", Sort: 0},
		{ID: 7, ParentID: 2, Name: "菜单2-1-2", Type: "", Sort: 0},
	}
	ml := MenuList(menuList)
	root := ml.AsNode()

	expectRoot := &MenuNode{
		ID:       0,
		ParentID: 0,
		Name:     "",
		Type:     "",
		Children: []*MenuNode{
			{ID: 5, ParentID: 0, Name: "菜单2", Type: "", Children: []*MenuNode{
				{ID: 3, ParentID: 5, Name: "菜单2-2", Type: "", Children: []*MenuNode{}},
				{ID: 2, ParentID: 5, Name: "菜单2-1", Type: "", Children: []*MenuNode{
					{ID: 6, ParentID: 2, Name: "菜单2-1-1", Type: "", Children: []*MenuNode{}},
					{ID: 7, ParentID: 2, Name: "菜单2-1-2", Type: "", Children: []*MenuNode{}},
				}},
			}},
			{ID: 4, ParentID: 0, Name: "菜单1", Type: "", Children: []*MenuNode{
				{ID: 1, ParentID: 4, Name: "菜单1-1", Type: "", Children: []*MenuNode{}},
			}},
		},
	}

	assert.EqualValues(t, expectRoot, root)
}
