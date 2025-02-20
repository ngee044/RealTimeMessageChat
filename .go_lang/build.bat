@echo off
echo Starting build process...

:: Common 모듈 빌드
cd Common
echo Building Common module...
go mod tidy
go build ./...
IF %ERRORLEVEL% NEQ 0 (
    echo Failed to build Common module.
    exit /b %ERRORLEVEL%
)
cd ..

:: RestAPI 빌드
cd RestAPI
echo Setting up Common module in RestAPI...
go mod edit -replace Common=../Common
go mod tidy
echo Building RestAPI...
go build ./...
IF %ERRORLEVEL% NEQ 0 (
    echo Failed to build RestAPI.
    exit /b %ERRORLEVEL%
)
cd ..

:: MessageQueue 빌드
cd MessageQueue
echo Setting up Common module in MessageQueue...
go mod edit -replace Common=../Common
go mod tidy
echo Building MessageQueue...
go build ./...
IF %ERRORLEVEL% NEQ 0 (
    echo Failed to build MessageQueue.
    exit /b %ERRORLEVEL%
)
cd ..

echo Build process completed successfully!
exit /b 0
