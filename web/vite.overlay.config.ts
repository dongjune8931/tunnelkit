import { defineConfig } from 'vite'

// 피드백 오버레이를 단독 IIFE 번들로 빌드 (React 없음)
export default defineConfig({
  build: {
    lib: {
      entry: 'src/overlay/overlay.ts',
      name: 'PreviewdOverlay',
      fileName: 'overlay',
      formats: ['iife'],
    },
    outDir: '../server/web/dist/static',
    emptyOutDir: false,
  },
})
