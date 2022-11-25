package bossExport

import "time"

// History Boss请求记录
type History struct {
	UID       int64  `json:"uid" bson:"uid"`
	Timestamp int64  `json:"timestamp" bson:"timestamp"`
	URL       string `json:"url" bson:"url"`
	Method    string `json:"method" bson:"method"`
	Params    string `json:"params" bson:"params"`
	Body      string `json:"body" bson:"body"`
}

// OperateRecord 数据更新记录
type OperateRecord struct {
	Uid         int64  `bson:"uid"`
	OperateType int    `bson:"operateType"`
	Uri         string `bson:"uri"`
	Name        string `bson:"name"`
	OldValue    string `bson:"oldValue"`
	NewValue    string `bson:"newValue"`
	Timestamp   int64  `bson:"timestamp"`
}

// AuthItem  Boss权限列表
type AuthItem struct {
	ID        string    `bson:"_id"              json:"_id"`       // pid:uri-新版版 or legacy:pid:uri-旧的权限
	PID       string    `bson:"pid"              json:"pid"`       // proposer id
	UID       int64     `bson:"uid"              json:"uid"`       // 申请人
	Passport  string    `bson:"-"                json:"passport"`  // 通行证
	Nick      string    `bson:"-"                json:"nick"`      // 昵称
	URI       string    `bson:"uri"              json:"uri"`       // index({uid: 1, uri: 1}, {unique: 1})
	Expire    int64     `bson:"expire"           json:"expire"`    // 过期时间
	ExpireAt  time.Time `bson:"expireAt"         json:"expireAt"`  // index({expireAt: 1}, {expireAfterSeconds: 0}) 过期后自动删除
	Tab       string    `bson:"tab"              json:"tab"`       // 菜单
	SubTab    string    `bson:"subTab"           json:"subTab"`    // 子菜单
	Approval  string    `bson:"approval"         json:"approval"`  // 审批人passport
	Timestamp int64     `bson:"timestamp"        json:"timestamp"` // 审批时间
	Reason    string    `bson:"-"                json:"reason"`    // 申请理由
}

type User struct {
	YY       int64  `json:"yy"       bson:"yy"`       // 审批人YY号
	UID      int64  `json:"uid"      bson:"uid"`      // 审批人UID
	Passport string `json:"passport" bson:"passport"` // 审批人通信证
}

// Rule 审批规则
type Rule struct {
	RID              int64    `json:"rid"          bson:"_id"`          // 规则ID
	Name             string   `json:"name"         bson:"name"`         // 规则名称
	Template         string   `json:"template"     bson:"template"`     // OA工作流模板ID 默认不走OA流程
	Secret           string   `json:"secret"       bson:"secret"`       // 秘钥
	Proposers        []User   `json:"proposers"    bson:"proposers"`    // 申请人 限制申请人 默认不限制
	Approvals        []User   `json:"approvals"    bson:"approvals"`    // 审批人 简易审批流程 可以配置多个审批人
	Secondary        []User   `json:"secondary"    bson:"secondary"`    // 二级审批人 默认为空
	AUIDList         []int64  `json:"aUIDList"     bson:"aUIDList"`     // 审批人UID 用于json post
	PUIDList         []int64  `json:"pUIDList"     bson:"pUIDList"`     // 申请人UID
	SUIDList         []int64  `json:"sUIDList"     bson:"sUIDList"`     // 二级审批人UID
	WCApr            int64    `json:"wcApr"        bson:"wcApr"`        // wChat approval 1-启用 0-不启用 -1-完全禁用
	TextMap          string   `json:"textMap"      bson:"textMap"`      // 文案中文标识
	Link             string   `json:"link"         bson:"link"`         // 跳转
	Version          int64    `json:"version"      bson:"version"`      // 版本号
	ShowMode         int64    `json:"showMode"     bson:"showMode"`     // 显示模式 0-Descriptions 1-带历史记录的Table模式 2-打包审批模式
	InnerMode        int64    `json:"innerMode"    bson:"innerMode"`    // 是否将解析内容外显
	AutoApr          int64    `json:"autoApr" bson:"autoApr"`           // 测试环境时,直接通过或驳回审批,不产生任何通知 0-不自动审批, 1-直接通过, 2-直接驳回
	Pictures         []string `json:"pictures"     bson:"pictures"`     // 需要解析成图片的字段
	PrivatePicFields []string `json:"privatePics"  bson:"privatePics"`  // 使用bs2dl私域的图片
	Files            []string `json:"files"        bson:"files"`        // 需要解析成文档的文件
	PrivateFiles     []string `json:"privateFiles" bson:"privateFiles"` // 使用bs2dl私域的文件
	InnerObject      []string `json:"innerObject"  bson:"innerObject"`  // 嵌套结构
}

