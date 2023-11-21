package structure

import "encoding/xml"

type MarkBody struct {
	Id   string `json:"id"`
	Key  string `json:"key"`
	Done bool   `json:"done"`
}

type DivInfo struct {
	XMLName xml.Name `xml:"div"`
	Text    string   `xml:",chardata"`
	Class   string   `xml:"class,attr"`
	Div     struct {
		Text  string `xml:",chardata"`
		Class string `xml:"class,attr"`
		Span  []struct {
			Text   string `xml:",chardata"`
			Class  string `xml:"class,attr"`
			DataJp string `xml:"data-jp,attr"`
		} `xml:"span"`
	} `xml:"div"`
	Time string `xml:"time"`
}
