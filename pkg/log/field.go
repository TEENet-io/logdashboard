package log

// Field 日志字段
// 用于扩展日志的自定义内容

type Field struct {
	Key   string
	Value interface{}
}

// NewField 构造函数
func NewField(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// FieldFunc 构造函数的别名，为了向后兼容
func FieldFunc(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}
