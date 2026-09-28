import { reactive } from 'vue'

export const toastState = reactive({ items: [] })
let seq = 0

export function toast(message, type = 'ok') {
  const id = ++seq
  toastState.items.push({ id, message, type })
  setTimeout(() => {
    const i = toastState.items.findIndex((t) => t.id === id)
    if (i > -1) toastState.items.splice(i, 1)
  }, 2600)
}
