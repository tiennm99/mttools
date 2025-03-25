@echo off
setlocal enabledelayedexpansion

if not exist output mkdir output

for /d %%D in (input\*) do (
    set "foldername=%%~nxD"
    if not exist "output\!foldername!.txt" type nul > "output\!foldername!.txt"
    
    for %%F in ("%%D\*.txt") do (
        type "%%F" >> "output\!foldername!.txt"
        echo. >> "output\!foldername!.txt"
    )
)

echo Process completed.