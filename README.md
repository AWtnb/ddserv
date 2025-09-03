# README

Markdown to html converter.


```
> .\m2h.exe -h
Usage of C:\Users\awtnb\Sync\develop\repo\m2h\m2h.exe:
  -plain
        flag to skip loading https://cdn.jsdelivr.net/gh/Awtnb/md-less@latest/style.less
  -src string
        markdown path
  -suffix string
        suffix of result html
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
