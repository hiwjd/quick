package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMenuNode(t *testing.T) {
	menuList := []Menu{
		{ID: 1, ParentID: 4, Name: "菜单1-1", Type: "", Sort: 0},
		{ID: 2, ParentID: 5, Name: "菜单2-1", Type: "", Sort: 1},
		{ID: 3, ParentID: 5, Name: "菜单2-2", Type: "", Sort: 0},
		{ID: 4, ParentID: 0, Name: "菜单1", Type: "", Sort: 1},
		{ID: 5, ParentID: 0, Name: "菜单2", Type: "", Sort: 0},
		{ID: 6, ParentID: 2, Name: "菜单2-1-1", Type: "", Sort: 0},
		{ID: 7, ParentID: 2, Name: "菜单2-1-2", Type: "", Sort: 0},
		{ID: 8, ParentID: 1, Name: "按钮1", Type: "action", Sort: 0, URL: "btn1"},
		{ID: 9, ParentID: 1, Name: "按钮2", Type: "action", Sort: 0, URL: "btn2"},
		{ID: 10, ParentID: 6, Name: "按钮3", Type: "action", Sort: 0, URL: "btn3"},
		{ID: 11, ParentID: 7, Name: "按钮4", Type: "action", Sort: 0, URL: "btn4"},
	}
	ml := MenuList(menuList)
	root := ml.AsNode()

	expectRoot := MenuNode{
		ID:       0,
		ParentID: 0,
		Name:     "",
		Type:     "",
		Children: []*MenuNode{
			{ID: 5, ParentID: 0, Name: "菜单2", Type: "", Children: []*MenuNode{
				{ID: 3, ParentID: 5, Name: "菜单2-2", Type: "", Children: []*MenuNode{}, Actions: []*Action{}},
				{ID: 2, ParentID: 5, Name: "菜单2-1", Type: "", Children: []*MenuNode{
					{ID: 6, ParentID: 2, Name: "菜单2-1-1", Type: "", Children: []*MenuNode{}, Actions: []*Action{
						{ID: "btn3", Name: "按钮3"},
					}},
					{ID: 7, ParentID: 2, Name: "菜单2-1-2", Type: "", Children: []*MenuNode{}, Actions: []*Action{
						{ID: "btn4", Name: "按钮4"},
					}},
				}, Actions: []*Action{}},
			}, Actions: []*Action{}},
			{ID: 4, ParentID: 0, Name: "菜单1", Type: "", Children: []*MenuNode{
				{ID: 1, ParentID: 4, Name: "菜单1-1", Type: "", Children: []*MenuNode{}, Actions: []*Action{
					{ID: "btn1", Name: "按钮1"},
					{ID: "btn2", Name: "按钮2"},
				}},
			}, Actions: []*Action{}},
		},
	}

	assert.EqualValues(t, expectRoot, root)
}
