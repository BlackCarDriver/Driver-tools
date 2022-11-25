package bossExport

import (
	"fmt"
	"git.yy.com/server/jiaoyou/go_projects/api/basedao/webdbdao"
	"git.yy.com/server/jiaoyou/go_projects/api/common/mgopoolclient"
	"git.yy.com/server/jiaoyou/go_projects/api/common/util"
	"github.com/BlackCarDriver/GoProject-api/color"
	"github.com/globalsign/mgo/bson"
	"os"
	"time"
)

// 导出Boss页面请求记录 (数据量大一个月约15万行,注意查询范围)
func exportBossHistory() {
	color.Yellow("本工具按照时间范围查询并导出Boss页面请求记录, 一个月约15万行,注意查询范围！")
	color.Yellow("请确保 startTime, endTime 填写无误")
	initUsualConfig()
	color.Yellow("准备就绪,任务即将开始")
	wait()
	startTime, endTime := usualConfig.StartTime, usualConfig.EndTime
	history, err := getHistoryByTimeRange(startTime.Unix(), endTime.Unix())

	color.HiBlack("获取数据结果: err=%v startTime=%s endTime=%s n_history=%d", err, startTime.Format(dateFormat), endTime.Format(dateFormat), len(history))

	var table [][]string
	header := []string{"用户UID", "时间", "请求方法", "URL", "参数", "请求主体"}
	table = append(table, header)
	for _, item := range history {
		row := []string{
			fmt.Sprint(item.UID),
			time.Unix(item.Timestamp, 0).Format(timeFormat),
			item.Method,
			item.URL,
			item.Params,
			item.Body,
		}
		table = append(table, row)
	}
	color.HiBlack("create table success: n_row=%d", len(table))
	xlsx, err := util.ParseFormToXlsx(table)
	if err != nil {
		color.Red("create xlsx fail: err=%v", err)
		return
	}
	fileName := fmt.Sprintf("./boss请求记录_%s_%s.xlsx", startTime.Format(dateFormat), endTime.Format(dateFormat))
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("open file fail: err=%v", err)
		return
	}
	file.Write(xlsx)
	file.Close()
	color.Green("任务完成~")
}

// 统计接口请求频率 和 用户访问次数
func exportBossUrlStat() {
	color.Yellow("本工具按照时间范围查询并导 接口请求频率,用户访问次数统计 报表")
	color.Yellow("请确保 startTime, endTime 填写无误")
	initUsualConfig()
	color.Yellow("准备就绪,任务即将开始")
	wait()

	startTime, endTime := usualConfig.StartTime, usualConfig.EndTime
	history, err := getHistoryByTimeRange(startTime.Unix(), endTime.Unix())
	color.HiBlack("err=%v startTime=%s endTime=%s n_history=%d", err, startTime.Format(dateFormat), endTime.Format(dateFormat), len(history))

	urlToCount := map[string]int64{}
	uidToCount := map[int64]int{}
	var uidList []int64

	for _, item := range history {
		if uidToCount[item.UID] == 0 {
			uidList = append(uidList, item.UID)
		}
		urlToCount[item.URL]++
		uidToCount[item.UID]++
	}

	var table [][]string
	table = append(table, []string{"", ""})
	header := []string{"URL", "请求数量"}
	table = append(table, header)
	for uri, count := range urlToCount {
		row := []string{
			uri,
			fmt.Sprint(count),
		}
		table = append(table, row)
	}

	nameMap, err := webdbdao.BatchGetUserPassport(uidList)
	if err != nil || nameMap == nil {
		color.Red("get passport fail: err=%v uidList=%v", err, uidList)
		nameMap = make(map[int64]string)
	}
	color.HiBlack("get passport: n_uid=%d n_passport=%d", len(uidList), len(nameMap))

	header = []string{"用户UID", "passport", "请求次数"}
	table = append(table, header)
	for uid, count := range uidToCount {
		row := []string{
			fmt.Sprint(uid),
			nameMap[uid],
			fmt.Sprint(count),
		}
		table = append(table, row)
	}
	color.HiBlack("create table success: n_row=%d", len(table))

	xlsx, err := util.ParseFormToXlsx(table)
	if err != nil {
		color.Red("create xlsx fail: err=%v", err)
		return
	}
	fileName := fmt.Sprintf("./boss请求数据统计_%s_%s.xlsx", startTime.Format(dateFormat), endTime.Format(dateFormat))
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("open file fail: err=%v", err)
		return
	}
	file.Write(xlsx)
	file.Close()
	color.HiGreen("任务完成~")
}

