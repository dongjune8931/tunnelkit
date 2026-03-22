// previewd 피드백 오버레이 - 순수 TypeScript (프레임워크 없음)
// Vite IIFE 번들로 빌드되어 서버에서 HTML에 주입됨

import './overlay.css'

declare global {
  interface Window {
    __PREVIEWD_SERVER__: string
  }
}

type State = 'IDLE' | 'SELECTING' | 'COMMENTING'

interface FeedbackPin {
  id: string
  element: Element
  x: number
  y: number
  comment: string
}

class PreviewdOverlay {
  private state: State = 'IDLE'
  private serverURL: string
  private sessionID: string
  private btn!: HTMLButtonElement
  private popover: HTMLDivElement | null = null
  private selectedEl: Element | null = null
  private pins: FeedbackPin[] = []

  constructor() {
    this.serverURL = window.__PREVIEWD_SERVER__ || 'http://localhost:8081'
    this.sessionID = this.extractSession()
    this.init()
  }

  private extractSession(): string {
    // 서브도메인에서 세션 추출 (간단히 hostname 사용)
    return location.hostname.split('.')[0] || 'default'
  }

  private init() {
    // "피드백 남기기" 버튼 생성
    this.btn = document.createElement('button')
    this.btn.id = '__previewd-btn'
    this.btn.textContent = '💬 피드백 남기기'
    this.btn.addEventListener('click', () => this.toggleSelecting())
    document.body.appendChild(this.btn)

    // ESC로 취소
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') this.cancel()
    })
  }

  private toggleSelecting() {
    if (this.state === 'IDLE') {
      this.enterSelecting()
    } else {
      this.cancel()
    }
  }

  private enterSelecting() {
    this.state = 'SELECTING'
    this.btn.textContent = '✕ 취소'
    this.btn.classList.add('active')
    document.body.style.cursor = 'crosshair'

    document.addEventListener('mouseover', this.handleHover)
    document.addEventListener('click', this.handleClick, { capture: true })
  }

  private handleHover = (e: MouseEvent) => {
    const el = e.target as Element
    if (el === this.btn || this.btn.contains(el)) return
    document.querySelectorAll('.__previewd-highlight').forEach((el) =>
      el.classList.remove('__previewd-highlight')
    )
    el.classList.add('__previewd-highlight')
  }

  private handleClick = (e: MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()

    const el = e.target as Element
    if (el === this.btn || this.btn.contains(el)) return

    this.selectedEl = el
    el.classList.remove('__previewd-highlight')
    this.showPopover(e.clientX, e.clientY)
  }

  private showPopover(x: number, y: number) {
    this.state = 'COMMENTING'
    document.removeEventListener('mouseover', this.handleHover)
    document.removeEventListener('click', this.handleClick, { capture: true })
    document.body.style.cursor = ''

    const authorName = localStorage.getItem('__previewd_name') || ''

    const pop = document.createElement('div')
    pop.id = '__previewd-popover'

    // 화면 경계 보정
    const popX = Math.min(x + 10, window.innerWidth - 300)
    const popY = Math.min(y + 10, window.innerHeight - 220)
    pop.style.left = `${popX}px`
    pop.style.top = `${popY}px`

    pop.innerHTML = `
      <div style="font-size:13px;font-weight:700;color:#374151;margin-bottom:8px">피드백 남기기</div>
      <input id="__pd-name" type="text" placeholder="이름 (선택)" value="${authorName}" />
      <textarea id="__pd-comment" placeholder="어떤 피드백인가요?"></textarea>
      <button class="submit-btn" style="margin-top:10px">제출</button>
      <button class="cancel-btn">취소</button>
    `

    pop.querySelector('.submit-btn')!.addEventListener('click', () => this.submit())
    pop.querySelector('.cancel-btn')!.addEventListener('click', () => this.cancel())

    document.body.appendChild(pop)
    this.popover = pop
    ;(pop.querySelector('#__pd-comment') as HTMLTextAreaElement).focus()
  }

  private async submit() {
    if (!this.popover) return

    const comment = (this.popover.querySelector('#__pd-comment') as HTMLTextAreaElement).value.trim()
    const authorName = (this.popover.querySelector('#__pd-name') as HTMLInputElement).value.trim()

    if (!comment) {
      alert('피드백 내용을 입력해주세요')
      return
    }

    if (authorName) {
      localStorage.setItem('__previewd_name', authorName)
    }

    const elementCSS = this.selectedEl ? getUniqueCSSSelector(this.selectedEl) : ''

    try {
      const res = await fetch(`${this.serverURL}/api/tunnels/${this.sessionID}/feedback`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          page_url: location.href,
          element_css: elementCSS,
          x_percent: 50,
          y_percent: 50,
          comment,
          author_name: authorName,
        }),
      })

      if (res.ok) {
        this.removePoper()
        this.state = 'IDLE'
        this.btn.textContent = '💬 피드백 남기기'
        this.btn.classList.remove('active')
        this.showSuccessToast()
      }
    } catch {
      alert('피드백 전송에 실패했습니다. 서버 연결을 확인하세요.')
    }
  }

  private cancel() {
    document.querySelectorAll('.__previewd-highlight').forEach((el) =>
      el.classList.remove('__previewd-highlight')
    )
    document.removeEventListener('mouseover', this.handleHover)
    document.removeEventListener('click', this.handleClick, { capture: true })
    document.body.style.cursor = ''
    this.removePoper()
    this.state = 'IDLE'
    this.btn.textContent = '💬 피드백 남기기'
    this.btn.classList.remove('active')
  }

  private removePoper() {
    if (this.popover) {
      this.popover.remove()
      this.popover = null
    }
  }

  private showSuccessToast() {
    const toast = document.createElement('div')
    toast.textContent = '✓ 피드백이 전송됐습니다!'
    toast.style.cssText = `
      position:fixed;bottom:80px;right:24px;z-index:999999;
      background:#22c55e;color:#fff;padding:12px 20px;border-radius:10px;
      font-family:system-ui,sans-serif;font-size:14px;font-weight:600;
      box-shadow:0 4px 16px rgba(34,197,94,0.4);
    `
    document.body.appendChild(toast)
    setTimeout(() => toast.remove(), 3000)
  }
}

// 엘리먼트의 고유 CSS 셀렉터 생성
function getUniqueCSSSelector(el: Element): string {
  const path: string[] = []
  let current: Element | null = el

  while (current && current !== document.body) {
    let selector = current.tagName.toLowerCase()

    if (current.id) {
      selector += `#${current.id}`
      path.unshift(selector)
      break
    }

    if (current.className) {
      const classes = Array.from(current.classList)
        .filter((c) => !c.startsWith('__previewd'))
        .slice(0, 2)
        .join('.')
      if (classes) selector += `.${classes}`
    }

    // 형제 중 순서 파악
    const parent = current.parentElement
    if (parent) {
      const siblings = Array.from(parent.children).filter(
        (s) => s.tagName === current!.tagName
      )
      if (siblings.length > 1) {
        const idx = siblings.indexOf(current) + 1
        selector += `:nth-of-type(${idx})`
      }
    }

    path.unshift(selector)
    current = current.parentElement
  }

  return path.join(' > ')
}

// 자동 초기화
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => new PreviewdOverlay())
} else {
  new PreviewdOverlay()
}
