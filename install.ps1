Param(
    [string]$Version = "latest"
)

$Owner = "shinshin86"
$Repo = "vpeak"
$BinName = "vpeak.exe"
$BinDir = "$env:LOCALAPPDATA\Programs\vpeak"

function Write-Info {
    Param([string]$Message)
    Write-Host "Info: $Message"
}

function Write-ErrorMessage {
    Param([string]$Message)
    Write-Host "Error: $Message" -ForegroundColor Red
}

function Assert-Command {
    Param([string]$Command)
    if (-not (Get-Command $Command -ErrorAction SilentlyContinue)) {
        throw "Required command not found: $Command"
    }
}

function Get-Arch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    if ($env:PROCESSOR_ARCHITEW6432) {
        $arch = $env:PROCESSOR_ARCHITEW6432
    }

    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { throw "Unsupported architecture: $arch" }
    }
}

function Get-LatestVersion {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Owner/$Repo/releases/latest"
    if (-not $release.tag_name) {
        throw "Failed to fetch latest version"
    }
    return $release.tag_name
}

function Get-Checksum {
    Param(
        [string]$ChecksumsPath,
        [string]$Asset
    )
    $lines = Get-Content -Path $ChecksumsPath
    foreach ($line in $lines) {
        if ($line -match "\s+$Asset$") {
            return $line.Split(" ")[0].ToLower()
        }
    }
    return $null
}

function Add-ToPath {
    Param([string]$Path)
    $current = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($null -eq $current) {
        $current = ""
    }
    if ($current -notlike "*$Path*") {
        $newPath = $current
        if ($newPath -ne "" -and -not $newPath.EndsWith(";")) {
            $newPath += ";"
        }
        $newPath += $Path
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        $env:Path = "$env:Path;$Path"
        Write-Info "Added $Path to PATH (user)"
    }
}

function Invoke-Download {
    Param(
        [string]$Uri,
        [string]$OutFile
    )
    Invoke-WebRequest -Uri $Uri -OutFile $OutFile -UseBasicParsing -MaximumRedirection 10 | Out-Null
}

$TempDir = $null

try {
    Assert-Command "PowerShell"
    Assert-Command "Expand-Archive"

    if ($Version -eq "latest") {
        Write-Info "Fetching latest version..."
        $Version = Get-LatestVersion
    }

    $Arch = Get-Arch
    $Asset = "vpeak_${Version}_windows_${Arch}.zip"
    $BaseUrl = "https://github.com/$Owner/$Repo/releases/download/$Version"

    $TempDir = New-Item -ItemType Directory -Path ([System.IO.Path]::GetTempPath()) -Name "vpeak_$((Get-Random))"
    $ZipPath = Join-Path $TempDir $Asset
    $ChecksumPath = Join-Path $TempDir "checksums.txt"

    Write-Info "Downloading $BaseUrl/$Asset..."
    Invoke-Download -Uri "$BaseUrl/$Asset" -OutFile $ZipPath

    Write-Info "Downloading checksums..."
    Invoke-Download -Uri "$BaseUrl/checksums.txt" -OutFile $ChecksumPath

    $Expected = Get-Checksum -ChecksumsPath $ChecksumPath -Asset $Asset
    if (-not $Expected) {
        throw "Checksum not found for $Asset"
    }

    Write-Info "Verifying checksum..."
    $Actual = (Get-FileHash -Algorithm SHA256 -Path $ZipPath).Hash.ToLower()
    if ($Actual -ne $Expected) {
        throw "Checksum mismatch. expected=$Expected actual=$Actual"
    }

    Write-Info "Installing to $BinDir..."
    if (-not (Test-Path $BinDir)) {
        New-Item -ItemType Directory -Path $BinDir | Out-Null
    }

    Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force

    $InstalledBin = Join-Path $BinDir $BinName
    if (Test-Path $InstalledBin) {
        $Backup = "$InstalledBin.bak"
        Write-Info "Backing up existing binary to $Backup"
        Copy-Item $InstalledBin $Backup -Force
    }

    Copy-Item (Join-Path $TempDir $BinName) $InstalledBin -Force

    Add-ToPath $BinDir

    Write-Info "Done! Run 'vpeak --version' to verify."
} catch {
    Write-ErrorMessage $_
    exit 1
} finally {
    if ($TempDir -and (Test-Path $TempDir)) {
        Remove-Item $TempDir -Recurse -Force
    }
}
