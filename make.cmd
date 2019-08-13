@echo off
setlocal
set "PROMPT=$G "
call :"%1"
endlocal
exit /b

:""
    echo Usage: %~dpfx0 {package,clean}
    exit /b

:"package"
    for %%I in (386 amd64) do call :package1 "%%I"
    exit /b

:package1
    setlocal & set "GOARCH=%~1"
    set "DIR=%~dp0.\Cmd\%GOARCH%"
    for /F %%I in ('go env GOOS') do set "GOOS=%%I"
    go build -o "%DIR%\tmt.exe"
    zip -j9 "tmt-%DATE:/=%-%GOOS%-%GOARCH%.zip" "%DIR%\tmt.exe"
    endlocal
    exit /b

:"update"
    for /F "skip=1" %%I in ('where tmt.exe') do copy /-Y tmt.exe "%%I"
    exit /b

:"clean"
    rmdir /s Cmd
    exit /b
