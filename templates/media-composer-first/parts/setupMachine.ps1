<#
    .SYNOPSIS
        Configure Windows 10 Workstation with Avid Media Composer First.

    .DESCRIPTION
        Configure Windows 10 Workstation with Avid Media Composer First.

        Example command line: .\setupMachine.ps1 Avid Media Composer First
#>
[CmdletBinding(DefaultParameterSetName="Standard")]
param(
    [string]
    [ValidateNotNullOrEmpty()]
    $MediaComposerURL
)

# the windows packages we want to remove
$global:AppxPkgs = @(
        "*windowscommunicationsapps*"
        "*windowsstore*"
        )

filter Timestamp {"$(Get-Date -Format o): $_"}

function
Write-Log($message)
{
    $msg = $message | Timestamp
    Write-Output $msg
}

function
DownloadFileOverHttp($Url, $DestinationPath)
{
     $secureProtocols = @()
     $insecureProtocols = @([System.Net.SecurityProtocolType]::SystemDefault, [System.Net.SecurityProtocolType]::Ssl3)

     foreach ($protocol in [System.Enum]::GetValues([System.Net.SecurityProtocolType]))
     {
         if ($insecureProtocols -notcontains $protocol)
         {
             $secureProtocols += $protocol
         }
     }
     [System.Net.ServicePointManager]::SecurityProtocol = $secureProtocols

    # make Invoke-WebRequest go fast: https://stackoverflow.com/questions/14202054/why-is-this-powershell-code-invoke-webrequest-getelementsbytagname-so-incred
    $ProgressPreference = "SilentlyContinue"
    Invoke-WebRequest $Url -UseBasicParsing -OutFile $DestinationPath -Verbose
    Write-Log "$DestinationPath updated"
}

function
Remove-WindowsApps($UserPath) 
{
    ForEach($app in $global:AppxPkgs){
        Get-AppxPackage -Name $app | Remove-AppxPackage -ErrorAction SilentlyContinue
    }
    try
    {
        ForEach($app in $global:AppxPkgs){
            Get-AppxPackage -Name $app | Remove-AppxPackage -User $UserPath -ErrorAction SilentlyContinue
        }
    }
    catch
    {
        # the user may not be created yet, but in case it is we want to remove the app
    }
    
    Remove-Item "c:\Users\Public\Desktop\Short_survey_to_provide_input_on_this_VM..url"
}

function
Install-DesktopLinks($UserPath) 
{
    #add a link to the desktop for Notepad++
    $wshshell = New-Object -ComObject WScript.Shell
    $lnk = $wshshell.CreateShortcut("c:\Users\$UserPath\Desktop\Notepad++.lnk")
    $lnk.TargetPath = "C:\Program Files\Notepad++\notepad++.exe"
    $lnk.Save()

    #add a link to the desktop for Media Composer First
    $wshshell = New-Object -ComObject WScript.Shell
    $lnk = $wshshell.CreateShortcut("c:\Users\$UserPath\Desktop\NewBlue Application Manager.lnk")
    $lnk.TargetPath = "C:\Program Files\NewBlueFX\Common\ApplicationManager64.exe"
    $lnk.Save()
}

function
Install-ChocolatyAndPackages
{
    Invoke-Expression ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))
    Write-Log "choco install -y 7zip.install"
    choco install -y 7zip.install
    Write-Log "choco install -y notepadplusplus.install"
    choco install -y notepadplusplus.install
}

function
Install-MediaComposerFirst
{
    Write-Log "downloading Media Composer First"
    # TODO: dynamically generate names based on download usrl
    $DestinationPath =  "C:\AzureData\NewBluePrime-170707.zip"
    DownloadFileOverHttp $MediaComposerURL $DestinationPath
    Write-Log "installing Media Composer First"
    # unzip media composer first
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    [System.IO.Compression.ZipFile]::ExtractToDirectory($DestinationPath, "C:\AzureData\")
    # use /S to silently install
    C:\AzureData\NewBluePrime-170707.exe /S
    # install media composer first
    Write-Log "finished installing Media Composer First"
}

try
{
    # Set to false for debugging.  This will output the start script to
    # c:\AzureData\CustomDataSetupScript.log, and then you can RDP
    # to the windows machine, and run the script manually to watch
    # the output.
    if ($true)
    {
        Write-Log("clean-up windows apps")
        Remove-WindowsApps $UserName

        try
        {
            Write-Log "Installing chocolaty and packages"
            Install-ChocolatyAndPackages
        }
        catch
        {
            # chocolaty is best effort
        }

        Write-Log "Install Media Composer First"
        Install-MediaComposerFirst

        Write-Log "Writing Desktop Links"
        Install-DesktopLinks "Default"
        
        Write-Log "Complete"
    }
    else
    {
        # keep for debugging purposes
        Write-Log "Set-ExecutionPolicy -ExecutionPolicy Unrestricted"
        Write-Log ".\CustomDataSetupScript.ps1 -MediaComposerURL $MediaComposerURL"
    }
}
catch
{
    Write-Error $_
}