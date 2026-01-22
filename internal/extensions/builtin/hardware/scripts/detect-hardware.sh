#!/bin/bash
# Hardware detection script for Linux/macOS
# Outputs JSON with CPU, RAM, and GPU/VRAM information

set -e

# Initialize JSON object
echo "{"

# CPU Detection
if command -v lscpu &> /dev/null; then
    CPU_CORES=$(lscpu | grep "^CPU(s):" | awk '{print $2}')
    CPU_MODEL=$(lscpu | grep "Model name:" | sed 's/Model name:[[:space:]]*//' | sed 's/"/\\"/g')
else
    # Fallback for macOS
    if [[ "$OSTYPE" == "darwin"* ]]; then
        CPU_CORES=$(sysctl -n hw.ncpu)
        CPU_MODEL=$(sysctl -n machdep.cpu.brand_string | sed 's/"/\\"/g')
    else
        CPU_CORES="unknown"
        CPU_MODEL="unknown"
    fi
fi

echo "  \"cpu_cores\": $CPU_CORES,"
echo "  \"cpu_model\": \"$CPU_MODEL\","

# RAM Detection
if command -v free &> /dev/null; then
    # Linux
    TOTAL_RAM_GB=$(free -g | grep "^Mem:" | awk '{print $2}')
elif [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    TOTAL_RAM_GB=$(sysctl -n hw.memsize | awk '{print int($1/1024/1024/1024)}')
elif [ -f /proc/meminfo ]; then
    # Fallback: parse /proc/meminfo
    TOTAL_RAM_KB=$(grep "^MemTotal:" /proc/meminfo | awk '{print $2}')
    TOTAL_RAM_GB=$((TOTAL_RAM_KB / 1024 / 1024))
else
    TOTAL_RAM_GB=0
fi

echo "  \"total_ram_gb\": $TOTAL_RAM_GB,"

# GPU Detection - Try nvidia-smi first (most accurate for NVIDIA)
GPU_DETECTED=false
GPU_NAME=""
GPU_VRAM_GB=0
DETECTION_METHOD="none"

if command -v nvidia-smi &> /dev/null; then
    # Try to get NVIDIA GPU info
    if nvidia-smi --query-gpu=name,memory.total --format=csv,noheader,nounits 2>/dev/null | head -n1 | read -r gpu_name gpu_memory; then
        GPU_DETECTED=true
        GPU_NAME=$(echo "$gpu_name" | sed 's/"/\\"/g' | xargs)
        GPU_VRAM_GB=$((gpu_memory / 1024))
        DETECTION_METHOD="nvidia-smi"
    fi
fi

# Fallback: Try lspci for other GPUs
if [ "$GPU_DETECTED" = false ] && command -v lspci &> /dev/null; then
    GPU_INFO=$(lspci | grep -iE "(vga|3d|display)" | head -n1)
    if [ -n "$GPU_INFO" ]; then
        GPU_DETECTED=true
        GPU_NAME=$(echo "$GPU_INFO" | sed 's/.*: //' | sed 's/"/\\"/g')
        GPU_VRAM_GB=0  # lspci doesn't provide VRAM info
        DETECTION_METHOD="lspci"
    fi
fi

# macOS GPU detection
if [ "$GPU_DETECTED" = false ] && [[ "$OSTYPE" == "darwin"* ]]; then
    GPU_NAME=$(system_profiler SPDisplaysDataType 2>/dev/null | grep "Chipset Model:" | head -n1 | sed 's/.*Chipset Model: //' | sed 's/"/\\"/g')
    if [ -n "$GPU_NAME" ]; then
        GPU_DETECTED=true
        # macOS doesn't easily provide VRAM info, try to get it
        GPU_VRAM_MB=$(system_profiler SPDisplaysDataType 2>/dev/null | grep "VRAM" | head -n1 | grep -oE "[0-9]+" | head -n1)
        if [ -n "$GPU_VRAM_MB" ]; then
            GPU_VRAM_GB=$((GPU_VRAM_MB / 1024))
        else
            GPU_VRAM_GB=0
        fi
        DETECTION_METHOD="system_profiler"
    fi
fi

echo "  \"gpu_detected\": $GPU_DETECTED,"
echo "  \"gpu_name\": \"$GPU_NAME\","
echo "  \"gpu_vram_gb\": $GPU_VRAM_GB,"
echo "  \"detection_method\": \"$DETECTION_METHOD\""

echo "}"
