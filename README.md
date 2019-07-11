# FET
FET is a go template engineer that can translate code to `html/template` code.


## Why FET
FET means Friendly, Easily for Templating.`html/template` has a basic support for templating, but it's not easy to use, so you need FET.
- Expression support
- Use `incldue` with defined variables scopes
- Use `extends` inherit base template with defined variables scopes
- Extends support for `for` loop, e.g

## Document
  
[Document](https://github.com/fefit/fet/wiki/Wiki)    

[中文文档](https://github.com/fefit/fet/wiki/%E4%B8%AD%E6%96%87%E6%96%87%E6%A1%A3)

## Usage

it's more like the php template engineer smarty.

- inherit

  ```php
  {%extends "base.html"%}
  ```

- blocks for inherit
  
  ```php
  {%block "header"%}
    <div>some code here</div>
  {%/block%}
  ```

- include

  ```php
  {%include "header.html"%}
  ```

- loop, do not support keyword `break` `continue`
  
  ```php
  // for Gofet mode
  {%for item,key in list%}
    // output
  {%/for%}
  
  {%for i = 0, j = 10; i < j; i++%}
    // output
  {%/for%}
  // for Smarty mode
  {%foreach $list as $key => $item%}
    // output
  {%/foreach%}
  // for
  {%for $i = 0, $j = 10; $i < $j; $i++%}
    // output
  {%/for%}
  ```

- if condition
  
  ```php
  {%if $num > 100%}
  
  {%elseif $num < 50%}
  
  {%else%}
  
  {%/if%}
  ```

- output
  
  ```php
  {%$item.url%}
  ```

- pipe funcs

  ```php
  {%$item.title|truncate:30%}
  ```

- variable define
  
  ```php
  {%$title = "this is a title"%}
  ```

- capture  
  
  ```php
  {%capture "hello"%}  
  {%/capture%} 
  {%$fet.capture.hello%}
  ```

- static variables  

  ```php
  {%$fet.capture.xxx%}
  {%$fet.config.leftDelimiter%}
  {%$fet.config.rightDelimiter%}
  {%$fet.config.templateDir%}
  {%$fet.config.compileDir%}
  {%$fet.now%}
  ```

- special variables  
  ```php
  {%$ROOT%}  // will output $
  {%$fet%} // will output .
  ```

### Expression
  
1. operators    
  `+ - * / % ! ** == >= <= != && || & ^`

2. keyword operators  
  `and` && `or` || `not` ! `eq` == `ne` != `gt` > `ge` >= `lt` < `le` <=  
  `bitor` for "|"

3. pipe   
  `|` pipeline funcs  
  `:` set arguments for pipeline funcs

4. numbers    
  hex: `0xffff`   
  octal: `0o777`  
  binary: `0b1000`  
  scientific notation `1e10`

Be careful of the `and` and `or` operators, they don't have short circuit with conditions. 

### Characters concat  
  ```php
  {% $sayHello = "world" %}
  {% "hello `$sayHello`"%} // output "hello world"
  ```
  use ` `` ` for variable or expression in strings. do not use `+`.
  

### Func Maps  
  - Math  
    `min`   `max` `floor` `ceil`
  
  - Helpers  
    `number_format` `truncate`
  
  - Assert  
    `empty`

  - Length  
    `count`
  
  - Output  
    `safe`  
  - [view more in funcs.go](./lib/funcs/funcs.go)
  
### Config types.Mode  
  - types.Smarty  
  the variable and field must begin with `$`, use `foreach` tag for loops.  
  
  
  - types.Gofet  
  the variable and field mustn't begin with `$`, use `for` tag for loops.  

### Demo code
```go
package main

import (
  "os"
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
    UcaseField: false, // default false, if true will auto uppercase field name to uppercase.
    CompileOnline: false, // default false, you should compile your template files offline 
    Glob: false, // default false, if true, will add {{define "xxx"}}{{end}} to wrap the compiled content,"xxx" is the relative pathname base on your templateDir, without the file extname.
    AutoRoot: false, // default false,if true, if the variable is not assign in the scope, will trait it as the root field of template data, otherwise you need use '$ROOT' to index the data field.
    Mode: types.Smarty, // default types.Smarty, also can be "types.Gofet"
  }
  fet, _ := fet.New(conf)
  // assign data
  data := map[string]interface{}{
    "Hello": "World"
  }
  // Display
  fet.Display("index.html", data, os.Stdout)
}
```
### API 
- `fet.Compile(tpl string, createFile bool) (result string, err error) `  

  compile a template file, if `createFile` is true, will create the compiled file.  

- `fet.CompileAll() error`  
  
  compile all files need to compile.  


- `fet.Display(tpl string, data interface{}, output io.Wirter) error`

  render the parsed html code into `output`.  

- `fet.Fetch(tpl string, data interface{}) (result string, err error)`

  just get the parsed `string` code, it always use `CompileOnline` mode.  


## Use in project

1.  `compile mode`  
    just use fet compile your template files offline, and add the FuncMap `lib/funcs/funcs.go` to `html/template` struct.

2. `install mode`  

    install `fet`,and use `fet.Display(tpl, data, io.Writer)` to render the template file.


## License

[MIT License](./LICENSE).
