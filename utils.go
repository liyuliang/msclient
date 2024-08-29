package msclient

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func ReadConf(path string) Config {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	c := Config{}
	_ = json.Unmarshal(b, &c)
	return c
}

func CmdCode() string {

	println()
	println("请输入code：")
	// 创建一个从标准输入读取数据的 Reader 对象
	reader := bufio.NewReader(os.Stdin)

	// 读取用户输入的一行文本，直到遇到换行符为止
	code, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("读取输入时出错:", err)
		return ""
	}
	// 去除输入文本末尾的换行符
	code = strings.TrimSpace(code)
	println("你输入的是：")
	println(code)
	return code
}

func RegexGet(content string, expr string) string {
	result := ""
	rule, _ := regexp.Compile(expr)
	allMatch := rule.FindStringSubmatch(content)
	if 2 == len(allMatch) {
		result = allMatch[1]
	}
	return result
}
