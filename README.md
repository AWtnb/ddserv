# README

markDown Document Server.


```
> .\ddserv.exe -h
Usage of ~\go\bin\ddserv.exe:
  -export
        export as sigle html file
  -plain
        prevent loading css from https://github.com/AWtnb/md-stylesheet
  -src string
        markdown path
```

Markdown file can include frontmatter as below:

```
---
title: title of html
load:
  - style.css
  - style2.css
---
```
