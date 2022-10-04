package common

// DriverToolFunction 通用功能接口定义
type DriverToolFunction interface {
	GetInfo() (name string, desc string) // 获取功能名称和描述
	Run() (retCmd string, err error)     // 启动入口, retCmd 返回下个功能的名称或返回 [turn\end\exit]
	Exit()                               // 关闭逻辑
}
