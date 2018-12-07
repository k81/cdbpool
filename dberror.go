package cdbpool

import "fmt"

type DBError struct {
	ErrCode int32
	ErrMsg  string
	SqlInfo *MysqlInfo
}

func (e *DBError) Error() string {
	return fmt.Sprintf("ErrCode: %v(%v), ErrMsg: %v, SqlInfo: <%v>", e.ErrCode, ErrCodeName(e.ErrCode), e.ErrMsg, e.formatSqlInfo())
}

func (e *DBError) GetMysqlErrno() uint32 {
	if ResultCode(e.ErrCode) == ResultCode_RC_SUCCESS || e.SqlInfo == nil {
		return 0
	}

	return e.SqlInfo.MysqlErrno
}

func (e *DBError) formatSqlInfo() string {
	if e.SqlInfo == nil {
		return ""
	}

	return fmt.Sprintf("vsid:%v, ip:%v, port:%v, dbname:%v, sql:%v, errno:%v",
		e.SqlInfo.Vsid,
		e.SqlInfo.MysqlIp,
		e.SqlInfo.MysqlPort,
		e.SqlInfo.Dbname,
		e.SqlInfo.Sql,
		e.SqlInfo.MysqlErrno,
	)
}

func NewDBError(code int32, msg string, sqlInfo *MysqlInfo) *DBError {
	return &DBError{
		ErrCode: code,
		ErrMsg:  msg,
		SqlInfo: sqlInfo,
	}
}

func ErrCodeName(code int32) string {
	if name, ok := ResultCode_name[code]; ok {
		return name
	}
	return "unknown"
}
