@echo off
setlocal enabledelayedexpansion

echo Building Windows executables...

set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0

if not exist bin mkdir bin

for /d %%i in (cmd\*) do (
    echo.
    echo Processing: %%i
    
    if exist "%%i\main.go" (
        for %%j in ("%%i") do set exe_name=%%~nj
        
        echo Building !exe_name!...
        
        rem build the whole command directory as a package (all *.go files)
        go build -v -trimpath -ldflags "-s -w" -o bin\!exe_name!.exe .\%%i
        
        if !errorlevel! equ 0 (
            echo Successfully built !exe_name!.exe
        ) else (
            echo Failed to build !exe_name!
        )
    ) else (
        echo Warning: main.go not found in %%i
    )
)

echo.
echo Build process completed.
::pause