// Proposer 发起审批
type Proposer struct {
	PID        string     `json:"pid"        bson:"_id"`                 // RID:SeqID 调用方确定,为空时将使用uuid.New()赋值
	RID        int64      `json:"rid"        bson:"rid"`                 // 规则ID
	NID        int64      `json:"nid"        bson:"nid"`                 // numeric pid	(由PID计算出的哈希值)
	Creator    User       `json:"creator"    bson:"creator"`             // 发起人
	Timestamp  int64      `json:"timestamp"  bson:"timestamp"`           // 发起时间 index({ timestamp: -1 })
	UpdateSecs int64      `json:"updateSecs" bson:"updateSecs"`          // 更新时间
	Expire     int64      `json:"expire"     bson:"expire"`              // 权限过期时间
	Text       string     `json:"text"       bson:"text"`                // BOSS或OA展示的内容, 格式: JSON序列化后的Map[string]string, map中的key对应Rule.textMap中的key;
	WCText     string     `json:"wcText"     bson:"wcText"`              // 微信通知文案
	Append     string     `json:"append"     bson:"append"`              // 附加文案
	Callback   string     `json:"callback"   bson:"callback"`            // 流程完成回调
	Result     string     `json:"result"     bson:"result"`              // 审批结果 ["OnGoing", "Passed", "Rejected"]
	Progress   int64      `json:"progress"   bson:"progress"`            // 进度 0-第一阶段 1-第二阶段
	AprMode    string     `json:"aprMode"    bson:"aprMode"`             // 审批方式 wChat oa boss api auto
	Sign       string     `json:"sign"       bson:"sign"`                // 秘钥
	Rule       Rule       `json:"rule"       bson:"rule,omitempty"`      // $lookup and $unwind foreign_key
	Approvals  []Approval `json:"approvals"  bson:"approvals,omitempty"` // $lookup foreign_key
	Reason     string     `json:"reason"     bson:"reason"`
	SpecText   string     `json:"specText"   bson:"specText"` // 特殊说明 JSON字符串 主要用于多条审批打包成一条记录的情况 [{ index: 0, text: "xxx" }]
}

// Approval 审批记录 unique_index{ pid: 1, uid: 1 }
type Approval struct {
	ID        string   `json:"id"          bson:"_id"`                // pid:uid:进度
	PID       string   `json:"pid"         bson:"pid"`                // 流ID
	Creator   User     `json:"creator"     bson:"creator"`            // 审批人
	Timestamp int64    `json:"timestamp"   bson:"timestamp"`          // 审批时间 index({ timestamp: -1 })
	Result    string   `json:"result"      bson:"result"`             // 审批结果
	Reason    string   `json:"reason"      bson:"reason"`             // 审批说明
	SpecText  string   `json:"specText"    bson:"specText"`           // 特殊说明 JSON字符串 主要用于多条审批打包成一条记录的情况 [{ index: 0, text: "xxx" }]
	Proposer  Proposer `json:"proposer"    bson:"proposer,omitempty"` // 审批记录 $lookup and $unwind
	Rule      Rule     `json:"rule"        bson:"rule,omitempty"`     // 审批规则 $lookup and $unwind
}
