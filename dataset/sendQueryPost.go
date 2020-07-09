/*
@Time : 2019/12/4 11:25 
@Author : yanKoo
@File : sendQueryPost
@Software: GoLand
@Description:
*/
package dataset

import (
	"dingding/dataset/model"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
)

const (
	JSP_URL      = "http://femhzs.mofcom.gov.cn/fecpmvc/pages/fem/fem_cert_stat_view_list.jsp"
	PIC_CODE_URL = "http://femhzs.mofcom.gov.cn/fecpmvc/pages/fem/1.randImages"
)

type cookieJar struct {
	cookies []*http.Cookie
}

func (jar *cookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.cookies = append(jar.cookies, cookies...)
}
func (jar *cookieJar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies
}

type QueryParser struct {
	jspUrl    string // 请求路径
	picUrl    string // 验证码请求地址
	cr        *CodeRecognizer
	picCode   string
	client    *http.Client
	cookieJar cookieJar
}

func NewQueryParser(picSavePath, picCodeResultFile string) *QueryParser {
	jar := new(cookieJar)
	jar.SetCookies(nil, []*http.Cookie{
		//{Name: "JSESSIONID", Value: "4F02ECA79120BC12742B2F48D12BD7A3", Path: "/fecpmvc"},
		//{Name: "insert_cookie", Value: "97222558", Path: "/"},
	})
	client := &http.Client{nil, nil, jar, math.MaxInt32}

	qp := &QueryParser{
		jspUrl: JSP_URL,
		picUrl: PIC_CODE_URL,
		cr: &CodeRecognizer{
			picSavePath:       picSavePath,
			picCodeResultFile: picCodeResultFile,
		},
		client: client,
	}

	// 初始化cookie
	return qp.initQueryParser()
}

func (qp *QueryParser) initQueryParser() *QueryParser {
	// 发送cookie 为空的post请求
	resp, err := qp.client.PostForm(qp.jspUrl,
		map[string][]string{
			"INVEST_CN1": {"山东汇峰装备科技股份有限公司"},
			"CORP_NA_CN": {"汇峰国际（乌兹别克斯坦）有限公司"},
			"IS_QUERY":   {"QUERY"},
			"forCode":    {"1"},
		})

	if err != nil {
		log.Printf("init queryparser err:%+v", err)
	}
	cookies := resp.Cookies()
	log.Printf("cookie: %+v \n", cookies)

	qp.cookieJar.SetCookies(nil, cookies)
	return qp
}

func (qp *QueryParser) StartQuery(index int, node *model.CompanyListNode, retry int) int {
	if retry > 10 {
		return -1
	}
	// 1. 获取图片保存起来，然后给后面去识别code
	qp.GetPicture()
	code := qp.cr.RecognitionVerificationCode()
	log.Printf("qp get code: %s node: %+v", code, node)
	// 2. 然后去获取查询的两个参数（公司名字）去查询
	resp, err := qp.client.PostForm(qp.jspUrl,
		map[string][]string{
			"INVEST_CN1": {node.Domestic},
			"CORP_NA_CN": {node.Overseas},
			"code":       {code},
			"IS_QUERY":   {"QUERY"},
			"forCode":    {"1"},
		})
	if err != nil {
		return -1
	}


	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
		return -1
	}

	tempRes := doc.Find(".listTableClass tr TD").Map(func(i int, s *goquery.Selection) string {
		text := strings.TrimSpace(s.Text())
		return text
	})

	//body, err := ioutil.ReadAll(resp.Body)
	//_ = ioutil.WriteFile("query.html", body, 0644)

	if len(tempRes) < 8 {
		//log.Printf("erro")
		return qp.StartQuery(index, node, retry+1)
	}
	log.Printf("temp Res: %+v, len: %d", tempRes[4:8], len(tempRes[4:8]))
	node.Overseas = tempRes[4]
	node.Domestic = tempRes[5]
	node.Citizenship = tempRes[6]
	node.AckDate = tempRes[7]

	return index
}

func (qp *QueryParser) GetPicture() {
	resp, err := qp.client.Get(qp.picUrl)
	if err != nil {
		fmt.Printf("with error: %+v", err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	_ = ioutil.WriteFile(qp.cr.picSavePath, body, 0644)
}
