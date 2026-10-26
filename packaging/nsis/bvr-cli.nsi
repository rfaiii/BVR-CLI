; NSIS installer script for BVR CLI on Windows x64.
;
; Built per-user (RequestExecutionLevel user) into $LOCALAPPDATA\Programs so it
; matches the winget manifest's InstallationMetadata and ElevationRequirement:
; the previous artifact for this slot was the bare binary, which winget would
; have "installed" by running the TUI.
;
; scripts/package/build-installers.sh passes VERSION, BINARY, README, LICENSE
; and OUTFILE as -D defines; nothing here is generated.

!include "MUI2.nsh"
!include "LogicLib.nsh"
!include "StrFunc.nsh"
!include "FileFunc.nsh"
!include "WinMessages.nsh"
!include "WinVer.nsh"

${Using:StrFunc} StrStr

!define APP_NAME    "BVR CLI"
!define APP_PUB     "richavery"
!define APP_ID      "richavery.BvrCLI"
!define MONIKER     "bvr-cli"
!define REG_HW      "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_ID}"
!define REG_USER    "Software\Environment"

Name "${APP_NAME}"
OutFile "${OUTFILE}"
InstallDir "$LOCALAPPDATA\Programs\${MONIKER}"
InstallDirRegKey HKCU "${REG_HW}" "InstallLocation"
RequestExecutionLevel user
SetCompressor /SOLID lzma
Unicode True
VIProductVersion "${VERSION}.0"
VIAddVersionKey /LANG=1033 "ProductName" "${APP_NAME}"
VIAddVersionKey /LANG=1033 "FileVersion" "${VERSION}"
VIAddVersionKey /LANG=1033 "ProductVersion" "${VERSION}"
VIAddVersionKey /LANG=1033 "LegalCopyright" "Copyright (c) ${APP_PUB}"
VIAddVersionKey /LANG=1033 "FileDescription" "${APP_NAME} installer"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

Section "Install"
  SetOutPath "$INSTDIR"
  File /r "${BINARY}"
  File /r "${README}"
  File /r "${LICENSE}"
  ; ARP entry: winget reads these to decide whether the package is present and
  ; how to remove it, so an installer without one cannot be upgraded.
  WriteRegStr HKCU "${REG_HW}" "DisplayName" "${APP_NAME}"
  WriteRegStr HKCU "${REG_HW}" "DisplayVersion" "${VERSION}"
  WriteRegStr HKCU "${REG_HW}" "Publisher" "${APP_PUB}"
  WriteRegStr HKCU "${REG_HW}" "URLInfoAbout" "https://github.com/rfaiii/BVR-CLI"
  WriteRegStr HKCU "${REG_HW}" "InstallLocation" "$INSTDIR"
  WriteRegStr HKCU "${REG_HW}" "UninstallString" '"$INSTDIR\uninstall.exe"'
  WriteRegStr HKCU "${REG_HW}" "QuietUninstallString" '"$INSTDIR\uninstall.exe" /S'
  WriteRegDWORD HKCU "${REG_HW}" "NoModify" 1
  WriteRegDWORD HKCU "${REG_HW}" "NoRepair" 1
  ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
  IntFmt $0 "0x%08X" $0
  WriteRegDWORD HKCU "${REG_HW}" "EstimatedSize" $0

  WriteUninstaller "$INSTDIR\uninstall.exe"
SectionEnd

Section "-AppendToUserPath"
  ; HKCU\Environment\Path is REG_EXPAND_SZ and must stay that way; writing it
  ; back as REG_SZ silently breaks any %USERPROFILE% style entries already in
  ; it. Read-modify-write, and only when the dir is genuinely absent.
  ReadRegStr $0 HKCU "${REG_USER}" "Path"
  ${If} $0 == ""
    WriteRegExpandStr HKCU "${REG_USER}" "Path" "$INSTDIR"
  ${Else}
    ${StrStr} $1 "$0\;" "$INSTDIR\;"
    ${If} $1 == ""
      WriteRegExpandStr HKCU "${REG_USER}" "Path" "$0;$INSTDIR"
    ${EndIf}
  ${EndIf}
  SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment" /TIMEOUT=2000
SectionEnd

Section "Uninstall"
  Delete "$INSTDIR\${MONIKER}.exe"
  Delete "$INSTDIR\README.*"
  Delete "$INSTDIR\LICENSE.*"
  Delete "$INSTDIR\uninstall.exe"
  RMDir "$INSTDIR"

  ; Removing one element from a semicolon list without a string-replace plugin
  ; means rebuilding the whole value, and a bug there writes the user's PATH.
  ; So: clear it only when ours was the entire value, otherwise leave the stale
  ; entry and say so rather than risk the user's environment.
  ReadRegStr $0 HKCU "${REG_USER}" "Path"
  ${If} $0 == "$INSTDIR"
    DeleteRegValue HKCU "${REG_USER}" "Path"
    SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment" /TIMEOUT=2000
  ${ElseIf} $0 != ""
    MessageBox MB_OK|MB_ICONINFORMATION "BVR CLI was removed, but $INSTDIR is still in your user PATH.$\r$\n$\r$\nRemove it with: System Properties -> Environment Variables." /SD IDOK
  ${EndIf}

  DeleteRegKey HKCU "${REG_HW}"
SectionEnd
