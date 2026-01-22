# Hardware detection script for Windows PowerShell
# Outputs JSON with CPU, RAM, and GPU/VRAM information

$ErrorActionPreference = "SilentlyContinue"

# Initialize output object
$output = @{}

# CPU Detection
try {
    $cpu = Get-CimInstance Win32_Processor | Select-Object -First 1
    $output.cpu_cores = $cpu.NumberOfCores * $cpu.NumberOfLogicalProcessors / $cpu.NumberOfCores
    if ($null -eq $output.cpu_cores -or $output.cpu_cores -eq 0) {
        $output.cpu_cores = $cpu.NumberOfLogicalProcessors
    }
    $output.cpu_model = $cpu.Name.Trim()
} catch {
    $output.cpu_cores = 0
    $output.cpu_model = "unknown"
}

# RAM Detection
try {
    $ram = Get-CimInstance Win32_ComputerSystem
    $output.total_ram_gb = [math]::Round($ram.TotalPhysicalMemory / 1GB)
} catch {
    $output.total_ram_gb = 0
}

# GPU Detection
$output.gpu_detected = $false
$output.gpu_name = ""
$output.gpu_vram_gb = 0
$output.detection_method = "none"

# Try nvidia-smi first (most accurate for NVIDIA)
try {
    $nvidiaSmi = & nvidia-smi.exe --query-gpu=name,memory.total --format=csv,noheader,nounits 2>$null
    if ($LASTEXITCODE -eq 0 -and $nvidiaSmi) {
        $gpuInfo = $nvidiaSmi | Select-Object -First 1
        $parts = $gpuInfo -split ","
        if ($parts.Length -ge 2) {
            $output.gpu_detected = $true
            $output.gpu_name = ($parts[0].Trim() -replace '"', '')
            $vramMB = [int]($parts[1].Trim())
            $output.gpu_vram_gb = [math]::Round($vramMB / 1024)
            $output.detection_method = "nvidia-smi"
        }
    }
} catch {
    # nvidia-smi not available, continue to fallback
}

# Fallback: Use Win32_VideoController
if (-not $output.gpu_detected) {
    try {
        $gpu = Get-CimInstance Win32_VideoController | Where-Object { $_.AdapterRAM -and $_.AdapterRAM -gt 0 } | Select-Object -First 1
        if ($gpu) {
            $output.gpu_detected = $true
            $output.gpu_name = $gpu.Name.Trim()
            
            # AdapterRAM is in bytes, convert to GB
            # Note: On some systems this may include shared memory, so it might be inaccurate
            $adapterRAMBytes = $gpu.AdapterRAM
            if ($adapterRAMBytes -gt 0) {
                $output.gpu_vram_gb = [math]::Round($adapterRAMBytes / 1GB)
                # Cap at reasonable maximum (some systems report very large values for shared memory)
                if ($output.gpu_vram_gb -gt 128) {
                    $output.gpu_vram_gb = 0  # Likely shared memory, not dedicated VRAM
                }
            }
            
            if ($output.detection_method -eq "none") {
                $output.detection_method = "win32_videocontroller"
            }
        }
    } catch {
        # GPU detection failed
    }
}

# Output as JSON
$output | ConvertTo-Json -Compress
