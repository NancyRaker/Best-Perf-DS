/*
@Time : 2019/9/18 15:27
@Author : yanKoo
@File : download
@Software: GoLand
@Description:
*/
package main

import (
	"bufio"
	"file-server/conf"
	"file-server/pub/fdfs_client"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"
)

var durl = "https://www.yunptt.com/group1/M00/00/B6/eQ6Vtl2B7Y2AQ9N_AAzodRl2Jqw352.jpg"

var client *fdfs_client.Client

func init() {
	var err error
	client, err = fdfs_client.NewClientWithConfig()
	if err != nil {
		fmt.Errorf("Client: %+v NewClientWithConfig fastdfs error: %+v", client, err)
	}
	// defer client.Destory() TODO destory?
}


func download() {
	uri, err := url.ParseRequestURI(durl)
	if err != nil {
		panic("网址错误")
	}

	filename := path.Base(uri.Path)
	log.Println("[*] Filename " + filename)

	cli := http.DefaultClient
	cli.Timeout = time.Second * 60 //设置超时时间
	resp, err := cli.Get(durl)
	if err != nil {
		panic(err)
	}
	if resp.ContentLength <= 0 {
		log.Println("[*] Destination server does not support breakpoint download.")
	}
	raw := resp.Body
	defer raw.Close()
	reader := bufio.NewReaderSize(raw, 1024*1024*10)

	b, err :=ioutil.ReadAll(reader)
	fmt.Println(len(b), cap(b))
	fmt.Println(err)

	// 存储文件到fastdfs
	fileId, err := client.UploadByBuffer(b, filename)
	if err != nil {
		fmt.Printf("UploadByBuffer to fastdfs error: ", err)
	}
	fmt.Println(conf.FILE_BASE_URL + fileId)
}
