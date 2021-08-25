package admin

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/hiwjd/quick/support/sqlex"
	"golang.org/x/crypto/bcrypt"
)

// Admin 是管理员
type Admin struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Account   string    `gorm:"type:varchar(30);uniqueIndex;not null" json:"account"`
	Password  string    `gorm:"type:varchar(150);not null" json:"-"`
	Name      string    `gorm:"type:varchar(30)" json:"name"`
	Mobile    string    `gorm:"type:varchar(20)" json:"mobile"`
	Code      string    `gorm:"type:varchar(30);not null;uniqueIndex" json:"code"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CheckPassword 检查传入的密码是否本实例的密码
func (o Admin) CheckPassword(pass string) bool {
	return bcrypt.CompareHashAndPassword([]byte(o.Password), []byte(pass)) == nil
}

// SetPassword 设置密码
func (o *Admin) SetPassword(pass string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	o.Password = string(hash)
	return nil
}

func (o *Admin) Desensitive() {
	o.Password = ""
}

// AdminRole 是管理员的角色
type AdminRole struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	AdminID uint `gorm:"index" json:"adminID"`
	RoleID  uint `gorm:"index" json:"roleID"`
}

// Role 是角色
type Role struct {
	ID        uint      `gorm:"primaryKey" json:"id"`                   // 角色ID
	Group     string    `gorm:"type:varchar(30);not null" json:"group"` // 所属组
	Key       string    `gorm:"type:varchar(30);not null" json:"key"`   // 角色特性
	Name      string    `gorm:"type:varchar(30);not null" json:"name"`  // 角色名称
	DataPerms DataPerms `gorm:"type:varchar(500)" json:"dataPerm"`      // 数据权限
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Roles 角色组 用来使多角色的一些判断更方便点
type Roles []*Role

// ContainsKey 查看角色是否包含某个key
func (rs Roles) ContainsKey(key string) bool {
	for _, r := range rs {
		if r.Key == key {
			return true
		}
	}
	return false
}

// BelongGroup 判断角色是否不是归属某个组 只要有一个角色有即true
func (rs Roles) BelongGroup(group string) bool {
	for _, r := range rs {
		if r.Group == group {
			return true
		}
	}
	return false
}

// IsCFDGroup 返回是不是净菜角色
func (rs Roles) IsCFDGroup() bool {
	return rs.BelongGroup("cleanfood")
}

// RoleMenu 是角色的菜单
type RoleMenu struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	RoleID uint `gorm:"index" json:"roleID"`
	MenuID uint `gorm:"index" json:"menuID"`
}

// Menu 是菜单
type Menu struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	ParentID  uint             `json:"parentID"`
	Name      string           `gorm:"type:varchar(30);not null" json:"name"`
	Type      string           `gorm:"type:varchar(10)" json:"type"`
	URL       string           `gorm:"type:varchar(50);not null" json:"url"`
	Sort      int              `json:"sort"`
	APIList   sqlex.StringList `gorm:"type:varchar(2000)" json:"apiList"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

// API 是接口
type API struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Name   string `gorm:"type:varchar(40)" json:"name"`            // 接口名
	Method string `gorm:"type:varchar(20);not null" json:"method"` // 接口方式
	Path   string `gorm:"type:varchar(100);not null" json:"path"`  // 接口路径
}

// WxAdmin 是微信openid映射的管理员
type WxAdmin struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Openid    string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"openid"`
	AdminID   uint      `gorm:"not null" json:"adminID"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// APIData 是简化API
type APIData struct {
	Method string
	Path   string
}

// APIDataList 是简化API列表
type APIDataList []APIData

// Value 实现driver.Valuer
func (r *APIDataList) Value() (driver.Value, error) {
	bs, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return string(bs), nil
}

// Scan 实现 driver.Scaner
func (r *APIDataList) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), r)
	case []uint8:
		return json.Unmarshal(v, r)
	default:
		return fmt.Errorf("unknown type in APIDataList.Scan")
	}
}

// RoleData 是角色
type RoleData struct {
	ID   uint
	Key  string
	Name string
}

// RoleDataList 是角色列表
type RoleDataList []RoleData

// Value 实现driver.Valuer
func (r *RoleDataList) Value() (driver.Value, error) {
	bs, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return string(bs), nil
}

// Scan 实现 driver.Scaner
func (r *RoleDataList) Scan(value interface{}) error {
	if v, ok := value.(string); ok {
		return json.Unmarshal([]byte(v), r)
	}

	return fmt.Errorf("unknown type in RoleDataList.Scan")
}

// MenuData 是菜单
type MenuData struct {
	ID       uint
	ParentID uint
	Name     string
	Type     string
}

// MenuDataList 是菜单列表
type MenuDataList []MenuData

// Value 实现driver.Valuer
func (r *MenuDataList) Value() (driver.Value, error) {
	bs, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return string(bs), nil
}

// Scan 实现 driver.Scaner
func (r *MenuDataList) Scan(value interface{}) error {
	if v, ok := value.(string); ok {
		return json.Unmarshal([]byte(v), r)
	}

	return fmt.Errorf("unknown type in MenuDataList.Scan")
}

// Action 表示有权限的动作，比如用户列表界面上的“导出”按钮
type Action struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// MenuNode 是菜单树
type MenuNode struct {
	ID       uint        `json:"id"`
	ParentID uint        `json:"parentID"`
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	URL      string      `json:"url"`
	Children []*MenuNode `json:"children"`
	Actions  []Action    `json:"actions"`
}

func (mn MenuNode) isAction() bool {
	return mn.Type == "action"
}

func (mn MenuNode) asAction() Action {
	return Action{
		ID:   mn.URL,
		Name: mn.Name,
	}
}

// MenuList 是菜单列表
type MenuList []*Menu

// AsNode 把菜单列表转成菜单树
func (ml MenuList) AsNode() *MenuNode {
	sort.Sort(ml)

	mapping := make(map[uint]*MenuNode)
	noParent := make(map[uint][]*MenuNode)

	root := new(MenuNode)
	for _, m := range ml {
		mn := &MenuNode{
			ID:       m.ID,
			ParentID: m.ParentID,
			Name:     m.Name,
			Type:     m.Type,
			URL:      m.URL,
			Children: make([]*MenuNode, 0),
			Actions:  make([]Action, 0),
		}
		if p, ok := noParent[m.ID]; ok {
			for _, pn := range p {
				if pn.isAction() {
					mn.Actions = append(mn.Actions, pn.asAction())
				} else {
					mn.Children = append(mn.Children, pn)
				}
			}
			delete(noParent, m.ID)
		}
		mapping[m.ID] = mn

		if m.ParentID == 0 {
			root.Children = append(root.Children, mn)
		} else {
			if parent, ok := mapping[m.ParentID]; ok {
				if mn.isAction() {
					parent.Actions = append(parent.Actions, mn.asAction())
				} else {
					parent.Children = append(parent.Children, mn)
				}
			} else {
				noParent[m.ParentID] = append(noParent[m.ParentID], mn)
			}
		}
	}

	return root
}

func (ml MenuList) Len() int { return len(ml) }
func (ml MenuList) Less(i, j int) bool {
	if ml[i].ParentID == ml[j].ParentID {
		if ml[i].Sort == ml[j].Sort {
			return ml[i].ID < ml[j].ID
		}
		return ml[i].Sort < ml[j].Sort
	}
	return ml[i].ParentID < ml[j].ParentID
}
func (ml MenuList) Swap(i, j int) { ml[i], ml[j] = ml[j], ml[i] }
