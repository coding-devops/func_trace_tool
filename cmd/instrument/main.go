package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/coding-devops/func_trace_tool/instrumenter"
	"github.com/coding-devops/func_trace_tool/instrumenter/ast"
)

//  instrument_trace/cmd/instrument/main.go

func main() {
	var file string
	flag.StringVar(&file, "f", "", "ddd")
	flag.Parse()
	fmt.Println(file)
	var ins instrumenter.Instrumenter = ast.New("e.coding.net/open-studio/go/instrument_trace", "trace", "Trace")
	// 创建以ast方式实现Instrumenter接口的ast.instrumenter实例
	newSrc, err := ins.Instrument(file) // 向Go源文件所有函数注入Trace函数
	if err != nil {
		panic(err)
	}

	if newSrc == nil {
		// add nothing to the source file. no change
		fmt.Printf("no trace added for %s\n", file)
		return
	}

	fmt.Println(string(newSrc)) // 将生成的新代码内容输出到stdout上

	// 将生成的新代码内容写回原Go源文件
	if err = ioutil.WriteFile(file, newSrc, 0666); err != nil {
		fmt.Printf("write %s error: %v\n", file, err)
		return
	}
	fmt.Printf("instrument trace for %s ok\n", file)
}
