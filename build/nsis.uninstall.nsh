Section "Uninstall"
  # uninstall for all users
  setShellVarContext all

  # Delete (optionally) installed files
  {{range $}}Delete $INSTDIR\{{.}}
  {{end}}
  Delete $INSTDIR\uninstall.exe

  # Delete install directory
  rmDir $INSTDIR

  # Delete start menu launcher
  Delete "$SMPROGRAMS\${APPNAME}\${APPNAME}.lnk"
  Delete "$SMPROGRAMS\${APPNAME}\Attach.lnk"
  Delete "$SMPROGRAMS\${APPNAME}\Uninstall.lnk"
  rmDir "$SMPROGRAMS\${APPNAME}"

  # Firewall - remove rules if exists
  SimpleFC::AdvRemoveRule "Gwsh incoming peers (TCP:30373)"
  SimpleFC::AdvRemoveRule "Gwsh outgoing peers (TCP:30373)"
  SimpleFC::AdvRemoveRule "Gwsh UDP discovery (UDP:30373)"

  # Remove IPC endpoint (https://github.com/wiseplat/EIPs/issues/147)
  ${un.EnvVarUpdate} $0 "WISEPLAT_SOCKET" "R" "HKLM" "\\.\pipe\gwsh.ipc"

  # Remove install directory from PATH
  Push "$INSTDIR"
  Call un.RemoveFromPath

  # Cleanup registry (deletes all sub keys)
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${GROUPNAME} ${APPNAME}"
SectionEnd
