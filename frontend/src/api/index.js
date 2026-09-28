import axios from 'axios'

const api = axios.create({ baseURL: '/api', timeout: 10000 })

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('mbti_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

api.interceptors.response.use(
  (res) => res.data,
  (err) => {
    const msg = err.response?.data?.error || '网络异常，请稍后再试'
    if (err.response?.status === 401) {
      localStorage.removeItem('mbti_token')
    }
    return Promise.reject(new Error(msg))
  }
)

export default api
