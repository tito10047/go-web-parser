package main

import (
	"stavkova/sites"
	"strings"
	xhtml "golang.org/x/net/html"
	"github.com/andybalholm/cascadia"
	"fmt"
)

func main() {
	d, err := sites.NewDownloader("GET", "https://www.skybet.com/")
	check(err)
	html, err := d.Download()
	check(err)
	doc, err := xhtml.Parse(strings.NewReader(html))

	selector, err := cascadia.Compile("#nav .section:nth-child(5n+4)>ul>a")
	check(err)

	aNodes := selector.MatchAll(doc)

	for _,aNode := range aNodes {
		for _,attr := range aNode.Attr{
			fmt.Println(attr.Val)
		}
	}
}



func check(e error) {
	if e!=nil {
		panic(e)
	}
}