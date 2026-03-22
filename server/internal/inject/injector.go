package inject

import (
	"bytes"
	"fmt"
)

// Injector는 HTML 응답에 피드백 오버레이 스크립트를 주입한다.
type Injector struct {
	DashboardURL string
}

func New(dashboardURL string) *Injector {
	return &Injector{DashboardURL: dashboardURL}
}

var bodyCloseTag = []byte("</body>")

// Inject는 HTML 본문의 </body> 직전에 오버레이 스크립트를 삽입한다.
func (inj *Injector) Inject(html []byte) []byte {
	idx := bytes.LastIndex(html, bodyCloseTag)
	if idx == -1 {
		// </body>가 없으면 끝에 추가
		return append(html, inj.snippet()...)
	}

	result := make([]byte, 0, len(html)+len(inj.snippet()))
	result = append(result, html[:idx]...)
	result = append(result, inj.snippet()...)
	result = append(result, html[idx:]...)
	return result
}

func (inj *Injector) snippet() []byte {
	s := fmt.Sprintf(`
<script>
  window.__PREVIEWD_SERVER__ = %q;
</script>
<script src="%s/static/overlay.js"></script>
`, inj.DashboardURL, inj.DashboardURL)
	return []byte(s)
}
