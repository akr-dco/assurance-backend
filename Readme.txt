### Running Development (With Powershell)
$env:APP_ENV="development"
go run main.go


### Running Production (With Powershell)
$env:APP_ENV="production"
go run main.go


### Check Env (With Powershell)
echo $env:APP_ENV


### Running Development (With Linux)
APP_ENV=development go run main.go


### Running Production (With Linux)
APP_ENV=production go run main.go
