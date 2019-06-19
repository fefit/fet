# FET
FET is a go template engineer that can translate code to `text/template` code.

FET能按自身支持的语法来将模板编译成text/template的模板语法，当然这只是第一步。

## Why FET
FET means Friendly, Easily for Templating.`text/template` has a basic support for templating, but it's not easy to use, so you need FET.

`text/template` 作为go官方的模板引擎，虽然基础的功能已有支持，但对于开发人员来说，其书写语法十分原始，写起来很困难，所以这才有了FET。


## Usage

it's more like the php template engineer smarty.

使用方式与PHP的模板引擎smarty类似，个别地方有差异

- [继承]inherit

  `{{extends "base.html"}}`

- [继承block块]blocks for inherit
  
  `{{block "header"}}`

    `<div>some code here</div>`
  
  `{{/block}}`

- [引入包含文件]include

  `{{include "header.html"}}`

- [循环]loop
  
  `{{for item,key in list}}`

  `{{/for}}`

- [条件判断]if condition
  
  `{{if num > 100}}`
  
  `{{elseif num < 50}}`
  
  `{{else}}`
  
  `{{/if}}`

- [数据输出]output
  
  `{{item.url}}`

- [使用方法管道]pipe funcs

  `{{item.title|truncate:30}}`

- [定义变量]variable assign
  
  `{{title = "this is a title"}}`

### [表达式支持]Expression
  
1. operators  
`+ - * / % ! ** == >= <= != && || & bitor ^`

2. pipe   
  `|` and `:` for arguments

3. numbers  
  hex: `0xffff` octal: `0o777` binary: `0b1000` scientific notation `1e10`

### [内置的函数]Func Maps

### [项目中使用]Use in project
```go
package main

import "fet"

func main(){
  conf := &fet.Config{
    TemplateDir: "tmpls", //  default "templates"
    CompileDir: "views", // default "templates_c",
    Ignores: []string{"inc/*"}, // ignore compile paths,files that only will include.use filepath.Match
    Ucfirst: true, // default true, will translate map keys to uppercase.
  }
  fet, _ := fet.New(conf)
  // compile all files
  fet.CompileAll()
  // data := map[string]string{
  // }
  // fet.Display(tpl, data)
}
```


## License

[MIT License](./LICENSE).
