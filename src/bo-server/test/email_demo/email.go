/*
@Time : 2019/10/17 9:31 
@Author : yanKoo
@File : email
@Software: GoLand
@Description:liuyang240917@sina.com
06cfcc819b3e546b

*/

// main.go
// main.go
package main

import (
	"github.com/jordan-wright/email"
	"log"
	"net/smtp"
	"os"
)

func SendMail(fromUser, toUser, subject string) error {
	// NewEmail返回一个email结构体的指针
	e := email.NewEmail()
	// 发件人
	e.From = fromUser
	// 收件人(可以有多个)
	e.To = []string{toUser}
	// 邮件主题
	e.Subject = subject
	//// 解析html模板
	//t, err := template.ParseFiles("test.html")
	//if err != nil {
	//	return err
	//}
	//// Buffer是一个实现了读写方法的可变大小的字节缓冲
	//body := new(bytes.Buffer)
	//// Execute方法将解析好的模板应用到匿名结构体上，并将输出写入body中
	//t.Execute(body, struct {
	//	FromUserName string
	//	ToUserName   string
	//	TimeDate     string
	//	Message      string
	//}{
	//	FromUserName: "go语言",
	//	ToUserName:   "Sixah",
	//	TimeDate:     time.Now().Format("2006/01/02"),
	//	Message:      "golang是世界上最好的语言！",
	//})
	//// html形式的消息
	//e.HTML = body.Bytes()

	file, _:= os.Open("nfc-api.pdf")
	//f := ioutil.ReadAll(file)
	// 从缓冲中将内容作为附件到邮件中
	e.Attach(file, "nfc.pdf", "application/pdf")
	// 以路径将文件作为附件添加到邮件中
	//e.AttachFile("nfc-api.pdf")
	// 发送邮件(如果使用QQ邮箱发送邮件的话，passwd不是邮箱密码而是授权码)
	return e.Send("smtp.sina.com:587", smtp.PlainAuth("", "liuyang240917@sina.com", "06cfcc819b3e546b", "smtp.sina.com"))
}

func main() {
	fromUser := "golang<liuyang240917@sina.com>"
	toUser := "506498066@qq.com"
	subject := "hello,world"
	err := SendMail(fromUser, toUser, subject)
	if err != nil {
		log.Println("发送邮件失败", err)
		return
	}
	log.Println("发送邮件成功")
}
