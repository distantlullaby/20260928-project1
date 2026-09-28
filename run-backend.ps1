# Go
$env:Path = "C:\mine\code\tool\Google\go\go1.22.1\sdk\bin;" + $env:Path
$env:GOPROXY = "https://goproxy.cn,direct"
Set-Location "$PSScriptRoot\backend"
go run .
