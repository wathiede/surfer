package htmlutil

import (
	"strings"

	"golang.org/x/net/html"
)

func GetText(n *html.Node) string {
	text := []string{}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			text = append(text, c.Data)
		default:
			text = append(text, GetText(c))
		}
	}

	return strings.TrimSpace(strings.Join(text, ""))
}
