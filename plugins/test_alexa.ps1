# Test script for Alexa Skill plugin (PowerShell)

Write-Host "Testing Alexa Skill Plugin" -ForegroundColor Cyan
Write-Host "==========================" -ForegroundColor Cyan
Write-Host ""

# Test 1: List plugins
Write-Host "1. Listing plugins..." -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "http://localhost:11435/v1/plugins" -Method Get
$response | ConvertTo-Json -Depth 10
Write-Host ""

# Test 2: Get Alexa plugin metadata
Write-Host "2. Getting Alexa plugin metadata..." -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "http://localhost:11435/v1/plugins/alexa_skill" -Method Get
$response | ConvertTo-Json -Depth 10
Write-Host ""

# Test 3: Test Alexa endpoint with wake word
Write-Host "3. Testing Alexa endpoint with wake word..." -ForegroundColor Yellow
$body = @{
    version = "1.0"
    session = @{
        new = $true
        sessionId = "test-session-1"
        application = @{
            applicationId = "test-app"
        }
        user = @{
            userId = "test-user"
        }
    }
    request = @{
        type = "IntentRequest"
        requestId = "test-request-1"
        timestamp = "2024-01-01T00:00:00Z"
        locale = "en-US"
        intent = @{
            name = "QueryIntent"
            slots = @{
                query = @{
                    name = "query"
                    value = "Smart Voice what is the weather today"
                }
            }
        }
    }
} | ConvertTo-Json -Depth 10

$response = Invoke-RestMethod -Uri "http://localhost:11435/v1/plugins/alexa_skill/alexa" -Method Post -Body $body -ContentType "application/json"
$response | ConvertTo-Json -Depth 10
Write-Host ""

# Test 4: Test Alexa endpoint without wake word
Write-Host "4. Testing Alexa endpoint without wake word..." -ForegroundColor Yellow
$body = @{
    version = "1.0"
    session = @{
        new = $true
        sessionId = "test-session-2"
        application = @{
            applicationId = "test-app"
        }
        user = @{
            userId = "test-user"
        }
    }
    request = @{
        type = "IntentRequest"
        requestId = "test-request-2"
        timestamp = "2024-01-01T00:00:00Z"
        locale = "en-US"
        intent = @{
            name = "QueryIntent"
            slots = @{
                query = @{
                    name = "query"
                    value = "what is the weather today"
                }
            }
        }
    }
} | ConvertTo-Json -Depth 10

$response = Invoke-RestMethod -Uri "http://localhost:11435/v1/plugins/alexa_skill/alexa" -Method Post -Body $body -ContentType "application/json"
$response | ConvertTo-Json -Depth 10
Write-Host ""

Write-Host "Tests complete!" -ForegroundColor Green
