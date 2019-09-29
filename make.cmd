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
    for %%I in (386 amd64) do call :package1 windows "%%I" .exe
    for %%I in (386 amd64) do call :package1 linux   "%%I"
    exit /b

:package1
    setlocal
    set "GOOS=%~1"
    set "GOARCH=%~2"
    for /F %%I in ('cd') do set "NAME=%%~nI"
    set "SUFFIX=%~3"
    go build
    zip -9m "%NAME%-%DATE:/=%-%GOOS%-%GOARCH%.zip" "%NAME%%SUFFIX%"
    endlocal
    exit /b

:"update"
    for /F "skip=1" %%I in ('where tmt.exe') do copy /-Y tmt.exe "%%I"
    exit /b

:"clean"
    rmdir /s Cmd
    exit /b
