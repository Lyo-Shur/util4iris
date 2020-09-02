package api

import (
	"github.com/kataras/iris"
	"github.com/lyoshur/gutils"
	"github.com/lyoshur/mysqlutils"
	"log"
	"strconv"
	"strings"
	"time"
)

// 返回状态码以及数据
const Success = 0
const Fail = -1

const DataAnalysisFailMessage = "数据解析失败"
const ExecuteFailMessage = "执行出错"
const SuccessMessage = "成功"

const NullData = ""

// 简单控制器层
type SimpleApi struct{}

// 主请求体
// clientAlias 链接别名
// tableName 数据库名
// operation 操作命令
// ctx IRIS上下文
// dbHelpers 多数据库查询集合(使用IRIS注入)
func (simpleApi *SimpleApi) PostBy(tableName string, operation string, ctx iris.Context, factory mysqlutils.SessionFactory) string {
	// 分发请求
	unknown := "unknown"
	if !factory.DataBase.TableExist(tableName) {
		simpleApi.operationFactory(unknown)(factory, tableName, ctx)
	}
	return simpleApi.operationFactory(operation)(factory, tableName, ctx)
}

// 操作分发工厂
func (simpleApi *SimpleApi) operationFactory(operation string) func(factory mysqlutils.SessionFactory, tableName string, ctx iris.Context) string {
	switch operation {
	case "list":
		return simpleApi.list
	case "model":
		return simpleApi.model
	case "insert":
		return simpleApi.insert
	case "update":
		return simpleApi.update
	case "delete":
		return simpleApi.delete
	default:
		return simpleApi.unknown
	}
}

func (simpleApi *SimpleApi) list(factory mysqlutils.SessionFactory, tableName string, ctx iris.Context) string {
	// 解析参数
	values, err := simpleApi.Values(factory, tableName, ctx)
	if err != nil {
		log.Println(err)
		return (&gutils.CodeModeDTO{
			Code:    Fail,
			Message: DataAnalysisFailMessage,
			Data:    NullData,
		}).ToJson()
	}
	// 执行查询
	table, err := factory.Services[tableName].GetList(values)
	if err != nil {
		log.Println(err)
		return (&gutils.CodeModeDTO{
			Code:    Fail,
			Message: ExecuteFailMessage,
			Data:    NullData,
		}).ToJson()
	}
	count, err := factory.Services[tableName].GetCount(values)
	if err != nil {
		log.Println(err)
		return (&gutils.CodeModeDTO{
			Code:    Fail,
			Message: ExecuteFailMessage,
			Data:    NullData,
		}).ToJson()
	}
	// 组合外键
	tableMap := table.ToMap()
	keys := factory.DataBase.GetTable(tableName).Keys
	for i := 0; i < len(tableMap); i++ {
		for j := 0; j < len(keys); j++ {
			query := tableMap[i][gutils.ToBigHump(keys[j].ColumnName)]
			table, err := factory.Services[keys[j].RelyTable].GetModel(query)
			if err != nil {
				log.Println(err)
				return (&gutils.CodeModeDTO{
					Code:    Fail,
					Message: ExecuteFailMessage,
					Data:    NullData,
				}).ToJson()
			}
			tempMap := table.ToMap()
			if len(tempMap) >= 1 {
				tableMap[i][gutils.ToBigHump(keys[j].RelyTable)] = tempMap[0]
			}
		}
	}
	return (&gutils.CodeModeDTO{
		Code:    Success,
		Message: SuccessMessage,
		Data: map[string]interface{}{
			"List":  tableMap,
			"Count": count,
		},
	}).ToJson()
}

