package dataset

import (
	"dingding/dataset/model"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func init() {

}

type CompanyParse struct {
	parseResult []*model.CompanyListNode
}

func (cp *CompanyParse)StartParse() {
	for i := 0; i < 4163; i++ {
		fmt.Println(i)
		cp.scrape(i)
	}
	fmt.Printf("parse result: %d 个\n", len(cp.parseResult))
	cp.saveCompanyInfo()
}

func (cp *CompanyParse)scrape(i int) {
	// Request the HTML page.
	res, err := http.Get("http://femhzs.mofcom.gov.cn/fecpmvc/pages/fem/CorpJWList_nav.pageNoLink.html?sp="  + strconv.Itoa(i))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	tempRes := doc.Find(".m-table tbody tr td").Map(func(i int, s *goquery.Selection) string {
		text := strings.TrimSpace(s.Text())
		return text
	})

	//fmt.Printf("%+v, %d", tempRes, len(tempRes))
	cp.parserCompanyInfo(tempRes)
}

func (cp *CompanyParse)parserCompanyInfo(singlePageResult []string) {
	var list = make([]*model.CompanyListNode, 0)

	for i := 0; i < len(singlePageResult); i += 3 {
		list = append(list, &model.CompanyListNode{
			Overseas:    singlePageResult[i],
			Domestic:    singlePageResult[i+1],
			Citizenship: singlePageResult[i+2],
		})
	}

	//saveCompanyInfo(list)
	cp.parseResult = append(cp.parseResult, list...)
}

const SheetName = "境外投资企业（机构）备案结果公开名录列表"

func (cp *CompanyParse)saveCompanyInfo() {
	file := excelize.NewFile()
	// Create a new sheet.
	index := file.NewSheet(SheetName)
	// Set value of a cell.
	file.SetCellValue(SheetName, "A1", "境外投资企业(机构)")
	file.SetCellValue(SheetName, "B1", "境内投资主体/(境内投资者名称)")
	file.SetCellValue(SheetName, "C1", "投资国别（地区）")
	file.SetCellValue(SheetName, "D1", "核准日期")
	// Set active sheet of the workbook.
	file.SetActiveSheet(index)

	for i, node := range cp.parseResult {
		//fmt.Printf("%+v\n", node)
		file.SetCellValue(SheetName, "A"+strconv.Itoa(i+2), node.Overseas)
		file.SetCellValue(SheetName, "B"+strconv.Itoa(i+2), node.Domestic)
		file.SetCellValue(SheetName, "C"+strconv.Itoa(i+2), node.Citizenship)
		file.SetCellValue(SheetName, "D"+strconv.Itoa(i+2), "2019-12-03")
	}

	err := file.SaveAs("./company_data.xlsx")
	if err != nil {
		fmt.Println(err)
	}
}
