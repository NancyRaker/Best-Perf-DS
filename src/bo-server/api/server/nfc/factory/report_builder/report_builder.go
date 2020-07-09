/*
@Time : 2019/10/17 11:15 
@Author : yanKoo
@File : builder
@Software: GoLand
@Description:
*/
package report_builder

type builder struct {

}

// 创建报告生成器
func newBuilder() *builder{
	return &builder{}
}

//
func (b *builder)generateReport(reportTask interface{}) interface{} {
	return nil
}