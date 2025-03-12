$d = $args[0]
if ($d.length -lt 1) {
    Write-Host "Specify directory to copy executable."
}
else {
    go build

    if ($LASTEXITCODE -eq 0) {
        if (-not (Test-Path $d -PathType Container)) {
            New-Item -Path $d -ItemType Directory
        }
        $m = (Get-Content "go.mod" | Select-Object -First 1) -replace "^module "
        $n = "{0}.exe" -f ($m -split "/" | Select-Object -Last 1)
        if (Test-Path $n) {
            Get-Item $n | Copy-Item -Destination $d -Force -ErrorAction Stop
            "COPIED {0} to: {1}" -f $n, $d | Write-Host -ForegroundColor Blue
        }
        else {
            "{0} not found!" -f $n | Write-Host -ForegroundColor Magenta
        }
    }
    else {
        "Failed to build. Nothing was copied." | Write-Host -ForegroundColor Magenta
    }
}
