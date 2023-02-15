package main

import (
	"crypto/tls"
	"embed"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

//go:embed form.html result.html
var content embed.FS
var files, _ = template.ParseFS(content, "*")
var re = regexp.MustCompile(`data-lazy-srcset="([^"]+)"`)
var prefix = "data-lazy-srcset="

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	http.ListenAndServe(":"+os.Getenv("PORT"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			return
		}
		renderForm(w, r)
		if r.Method == "POST" || r.Form.Has("url") {
			w.Write([]byte("<hr>"))
			renderResult(w, r)
		}
	}))
}

func getTitle(url string) string {
	url = strings.TrimSuffix(url, "/")
	idx := strings.LastIndex(url, "/")
	if idx > 0 {
		return strings.TrimPrefix(url[idx:], "_")
	}
	return "funda image downloader"
}

func renderForm(w http.ResponseWriter, r *http.Request) {
	err := files.ExecuteTemplate(w, "form.html", map[string]interface{}{"Url": r.Form.Get("url"), "Title": getTitle(r.Form.Get("url"))})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func renderResult(w http.ResponseWriter, r *http.Request) {
	result, err := lookupImages(r.URL.Query().Get("url"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = files.ExecuteTemplate(w, "result.html", result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type result struct {
	Title   string
	Images  []struct{ Src string }
	Details string
}

func lookupImages(url string) (result, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return result{}, err
	}
	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result{}, err
	}
	limitedReader := &io.LimitedReader{R: resp.Body, N: 200 * 1000}
	defer resp.Body.Close()

	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return result{}, err
	}
	r := result{}
	for _, m := range re.FindAllSubmatch(data, -1) {
		if len(m) >= 1 {
			param := strings.Trim(strings.TrimPrefix(string(m[1]), prefix), `"`)
			srcs := Map(strings.Split(param, ","), parseSrc)
			sort.Sort(ImgList(srcs))
			r.Images = append(r.Images, struct{ Src string }{Src: srcs[0].url})
		}
	}
	// r.Details = "<code><pre>" + string(data) + "</pre></code>"
	return r, nil
}

func Map[T any, R any](t []T, fn func(t T) R) (out []R) {
	out = make([]R, len(t))
	for _, item := range t {
		out = append(out, fn(item))
	}
	return out
}

type img struct {
	url  string
	size int
}

func TrimToNum(r rune) bool {
	n := r - '0'
	return !(n >= 0 && n <= 9)
}

func parseSrc(in string) img {
	in = strings.TrimSpace(in)
	p := strings.SplitN(in, " ", 2)
	if len(p) > 1 {
		size, _ := strconv.Atoi(strings.TrimFunc(p[1], TrimToNum))
		return img{p[0], size}
	}
	return img{in, 0}
}

var _ sort.Interface = ImgList([]img{})

type ImgList []img

func (list ImgList) Len() int {
	return len(list)
}

func (list ImgList) Less(i, j int) bool {
	return list[i].size-list[j].size > 0
}

func (list ImgList) Swap(i, j int) {
	tmp := list[i]
	list[i] = list[j]
	list[j] = tmp
}
