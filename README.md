# FET
FET is a go template engineer that can translate code to `html/template` code.

FET能按自身支持的语法来将模板编译成html/template的模板语法，当然这只是第一步。

## [FET的来由]Why FET
FET means Friendly, Easily for Templating.`html/template` has a basic support for templating, but it's not easy to use, so you need FET.
- Expression support
- Use `incldue` with defined variables scopes
- Use `extends` inherit base template with defined variables scopes
- Extends support for `for` loop, e.g

`html/template` 作为go官方的模板引擎，虽然基础的功能已有支持，但对于开发人员来说，其书写语法十分原始，写起来很困难，所以这才有了FET。


## [使用方式]Usage

it's more like the php template engineer smarty.

使用方式与PHP的模板引擎smarty基本类似，但个别功能因为`html/template`无法实现而放弃支持。

- [继承]inherit

  `{%extends "base.html"%}`

- [继承block块]blocks for inherit
  
  `{%block "header"%}`

    `<div>some code here</div>`
  
  `{%/block%}`

- [引入包含文件]include

  `{%include "header.html"%}`

- [循环]loop, do not support `break` `continue`
  
  ```go
  // for Gofet mode
  {{for item,key in list}}
    // output
  {{/for}}
  
  {{for i = 0, j = 10; i < j; i++}}
    // output
  {{/for}}
  // for Smarty mode
  {{foreach $list as $key => $item}
    // output
  {{/foreach}}}
  // for
  {{for $i = 0, $j = 10; $i < $j; $i++}}
    // output
  {{/for}}
  ```

- [条件判断]if condition
  
  `{{if $num > 100}}`
  
  `{{elseif $num < 50}}`
  
  `{{else}}`
  
  `{{/if}}`

- [数据输出]output
  
  `{{$item.url}}`

- [使用方法管道]pipe funcs

  `{{$item.title|truncate:30}}`

- [定义变量]variable assign
  
  `{{$title = "this is a title"}}`

### [表达式支持]Expression
  
1. operators    
  `+ - * / % ! ** == >= <= != && || & ^`

2. keyword operators  
  `and` && `or` || `not` ! `eq` == `ne` != `gt` < `ge` <= `lt` > `le` >=  
  `bitor` for "|"

3. pipe   
  `|` pipeline funcs  
  `:` set arguments for pipeline funcs

4. numbers    
  hex: `0xffff`   
  octal: `0o777`  
  binary: `0b1000`  
  scientific notation `1e10`
### [字符串拼接]Characters concat  
  ```go
  {{ $var = "world" }}
  {{ "hello `$var`"}} // output "hello world"
  ```
  use ` `` ` for variable or expression in strings. do not use `+`.
  
  使用 ` `` `符号来包裹变量或者表达式来达到拼接字符串的目的，`+`号仅用作数字类型的加法运算，请勿使用。

### [内置的函数]Func Maps  

  [view func.go](./lib/funcs/func.go)
### [渲染模式]Config types.Mode  
  - types.Smarty  
  the variable and field must begin with `$`, use `foreach` tag for loops.  
  变量和字段必须使用$符号做为开头，否则会报错，无法编译。按照smarty的方式使用foreach来做循环。
  
  - types.Gofet  
  the variable and field mustn't begin with `$`, use `for` tag for loops.  
  变量和字段不能以$符号开头，否则会报错，无法编译。 

### [示例代码]Demo code
```go
package main

import (
  "github.com/fefit/fet"
  "github.com/fefit/fet/types"
)

func main(){
  conf := &fet.Config{
    LeftDelimiter: "{%", // default "{%"
    RightDelimiter: "%}", // default "%}"
    TemplateDir: "tmpls", //  default "templates"
    CompileDir: "views", // default "templates_c",
    Ignores: []string{"inc/*"}, // ignore compile paths,files that only will include.use filepath.Match
    LowerField: true, // default false, if true will not translate keys to uppercase.
    CompileOnline: false, // default false, you should compile your template files offline 
    Mode: types.Smarty, // default types.Smarty, also can be "types.Gofet"
  }
  fet, _ := fet.New(conf)
  // compile all files
  fet.CompileAll()
  // data := map[string]string{
  // }
  // fet.Display(tpl, data)
}
```
### API 
- `fet.Compile(tpl string, createFile bool) (result string, err error) `  

  compile a template file, if `createFile` is true, will create the compiled file.  

  编译单个文件，如果createFile参数设为true，将生成对应的编译文件到编译目录里，文件的相对目录路径和原始的fet模板文件相对路径保持一致。

- `fet.CompileAll() error`  
  
  compile all files need to compile.  

  编译所有需要编译的文件，除了那些在Ignores配置中设置了无需编译的文件。

- `fet.Display(tpl string, data interface{}, output io.Wirter) error`

  render the parsed html code into `output`.  

  将模板文件渲染后的结果输出显示。

- `fet.Fetch(tpl string, data interface{}) (result string, err error)`

  just get the parsed `string` code, it always use `CompileOnline` mode.  

  获取模板文件渲染后的代码，始终按当前fet模板文件的内容编译，然后渲染得到结果。

## [如何在项目中使用]Use in project

1. [仅编译模式] `compile mode`  
    just use fet compile your template files offline, and add the FuncMap `lib/funcs/func.go` to `html/template` struct.

    在项目中，使用fet的方式来编写模板，然后将编写好的模板编译到项目中，引入`lib/funcs/func.go`里写好的通用方法，由go后端同学注册进去，走线下编译模式。

2. [安装模式] `install mode`  

    install `fet`,and use `fet.Display(tpl, data, os.Stdout)` to render the template file.

    将fet安装到项目中，同时需要将fet打包到项目内上线，使用fet的api来渲染输出模板。

## License

[MIT License](./LICENSE).