func (simpleApi *SimpleApi) model(factory mysqlutils.SessionFactory, tableName string, ctx iris.Context) string {
	// 解析参数
	values, err := simpleApi.Values(factory, tableName, ctx)
	if err != nil {
		log.Println(err)
		return (&gutils.CodeModeDTO{
			Code:    Fail,
			Message: DataAnalysisFailMessage,
			Data:    NullData,
		}).ToJson()
	}
	// 根据索引确定主键
	key := ""
	indexs := factory.DataBase.GetTable(tableName).Indexs
	for i := 0; i < len(indexs); i++ {
		index := indexs[i]
		if index.Name == "PRIMARY" {
			key = gutils.ToBigHump(index.ColumnName)
		}
	}
	// 判断主键是非存在
	v, ok := values[key]
	if !ok {
		return (&gutils.CodeModeDTO{
			Code:    Fail,
			Message: "缺少参数" + key,
			Data:    NullData,
		}).ToJson()
	}
	// 执行查询
	table, err := factory.Services[tableName].GetModel(v)
	if err != nil {
		log.Println(err)
		return (&gutils.CodeModeDTO{
			Code:    Fail,
			Message: ExecuteFailMessage,
			Data:    NullData,
		}).ToJson()
	}
	// 判断数据是否存在
	m := table.ToMap()
	if len(m) < 1 {
		return (&gutils.CodeModeDTO{
			Code:    Success,
			Message: SuccessMessage,
			Data:    NullData,
		}).ToJson()
	}
	data := m[0]
	// 组合外键
	keys := factory.DataBase.GetTable(tableName).Keys
	for i := 0; i < len(keys); i++ {
		query := data[gutils.ToBigHump(keys[i].ColumnName)]
		table, err := factory.Services[gutils.ToSmallHump(keys[i].RelyTable)].GetModel(query)
		if err != nil {
			log.Println(err)
			return (&gutils.CodeModeDTO{
				Code:    Fail,
				Message: ExecuteFailMessage,
				Data:    NullData,
			}).ToJson()
		}
		tempMap := table.ToMap()
		if len(tempMap) >= 1 {
			data[gutils.ToBigHump(keys[i].RelyTable)] = tempMap[0]
		}
	}
	return (&gutils.CodeModeDTO{
		Code:    Success,
		Message: SuccessMessage,
		Data:    data,
	}).ToJson()
}

func (simpleApi *SimpleApi) insert(factory mysqlutils.SessionFactory, tableName string, ctx iris.Context) string {
	// 解析参数
	values, err := simpleApi.Values(factory, tableName, ctx)
	if err != nil {
		log.Println(err)
		return (&gutils.CodeModeDTO{
			Code:    Fail,
			Message: DataAnalysisFailMessage,
			Data:    NullData,
		}).ToJson()
	}
	// 类型推断格式校验
	table := factory.DataBase.GetTable(tableName)
	for i := 0; i < len(table.Columns); i++ {
		// 栏位信息
		column := table.Columns[i]
		key := gutils.ToBigHump(column.Name)
		// 分析字段类型 长度
		sp := strings.Split(column.Type, "(")
		columnType := sp[0]
		switch columnType {
		case "varchar":
			{
				l := int64(len(values[key].(string)))
				strColumnLength := strings.Split(sp[1], ")")[0]
				columnLength, err := strconv.ParseInt(strColumnLength, 10, 64)
				if err != nil {
					log.Println(err)
					return (&gutils.CodeModeDTO{
						Code:    Fail,
						Message: "数据长度错误",
						Data:    NullData,
					}).ToJson()
				}
				if l == 0 || l > columnLength {
					return (&gutils.CodeModeDTO{
						Code:    Fail,
						Message: key + "字段长度错误",
						Data:    NullData,
					}).ToJson()
				}
				break
			}
		default:
			{
				break
			}
		}
	}
	// 数据插入
	key, err := factory.Services[tableName].Insert(values)
	if err != nil {
		log.Println(err)
		return (&gutils.CodeModeDTO{
			Code:    Fail,
			Message: ExecuteFailMessage,
			Data:    NullData,
		}).ToJson()
	}
	return (&gutils.CodeModeDTO{
		Code:    Success,
		Message: SuccessMessage,
		Data:    key,
	}).ToJson()
}

func (simpleApi *SimpleApi) update(factory mysqlutils.SessionFactory, tableName string, ctx iris.Context) string {
	// 解析参数
	values, err := simpleApi.Values(factory, tableName, ctx)
	if err != nil {
		log.Println(err)
		return (&gutils.CodeModeDTO{
			Code:    Fail,
			Message: DataAnalysisFailMessage,
			Data:    NullData,
		}).ToJson()
	}
	// 数据更新
	key, err := factory.Services[tableName].Update(values)
	if err != nil {
		log.Println(err)
		return (&gutils.CodeModeDTO{
			Code:    Fail,
			Message: ExecuteFailMessage,
			Data:    NullData,
		}).ToJson()
	}
	return (&gutils.CodeModeDTO{
		Code:    Success,
		Message: SuccessMessage,
		Data:    key,
	}).ToJson()
}