// 导出更新记录
func exportBossOpLog() {
	color.Yellow("本工具按照时间范围查询并导出 boss操作记录 报表")
	color.Yellow("请确保 startTime, endTime 填写无误")
	initUsualConfig()
	color.Yellow("任务即将开始")
	wait()

	startTime, endTime := usualConfig.StartTime, usualConfig.EndTime
	opList, err := getOpLogByTimeRange(startTime.Unix(), endTime.Unix())
	color.HiBlack("查询数据完成: err=%v startTime=%s endTime=%s n_history=%d", err, startTime.Format(dateFormat), endTime.Format(dateFormat), len(opList))

	var table [][]string
	opTypeMap := map[int]string{1: "ADD", 2: "MOD", 3: "DEL"}
	header := []string{"相关UID", "时间", "操作描述", "URI", "操作类型", "更新前", "更新后"}
	table = append(table, header)
	for _, item := range opList {
		row := []string{
			fmt.Sprint(item.Uid),
			time.Unix(item.Timestamp, 0).Format(timeFormat),
			item.Name,
			item.Uri,
			opTypeMap[item.OperateType],
			item.OldValue,
			item.NewValue,
		}
		table = append(table, row)
	}
	color.HiBlack("create table success: n_row=%d", len(table))
	xlsx, err := util.ParseFormToXlsx(table)
	if err != nil {
		color.HiBlack("create xlsx fail: err=%v", err)
		return
	}
	fileName := fmt.Sprintf("./数据更新历史_%s_%s.xlsx", startTime.Format(dateFormat), endTime.Format(dateFormat))
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("open file fail: err=%v", err)
		return
	}
	file.Write(xlsx)
	file.Close()
	color.HiGreen("任务完成")
}

// 导出当前的页面权限列表
func exportBossAuthItem() {
	color.Yellow("本工具按照时间范围查询并导出 当前时刻的BOSS页面权限 报表")
	color.Yellow("无需填写配置")
	color.Yellow("任务即将开始")
	wait()

	autoInfo, err := getBossAuthItem()
	color.HiBlack("err=%v n_autoInfo=%d", err, len(autoInfo))

	uidSet := map[int64]int{}
	var allUID []int64
	var table [][]string

	for _, item := range autoInfo {
		uidSet[item.UID]++
		if uidSet[item.UID] == 1 {
			allUID = append(allUID, item.UID)
		}
	}
	uidToPassport, err := webdbdao.BatchGetUserPassport(allUID)
	color.HiBlack("get passport: n_user=%d n_passport=%d err=%v", len(allUID), len(uidToPassport), err)
	for i, item := range autoInfo {
		autoInfo[i].Nick = uidToPassport[item.UID]
	}

	header := []string{"UID", "昵称", "URI", "页面", "权限申请时间", "权限过期时间", "权限审批人", "相关审批流ID"}
	table = append(table, header)
	for _, item := range autoInfo {

		row := []string{
			fmt.Sprint(item.UID),
			item.Nick,
			item.URI,
			fmt.Sprintf("%s%s", item.Tab, item.SubTab),
			time.Unix(item.Timestamp, 0).Format(timeFormat),
			item.ExpireAt.Format(timeFormat),
			item.Approval,
			item.PID,
		}
		table = append(table, row)
	}
	color.HiBlack("create table success: n_row=%d", len(table))
	xlsx, err := util.ParseFormToXlsx(table)
	if err != nil {
		color.Red("create xlsx fail: err=%v", err)
		return
	}
	fileName := fmt.Sprintf("./boss页面权限_%s.xlsx", time.Now().Format(dateFormat))
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("open file fail: err=%v", err)
		return
	}
	file.Write(xlsx)
	file.Close()
	color.HiGreen("任务完成")
}

