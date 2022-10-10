module github.com/BlackCarDriver/Driver-tools

go 1.16

require (
	github.com/BlackCarDriver/GoProject-api v1.0.5
	github.com/astaxie/beego v1.12.3
	github.com/fatih/color v1.13.0
	golang.design/x/clipboard v0.6.1
)

exclude golang.design/x/clipboard v0.6.2

//replace github.com/BlackCarDriver/GoProject-api => ../api
