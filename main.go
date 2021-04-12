package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"unicode"

	"github.com/reiver/go-oi"
	"github.com/reiver/go-telnet"
	"jaytaylor.com/html2text"
)

var (
	M4Handler telnet.Handler = m4webproxyHandler{}
)

type m4webproxyHandler struct {
}

func (handler m4webproxyHandler) ServeTELNET(ctx telnet.Context, w telnet.Writer, r telnet.Reader) {
	oi.LongWriteString(w, "Welcome to m4 proxy web.\r\n")
	p := make([]byte, 1)
	var url string
	for {

		n, err := r.Read(p)
		//fmt.Printf("nb char read [%d] and char [%d]\n", n, p[0])
		if n > 0 {
			switch p[0] {
			case 10: // eol
				oi.LongWriteByte(w, p[0])
				oi.LongWriteString(w, " go to the site ["+url+"]\r\n")
				goWeb(string(url[:len(url)-1]), w, r)
				url = ""
			default:
				oi.LongWriteByte(w, p[0])
				url += string(p[0])
			}
		}

		if nil != err {
			break
		}
	}
}

// space byte 32

func goWeb(url string, w telnet.Writer, r telnet.Reader) {
	if url[0:4] != "http" { // add scheme if not exists
		url = "https://" + url
	}
	resp, err := http.Get(url)
	if err != nil {
		msg := fmt.Sprintf("Error while getting [%s] error :%v", url, err)
		fmt.Print(msg)
		oi.LongWriteString(w, msg+"\r \r\n")
	} else {
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("error while reading html page :%v\n", err)
		}
		out, err := html2text.FromString(string(b), html2text.Options{PrettyTables: true, OmitLinks: true})
		if err != nil {
			fmt.Printf("error while reading html page :%v\n", err)
		}

		out = strings.TrimSpace(cleanNonAscii(out))
		p := make([]byte, 1)
		outs := split(out, 18*80)
		for _, s := range outs {
			oi.LongWriteString(w, s)
			for {
				n, err := r.Read(p)
				if n > 0 {
					if p[0] == 32 { // space is pressed
						break
					}
				}
				if nil != err {
					break
				}
			}
		}
	}
}

func cleanNonAscii(s string) string {
	out := ""
	for _, v := range s {
		if v <= unicode.MaxASCII {
			out += string(v)
		}
	}
	return out
}

func split(s string, size int) []string {
	a := make([]string, 0)
	for i := 0; i < len(s); i += size {
		last := i + size
		if i+last > len(s) {
			last = len(s)
		}
		a = append(a, s[i:last])
	}
	return a
}

func main() {

	handler := M4Handler
	addr := ":23"
	if err := telnet.ListenAndServe(addr, handler); nil != err {
		panic(err)
	}
}
