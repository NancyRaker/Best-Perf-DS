/*
@Time : 2019/12/4 11:13 
@Author : yanKoo
@File : main
@Software: GoLand
@Description:
*/
package main

import "dingding/dataset"

func main() {
	//dataset.Post()
	//dataset.GetPicture()

	//queryParser := dataset.NewQueryParser("codepic.jpg", "codeResult.txt")
	//queryParser.StartQuery()
	//queryParser.StartQuery()
	//queryParser.StartQuery()

	//dataset.ReadQuerySource()

	e := dataset.NewEngine(dataset.NewScheduler(), 100)
	e.Run()
}
