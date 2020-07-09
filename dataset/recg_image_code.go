/*
@Time : 2019/12/4 11:05 
@Author : yanKoo
@File : recg_image_code
@Software: GoLand
@Description:
*/
package dataset

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)
const ShellToUse = "tesseract"

type CodeRecognizer struct {
	picSavePath string
	picCodeResultFile string  // "codeResult.txt"
}

func (cr *CodeRecognizer)RecognitionVerificationCode() string {
	e, out, errInfo := cr.GenerateVerifCodeFile(cr.picSavePath)
	if e != nil {
		log.Println(e, out, errInfo)
		return ""
	}

	file, e := os.OpenFile(cr.picCodeResultFile, os.O_RDWR, 066)
	if e != nil {
		log.Printf("open file error : %+v", e)
		return ""
	}
	defer file.Close()

	res, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("read file err := %+v", e)
		return  ""
	}

	//err = os.Remove(cr.picCodeResultFile)

	return strings.TrimSpace(string(res))
}

func (cr *CodeRecognizer) GenerateVerifCodeFile(picPath string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	codeResultFileName := strings.Split(cr.picCodeResultFile, ".")
	cmd := exec.Command(ShellToUse ,picPath , codeResultFileName[0], "-l" , "eng")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}



