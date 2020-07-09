/*
@Time : 2019/12/4 14:53 
@Author : yanKoo
@File : persistence
@Software: GoLand
@Description:
*/
package dataset

import (
	"dingding/dataset/model"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"log"
	"strconv"
)

func ReadQuerySource() []*model.CompanyListNode {
	f, err := excelize.OpenFile("./company_data.xlsx")
	if err != nil {
		log.Println(err)
		return nil
	}

	// Get value from cell by given worksheet name and axis.
	cell := f.GetCellValue(SheetName, "B2")
	if err != nil {
		log.Println(err)
		return nil
	}
	log.Println(cell)
	// Get all the rows in the Sheet1.
	rows := f.GetRows(SheetName)
	log.Printf("rows: %d cols: %d", len(rows), len(rows[0]))
	var source []*model.CompanyListNode
	for i, row := range rows{
		if i == 0 {
			continue
		}
		source = append(source, &model.CompanyListNode{
			Overseas:    row[0],
			Domestic:    row[1],
			Citizenship: row[2],
		})
	}

	return source//[0:1000]
}

func SaveCompanyInfo(list []*model.CompanyListNode) {
	file, err := excelize.OpenFile("./company_data.xlsx")
	if err != nil {
		log.Println(err)
		return
	}
	// Set value of a cell.
	file.SetCellValue(SheetName, "F1", "境外投资企业(机构)")
	file.SetCellValue(SheetName, "G1", "境内投资主体/(境内投资者名称)")
	file.SetCellValue(SheetName, "H1", "投资国别（地区）")
	file.SetCellValue(SheetName, "I1", "核准日期")

	for i, node := range list {
		if node == nil {
			continue
		}
		//fmt.Printf("%+v\n", node)
		file.SetCellValue(SheetName, "F"+strconv.Itoa(i+2), node.Overseas)
		file.SetCellValue(SheetName, "G"+strconv.Itoa(i+2), node.Domestic)
		file.SetCellValue(SheetName, "H"+strconv.Itoa(i+2), node.Citizenship)
		file.SetCellValue(SheetName, "I"+strconv.Itoa(i+2), node.AckDate)
	}

	err = file.SaveAs("./company_data.xlsx")
	if err != nil {
		fmt.Println(err)
	}
}