// 审批流程数据导出 (注意时间参数)
func exportApprovalFlow() {
	color.Yellow("请确保 startTime, endTime 填写无误")
	initUsualConfig()
	color.Yellow("任务即将开始")
	wait()
	startTime, endTime := usualConfig.StartTime, usualConfig.EndTime

	ruleMap, err := getRuleMap()
	color.HiBlack("get ruleMap: err=%v n_ruleMap=%d", err, len(ruleMap))

	proposer, err := getProposerWithTimeRange(startTime.Unix(), endTime.Unix())
	color.HiBlack("getProposer: err=%v startTime=%s endTime=%s n_proposer=%d", err, startTime.Format(dateFormat), endTime.Format(dateFormat), len(proposer))
	approvalMap, err := getApproval()
	color.HiBlack("get approval: err=%v n_approvalMap=%d", err, len(approvalMap))

	var table [][]string
	header := []string{"规则ID", "规则名称", "流程ID", "申请人", "申请时间", fmt.Sprintf("当前状态 (%s)", time.Now().Format(timeFormat)), "审批日志", "通知文案", "附加信息"}
	table = append(table, header)
	for _, item := range proposer {
		var approvalLot = fmt.Sprintf("%s, %s(uid=%d)创建了审批流, 理由='%s', nid=%d, 过期时间=%s, 流程id=%s",
			time.Unix(item.Timestamp, 0).Format(timeFormat),
			item.Creator.Passport,
			item.Creator.UID,
			item.Reason,
			item.NID,
			time.Unix(item.Expire, 0).Format(timeFormat),
			item.PID,
		)
		for _, apr := range approvalMap[item.PID] {
			approvalLot += fmt.Sprintf("\n%s,%s(uid=%d)执行审批,操作='%s',理由='%s',审批ID=%s",
				time.Unix(apr.Timestamp, 0).Format(timeFormat),
				apr.Creator.Passport,
				apr.Creator.UID,
				apr.Result,
				apr.Reason,
				apr.ID)
		}

		row := []string{
			fmt.Sprint(ruleMap[item.RID].RID),
			ruleMap[item.RID].Name,
			item.PID,
			fmt.Sprintf("%d (%s)", item.Creator.UID, item.Creator.Passport),
			time.Unix(item.Timestamp, 0).Format(timeFormat),
			item.Result,
			approvalLot,
			item.WCText,
			item.Text,
		}
		table = append(table, row)
	}
	color.HiBlack("create table success: n_row=%d", len(table))
	xlsx, err := util.ParseFormToXlsx(table)
	if err != nil {
		color.Red("create xlsx fail: err=%v", err)
		return
	}
	fileName := fmt.Sprintf("./审批流记录_%s_%s.xlsx", startTime.Format(dateFormat), endTime.Format(dateFormat))
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("open file fail: err=%v", err)
		return
	}
	file.Write(xlsx)
	file.Close()
	color.HiGreen("任务完成")
}

// ============ 查询数据 ===================

// 获取页面访问历史数据 (Boss页面查看或调用数据)
func getHistoryByTimeRange(start, end int64) (history []History, err error) {
	session, err := mgopoolclient.GetMgoSession()
	if err != nil {
		color.Red("Get Session err: ", err)
		return
	}
	defer session.Close()
	client := session.DB("fts_boss").C("record_history")
	iter := client.Find(bson.M{"timestamp": bson.M{"$gte": start, "$lte": end}}).Iter()

	var item History
	for iter.Next(&item) {
		history = append(history, item)
	}
	if err = iter.Err(); err != nil {
		color.Red("get history fail: err=%v", err)
		return
	}
	color.HiBlack("get all history success: start=%d end=%d n_history=%d", start, end, len(history))
	return
}

