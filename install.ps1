$ErrorActionPreference = "Stop"

$Repo = "zaaack/go-bin"
$Binary = "go-bin"
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "." }

$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { Write-Error "Only 64-bit systems are supported"; exit 1 }

$Platform = "windows-$Arch"
$Url = "https://github.com/$Repo/releases/latest/download/go-bin-$Platform.zip"

Write-Host "Downloading $Binary for $Platform..."
$TmpDir = New-Item -ItemType Directory -Path ([System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString())

try {
    Invoke-WebRequest -Uri $Url -OutFile "$TmpDir\go-bin.zip" -UseBasicParsing
    Expand-Archive -Path "$TmpDir\go-bin.zip" -DestinationPath "$TmpDir" -Force

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item "$TmpDir\go-bin-windows-amd64.exe" "$InstallDir\$Binary.exe" -Force

    Write-Host "Installed to $InstallDir\$Binary.exe"
    Write-Host "Run: .\$Binary.exe serve"
} finally {
    Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
}
