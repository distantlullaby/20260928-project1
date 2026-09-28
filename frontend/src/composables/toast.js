import { reactive } from 'vue'

// 全局轻提示
export const toastState = reactive({ message: '', type: 'success', visible: false })
let timer = null

export function toast(message, type = 'success') {
  toastState.message = message
  toastState.type = type
  toastState.visible = true
  clearTimeout(timer)
  timer = setTimeout(() => { toastState.visible = false }, 2600)
}