// 获取数据更新历史
func getOpLogByTimeRange(start, end int64) (history []OperateRecord, err error) {
	session, err := mgopoolclient.GetMgoSession()
	if err != nil {
		color.Red("Get Session err: ", err)
		return
	}
	defer session.Close()
	client := session.DB("fts_boss").C("OperateRecordInfo")
	iter := client.Find(bson.M{"timestamp": bson.M{"$gte": start, "$lte": end}}).Iter()

	var item OperateRecord
	for iter.Next(&item) {
		history = append(history, item)
	}
	if err = iter.Err(); err != nil {
		color.Red("get opLog fail: err=%v", err)
		return
	}
	color.HiBlack("get all opLog success: start=%d end=%d n_history=%d", start, end, len(history))
	return
}

// 查询最新的Boss权限状况
func getBossAuthItem() (authInfo []AuthItem, err error) {
	session, err := mgopoolclient.GetMgoSession()
	if err != nil {
		color.Red("Get Session err: ", err)
		return
	}
	defer session.Close()
	client := session.DB("fts_boss").C("boss_auth_item")
	iter := client.Find(bson.M{}).Iter()

	var item AuthItem
	for iter.Next(&item) {
		authInfo = append(authInfo, item)
	}
	if err = iter.Err(); err != nil {
		color.Red("get authInfo fail: err=%v", err)
		return
	}
	color.HiBlack("get all authInfo success: n_history=%d", len(authInfo))
	return
}

// 获取所有审批规则
func getRuleMap() (ruleMap map[int64]Rule, err error) {
	session, err := mgopoolclient.GetMgoSession()
	if err != nil {
		color.Red("Get Session err: ", err)
		return
	}
	defer session.Close()
	client := session.DB("fts_boss").C("table_approval_rule")
	iter := client.Find(bson.M{}).Iter()

	var tmp Rule
	ruleMap = make(map[int64]Rule)
	for iter.Next(&tmp) {
		ruleMap[tmp.RID] = tmp
	}
	if err = iter.Err(); err != nil {
		color.Red("get rule fail: err=%v", err)
		return
	}
	color.HiBlack("get all rule success: n_rule=%d", len(ruleMap))
	return
}

// 按照时间范围获取审批流程
func getProposerWithTimeRange(start, end int64) (list []Proposer, err error) {
	session, err := mgopoolclient.GetMgoSession()
	if err != nil {
		color.Red("Get Session err: ", err)
		return
	}
	defer session.Close()
	client := session.DB("fts_boss").C("table_approval_proposer")
	iter := client.Find(bson.M{"timestamp": bson.M{"$gte": start, "$lte": end}}).Sort("rid", "-timestamp").Iter()

	var tmp Proposer
	for iter.Next(&tmp) {
		list = append(list, tmp)
	}
	if err = iter.Err(); err != nil {
		color.Red("get proposer fail: err=%v", err)
		return
	}
	color.HiBlack("get all proposer success: start=%d end=%d n_proposer=%d", start, end, len(list))
	return
}

// PID->审批记录
func getApproval() (pidToLog map[string][]Approval, err error) {
	session, err := mgopoolclient.GetMgoSession()
	if err != nil {
		color.Red("Get Session err: ", err)
		return
	}
	defer session.Close()
	client := session.DB("fts_boss").C("table_approval_approval")
	iter := client.Find(bson.M{}).Sort("pid", "-timestamp").Iter()

	var item Approval
	pidToLog = make(map[string][]Approval)
	for iter.Next(&item) {
		pidToLog[item.PID] = append(pidToLog[item.PID], item)
	}
	if err = iter.Err(); err != nil {
		color.Red("get approval fail: err=%v", err)
		return
	}
	color.HiBlack("get all approval success: approval=%d", len(pidToLog))
	return
}
