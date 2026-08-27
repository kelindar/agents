# Human-in-the-loop reproduction loop for Windows PowerShell.
# Copy this file, edit the steps below, and run it.
# The agent runs the script; the user follows prompts in their terminal.
#
# Usage:
#   powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\hitl-loop.template.ps1
#
# Two helpers:
#   Invoke-Step "<instruction>"     shows an instruction and waits for Enter
#   Read-Capture "<question>"       asks a question and returns the answer
#
# Captured values are printed as KEY=VALUE for the agent to parse.
# Capture observations only. Leave signing in and other secret input to the user
# through Invoke-Step so credentials never appear in captured output.

function Invoke-Step {
    param([Parameter(Mandatory)][string]$Instruction)

    Write-Host "`n>>> $Instruction"
    [void](Read-Host "    Press Enter when done")
}

function Read-Capture {
    param([Parameter(Mandatory)][string]$Question)

    Write-Host "`n>>> $Question"
    return (Read-Host "    >")
}

# --- edit below ---------------------------------------------------------

Invoke-Step "Open the app at http://localhost:3000 and sign in."

$errored = Read-Capture "Click the 'Export' button. Did it throw an error? (y/n)"

$errorMessage = Read-Capture "Paste the error message, or enter 'none':"

# --- edit above ---------------------------------------------------------

Write-Host "`n--- Captured ---"
Write-Output "ERRORED=$errored"
Write-Output "ERROR_MSG=$errorMessage"
