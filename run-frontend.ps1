# 前端 Vite 开发服务器（Node 通过 nvm 管理，直接使用系统 node）
Set-Location "$PSScriptRoot\frontend"
if (-not (Test-Path "node_modules")) {
  npm install --registry=https://registry.npmmirror.com
}
npm run dev
