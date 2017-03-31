package scenario

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"sync"

	"github.com/sesta/isucon6-final/bench/action"
	"github.com/sesta/isucon6-final/bench/fails"
	"github.com/sesta/isucon6-final/bench/session"
)

type StaticFile struct {
	Path string
	MD5  string
}

func loadStaticFiles(s *session.Session, checkHash bool) bool {
	assets := []StaticFile{
		StaticFile{Path: "/css/rc-color-picker.css", MD5: "78055c5c02a2dd66f6207fa19f7ca928"},
		StaticFile{Path: "/css/sanitize.css", MD5: "7375990d0f1f7d436a952314e3ac7fd0"},
		StaticFile{Path: "/bundle.js", MD5: "b1070be102c1f9d5aaace9300fe6f193"},
	}
	var wg sync.WaitGroup

	OK := true
	for _, asset := range assets {
		wg.Add(1)
		go func(asset StaticFile) {
			defer wg.Done()

			ok := action.Get(s, asset.Path, action.OK(func(body io.Reader, l *fails.Logger) bool {
				content, err := ioutil.ReadAll(body)
				if err != nil {
					l.Add("ファイルの読み込みに失敗しました", err)
					return false
				}

				if checkHash {
					actual := fmt.Sprintf("%x", md5.Sum(content))
					if actual != asset.MD5 {
						l.Add("ファイルの内容が正しくありません",
							fmt.Errorf("expected %s, actual %s, content: %s", asset.MD5, actual, string(content[:20])))
						return false
					}
				}
				return true
			}))
			if !ok {
				OK = false
			}
		}(asset)
	}
	wg.Wait()
	return OK
}