func (simpleApi *SimpleApi) delete(factory mysqlutils.SessionFactory, tableName string, ctx iris.Context) string {
	// 解析参数
	values, err := simpleApi.Values(factory, tableName, ctx)
	if err != nil {
		log.Println(err)
		return (&gutils.CodeModeDTO{
			Code:    Fail,
			Message: DataAnalysisFailMessage,
			Data:    NullData,
		}).ToJson()
	}
	// 根据索引确定主键
	key := ""
	indexs := factory.DataBase.GetTable(tableName).Indexs
	for i := 0; i < len(indexs); i++ {
		index := indexs[i]
		if index.Name == "PRIMARY" {
			key = gutils.ToBigHump(index.ColumnName)
		}
	}
	// 判断主键是非存在
	v, ok := values[key]
	if !ok {
		return (&gutils.CodeModeDTO{
			Code:    Fail,
			Message: "缺少参数" + key,
			Data:    NullData,
		}).ToJson()
	}
	// 执行查询
	count, err := factory.Services[tableName].Delete(v)
	if err != nil {
		log.Println(err)
		return (&gutils.CodeModeDTO{
			Code:    Fail,
			Message: ExecuteFailMessage,
			Data:    NullData,
		}).ToJson()
	}
	return (&gutils.CodeModeDTO{
		Code:    Success,
		Message: SuccessMessage,
		Data:    count,
	}).ToJson()
}

// unknown操作
func (simpleApi *SimpleApi) unknown(factory mysqlutils.SessionFactory, tableName string, ctx iris.Context) string {
	return "unknown request path"
}

// ========================================= 参数处理层 ========================================= //

// 读取数据
func (simpleApi *SimpleApi) Values(factory mysqlutils.SessionFactory, tableName string, ctx iris.Context) (map[string]interface{}, error) {
	m := make(map[string]string)
	// 读取表单中的数据
	formMap, err := simpleApi.formValues(factory, tableName, ctx)
	if err != nil {
		return nil, err
	}
	for k, v := range formMap {
		m[gutils.ToBigHump(k)] = v
	}
	// 读取JSON中的数据
	jsonMap, err := simpleApi.jsonValues(factory, tableName, ctx)
	if err != nil {
		return nil, err
	}
	for k, v := range jsonMap {
		m[gutils.ToBigHump(k)] = v
	}
	// 清洗数据
	attr, err := simpleApi.clear(factory, tableName, m)
	if err != nil {
		return nil, err
	}
	return attr, nil
}

// 读取表单数据
func (simpleApi *SimpleApi) formValues(factory mysqlutils.SessionFactory, tableName string, ctx iris.Context) (map[string]string, error) {
	values := make(map[string]string)
	for k, v := range ctx.FormValues() {
		values[k] = v[0]
	}
	return values, nil
}

// 读取JSON数据
func (simpleApi *SimpleApi) jsonValues(factory mysqlutils.SessionFactory, tableName string, ctx iris.Context) (map[string]string, error) {
	values := make(map[string]string)
	if ctx.GetHeader("Content-Type") != "application/json" {
		return values, nil
	}
	err := ctx.ReadJSON(&values)
	return values, err
}

// 清洗数据
func (simpleApi *SimpleApi) clear(factory mysqlutils.SessionFactory, tableName string, m map[string]string) (map[string]interface{}, error) {
	// 清洗数据
	// 规则是遍历表结构信息
	// 并按表列类型初始化attrMap或者转化m中的值
	attrMap := make(map[string]interface{})
	// 获取并遍历表结构信息
	table := factory.DataBase.GetTable(tableName)
	for i := 0; i < len(table.Columns); i++ {
		// 栏位信息
		column := table.Columns[i]
		key := gutils.ToBigHump(column.Name)
		v, ok := m[key]
		switch strings.Split(column.Type, "(")[0] {
		case "int":
			{
				if ok {
					value, err := strconv.ParseInt(v, 10, 64)
					if err != nil {
						return nil, err
					}
					attrMap[key] = value
					break
				}
				attrMap[key] = 0
				break
			}
		case "datetime":
			{
				if ok {
					value, err := time.Parse("2006-01-02 15:04:05", v)
					if err != nil {
						return nil, err
					}
					attrMap[key] = value
					break
				}
				attrMap[key] = time.Time{}
				break
			}
		default:
			{
				if ok {
					attrMap[key] = v
					break
				}
				attrMap[key] = nil
				break
			}
		}
	}
	// 处理分页
	v, ok := m["Start"]
	if ok {
		value, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		attrMap["Start"] = value
	} else {
		attrMap["Start"] = 0
	}
	v, ok = m["Length"]
	if ok {
		value, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		attrMap["Length"] = value
	} else {
		attrMap["Length"] = 10
	}
	return attrMap, nil
}
