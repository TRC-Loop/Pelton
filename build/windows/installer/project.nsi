Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them with the values from ProjectInfo.
## If they are defined here, "wails_tools.nsh" will not touch them. This allows to use this project.nsi manually
## from outside of Wails for debugging and development of the installer.
##
## For development first make a wails nsis build to populate the "wails_tools.nsh":
## > wails build --target windows/amd64 --nsis
## Then you can call makensis on this file with specifying the path to your binary:
## For a AMD64 only installer:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app.exe
## For a ARM64 only installer:
## > makensis -DARG_WAILS_ARM64_BINARY=..\..\bin\app.exe
## For a installer with both architectures:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe
####
## The following information is taken from the ProjectInfo file, but they can be overwritten here.
####
## !define INFO_PROJECTNAME    "MyProject" # Default "{{.Name}}"
## !define INFO_COMPANYNAME    "MyCompany" # Default "{{.Info.CompanyName}}"
## !define INFO_PRODUCTNAME    "MyProduct" # Default "{{.Info.ProductName}}"
## !define INFO_PRODUCTVERSION "1.0.0"     # Default "{{.Info.ProductVersion}}"
## !define INFO_COPYRIGHT      "Copyright" # Default "{{.Info.Copyright}}"
###
## !define PRODUCT_EXECUTABLE  "Application.exe"      # Default "${INFO_PROJECTNAME}.exe"
## !define UNINST_KEY_NAME     "UninstKeyInRegistry"  # Default "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
####
## !define REQUEST_EXECUTION_LEVEL "admin"            # Default "admin"  see also https://nsis.sourceforge.io/Docs/Chapter4.html
####
## MultiUser (per-machine vs per-user install choice) needs "Highest" instead
## of the default "admin" so the installer can run un-elevated and offer to
## elevate only if the user picks the all-users option. Must be set before
## wails_tools.nsh is included, since it only defines REQUEST_EXECUTION_LEVEL
## if nothing has already.
####
!define REQUEST_EXECUTION_LEVEL "Highest"
####
## Include the wails tools
####
!include "wails_tools.nsh"

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
# !define MUI_WELCOMEFINISHPAGE_BITMAP "resources\leftimage.bmp" #Include this to add a bitmap on the left side of the Welcome Page. Must be a size of 164x314
!define MUI_FINISHPAGE_NOAUTOCLOSE # Wait on the INSTFILES page so the user can take a look into the details of the installation steps
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.

####
## Per-machine (all users, needs admin) vs per-user (no admin) install choice.
## Defaults preserve the old behavior (all users) unless the user picks the
## other option on the new page this adds.
####
!define MULTIUSER_EXECUTIONLEVEL Highest
!define MULTIUSER_MUI
!define MULTIUSER_INSTALLMODE_COMMANDLINE
!define MULTIUSER_INSTALLMODE_DEFAULT_REGISTRY_KEY "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
!define MULTIUSER_INSTALLMODE_INSTDIR "${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
!include "MultiUser.nsh"

!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
!insertmacro MUI_PAGE_LICENSE "..\..\..\LICENSE" # GPL-3.0. Kept in the original English text on every install
                                                  # language: unofficial translations of the GPL aren't legally
                                                  # authoritative, so the FSF recommends distributing only this text.
!insertmacro MULTIUSER_PAGE_INSTALLMODE # All users (admin) vs just me (no admin).
!insertmacro MUI_PAGE_DIRECTORY # In which folder install page.
!insertmacro MUI_PAGE_COMPONENTS # Optional components (currently just the desktop shortcut).
!insertmacro MUI_PAGE_INSTFILES # Installing page.
!define MUI_FINISHPAGE_RUN "$INSTDIR\${PRODUCT_EXECUTABLE}" # Offer to launch Pelton right after install, checked by default.
!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_INSTFILES # Uinstalling page

# Installer languages, kept in lockstep with Pelton's own UI languages
# (frontend/src/lib/locales). English first so it's the dialog's default. This
# only controls the installer's own UI language; Pelton itself always ships
# with every language embedded (they're a few KB each), so there's nothing to
# select or exclude there.
!insertmacro MUI_LANGUAGE "English"
!insertmacro MUI_LANGUAGE "German"
!insertmacro MUI_LANGUAGE "French"
!insertmacro MUI_LANGUAGE "Dutch"
!insertmacro MUI_LANGUAGE "Spanish"

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe" # Name of the installer's file.
InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" # Fallback default; MULTIUSER_INIT overrides this based on the page above.
ShowInstDetails show # This will always show the installation details.

# wails.setShellContext (from wails_tools.nsh) picks the shell context from the
# REQUEST_EXECUTION_LEVEL *string define*, which is now "Highest" to support
# both install modes - so that macro would always fall through to the
# per-user branch. This reads the actual runtime choice from MultiUser instead.
!macro pelton.setShellContext
    ${If} $MultiUser.InstallMode == "AllUsers"
        SetShellVarContext all
    ${Else}
        SetShellVarContext current
    ${EndIf}
!macroend

Function .onInit
   !insertmacro MULTIUSER_INIT # reads/handles the all-users vs per-user choice first; may relaunch elevated.
   !insertmacro MUI_LANGDLL_DISPLAY # ask which of the installer languages above to use
   !insertmacro wails.checkArchitecture
FunctionEnd

Function un.onInit
   !insertmacro MULTIUSER_UNINIT # recovers the install mode this copy was installed with.
   !insertmacro MUI_UNGETLANGUAGE
FunctionEnd

Section "-Core" SecCore
    ; leading "-" hides this from the components list and keeps it mandatory:
    ; the app itself, its Start Menu shortcut, and file/protocol associations
    ; always install regardless of what's checked below.
    !insertmacro pelton.setShellContext

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro wails.writeUninstaller
SectionEnd

Section "Desktop Shortcut" SecDesktopShortcut
    !insertmacro pelton.setShellContext
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
SectionEnd

Section "uninstall"
    !insertmacro pelton.setShellContext

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
SectionEnd
