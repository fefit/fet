# FET

[![tag](https://img.shields.io/github/v/tag/fefit/fet.svg?sort=semver)](https://github.com/fefit/fet/tags)
[![Build Status](https://travis-ci.com/fefit/fet.svg?branch=master)](https://travis-ci.com/github/fefit/fet)
[![codecov](https://codecov.io/gh/fefit/fet/branch/master/graph/badge.svg)](https://codecov.io/gh/fefit/fet)

FET is a golang template engine that can tranlate smarty like template code into golang `html/template`.

## Why FET

FET means Friendly, Easily for Template code. The official golang package `html/template` is a fully functional template engine, but it has a few defects with user experience, so you need FET.

### Features

- Expression logics
- Use `incldue` with defined variables scopes
- Use `extends` inherit base template with defined variables scopes
- Limited support for `for` and `capture` 

## Document

[Document](https://github.com/fefit/fet/wiki/Wiki)

[中文文档](https://github.com/fefit/fet/wiki/%E4%B8%AD%E6%96%87%E6%96%87%E6%A1%A3)

## Usage

it's more likely to the php template engine smarty.

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
  {%include file="header.html" var=1%}
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
  {%capture "hello"%}hello{%/capture%}
  {%if true%}
    {%$fet.capture.hello%},wolrd!
  {%else%}
    just {%$fet.capture.hello%}!
  {%/if%}
  ```

- static variables

  ```php
  {%$fet.capture.xxx%}
  {%$fet.config.leftDelimiter%}
  {%$fet.config.rightDelimiter%}
  {%$fet.config.templateDir%}
  {%$fet.config.compileDir%}
  {%$fet.now%}
  {%$fet.debug%} // will output all the variables that assigned in the template<include variables that difined in the template by yourself> to the js devtools's Console panel.
  ```

- special variables
  ```php
  {%$ROOT%}  // will output $
  {%$fet%} // will output .
  ```

### Expression

1. operators: You can either use the keyword operator or the punctuation.

  
    | keyword  | punctuation  | example |
    |---|---|---|
    |  `and` | `&&`  | `1 && 2`  <=>  `1 and 2` |
    |  `or` | `\|\|`   | `1 \|\| 2` <=> `1 or 2`|
    |  `not` | `!`  | `!a` <=> `not a`|
    |`eq`| `==`| `a == b` <=> `a eq b`|
    |  `ne` | `!=`  | `a != b` <=> `a ne b`|
    |`gt`| `>`| `a > b` <=> `a gt b`|
    |`ge`| `>=`| `a >= b` <=> `a ge b`|
    |`lt`| `<`| `a < b` <=> `a lt b`|
    |`le`| `<=`| `a <= b` <=> `a le b`|
    | `bitor` | - | `a bitor b` |
    | - | `&` | `a & b`|
    | - | `^` | `a ^ b`| 
    | - | `+` | `a + b`| 
    | - | `-` | `a - b`| 
    | - | `*` | `a * b`| 
    | - | `/` | `a / b`| 
    | - | `%` | `a % b`| 
    | - | `**` | `a ** b`|

  
    Be careful of the `and` and `or` operators, they don't have short circuit with conditions.

2. pipe  
   `|` pipeline funcs  
   `:` set parameters for pipeline funcs

3. numbers  
   hex: `0xffff`  
   octal: `0o777`  
   binary: `0b1000`  
   scientific notation `1e10`

### String concat

```php
{% $sayHello = "world" %}
{% "hello `$sayHello`"%} // output "hello world"
```

use ` `` ` for variable or expression in strings. do not use `+`.

### Func Maps

- Math  
  `min` `max` `floor` `ceil`

- Formats  
  `number_format` 
  
- Strings 
  `truncate` `concat` `ucwords`

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

* types.Gofet  
  the variable and field mustn't begin with `$`, use `for` tag for loops.

### In development

```bash
# install command line tool `fetc`
go get -v github.com/fefit/fetc
# then init the config, will make a config file `fet.config.json`
fetc init
# then watch the file change, compile your fet template file immediately
fetc watch
```

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
    Ignores: []string{"inc/*"}, // ignore compile,paths and files that will be included and extended. use filepath.Match() method.
    UcaseField: false, // default false, if true will auto uppercase field name to uppercase.
    CompileOnline: false, // default false, you should compile your template files offline
    Glob: false, // default false, if true, will add {{define "xxx"}}{{end}} to wrap the compiled content,"xxx" is the relative pathname base on your templateDir, without the file extname.
    AutoRoot: false, // default false,if true, if the variable is not assign in the scope, will treat it as the root field of template data, otherwise you need use '$ROOT' to index the data field.
    Mode: types.Smarty, // default types.Smarty, also can be "types.Gofet"
  }
  fet, _ := fet.New(conf)
  // assign data
  data := map[string]interface{}{
    "Hello": "World"
  }
  // the index.html {%$ROOT.Hello%}
  // Display
  fet.Display("index.html", data, os.Stdout)
  // will output: World
}

```

### API

#### static methods

- `fet.LoadConf(configFile string) (*types.FetConfig, error)`

  if you use the command line `fetc` build a config file `fet.config.json`, then you can use `fet.LoadConf` to get the config.

- `fet.New(config *types.FetConfig) (instance *Fet, error)`

  get a fet instance.

#### instance methods

- `instance.Compile(tpl string, createFile bool) (result string, err error)`

  compile a template file, if `createFile` is true, will create the compiled file.

- `instance.CompileAll() error`

  compile all files need to compile.

* `instance.Display(tpl string, data interface{}, output io.Wirter) error`

  render the parsed html code into `output`.

* `instance.Fetch(tpl string, data interface{}) (result string, err error)`

  just get the parsed `string` code, it always use `CompileOnline` mode.

## Use in project

1.  `compile mode`

    just use fet compile your template files offline, and add the FuncMap `lib/funcs/funcs.go` to your project.

2.  `install mode`

    install `fet`,and use `fet.Display(tpl, data, io.Writer)` to render the template file.

## License

[MIT License](./LICENSE).
