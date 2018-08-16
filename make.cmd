@setlocal
@set "PROMPT=$G "
@call :"%1"
@endlocal
@exit /b

:""
    @for %%I in (386 amd64) do (
        mkdir "%~dp0.\Cmd\%%~I"
        set "GOARCH=%%~I"
        go build -o "%~dp0.\Cmd\%%~I\tmt.exe"
    )
    @exit /b

:"clean"
    rmdir /s Cmd
    @exit /b
