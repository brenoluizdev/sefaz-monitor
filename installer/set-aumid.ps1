# Executado uma vez pelo instalador (ver [Run] em sefaz-monitor.iss).
#
# Sem isso, notificações do Windows (toast) para um app Win32 "não
# empacotado" como este são aceitas pela API mas nunca renderizadas — o
# Windows exige um atalho no Menu Iniciar com a propriedade
# System.AppUserModel.ID (AUMID) preenchida. Ver:
# https://learn.microsoft.com/windows/win32/shell/enable-desktop-toast-with-appusermodelid
param(
    [Parameter(Mandatory = $true)][string]$ShortcutPath,
    [Parameter(Mandatory = $true)][string]$TargetPath,
    [Parameter(Mandatory = $true)][string]$AppId
)

Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;
using System.Runtime.InteropServices.ComTypes;
using System.Text;

namespace AumidHelper
{
    [ComImport, Guid("00021401-0000-0000-C000-000000000046")]
    internal class CShellLink { }

    [ComImport, InterfaceType(ComInterfaceType.InterfaceIsIUnknown), Guid("000214F9-0000-0000-C000-000000000046")]
    internal interface IShellLinkW
    {
        void GetPath(StringBuilder pszFile, int cchMaxPath, IntPtr pfd, uint fFlags);
        void GetIDList(out IntPtr ppidl);
        void SetIDList(IntPtr pidl);
        void GetDescription(StringBuilder pszName, int cchMaxName);
        void SetDescription(string pszName);
        void GetWorkingDirectory(StringBuilder pszDir, int cchMaxPath);
        void SetWorkingDirectory(string pszDir);
        void GetArguments(StringBuilder pszArgs, int cchMaxPath);
        void SetArguments(string pszArgs);
        void GetHotkey(out short pwHotkey);
        void SetHotkey(short wHotkey);
        void GetShowCmd(out int piShowCmd);
        void SetShowCmd(int iShowCmd);
        void GetIconLocation(StringBuilder pszIconPath, int cchIconPath, out int piIcon);
        void SetIconLocation(string pszIconPath, int iIcon);
        void SetRelativePath(string pszPathRel, uint dwReserved);
        void Resolve(IntPtr hwnd, uint fFlags);
        void SetPath(string pszFile);
    }

    [StructLayout(LayoutKind.Sequential, Pack = 4)]
    internal struct PropertyKey
    {
        public Guid fmtid;
        public int pid;
        public PropertyKey(string fmtid, int pid) { this.fmtid = new Guid(fmtid); this.pid = pid; }
    }

    [StructLayout(LayoutKind.Explicit)]
    internal sealed class PropVariant : IDisposable
    {
        [FieldOffset(0)] ushort vt;
        [FieldOffset(8)] IntPtr ptr;

        public PropVariant(string value)
        {
            vt = (ushort)VarEnum.VT_LPWSTR;
            ptr = Marshal.StringToCoTaskMemUni(value);
        }

        public void Dispose()
        {
            PropVariantClear(this);
            GC.SuppressFinalize(this);
        }

        [DllImport("Ole32.dll", PreserveSig = false)]
        private static extern void PropVariantClear([In, Out] PropVariant pvar);
    }

    [ComImport, InterfaceType(ComInterfaceType.InterfaceIsIUnknown), Guid("886D8EEB-8CF2-4446-8D02-CDBA1DBDCF99")]
    internal interface IPropertyStore
    {
        uint GetCount(out uint cProps);
        uint GetAt(uint iProp, out PropertyKey pkey);
        uint GetValue(ref PropertyKey key, out PropVariant pv);
        uint SetValue(ref PropertyKey key, PropVariant pv);
        uint Commit();
    }

    public static class ShortcutAumid
    {
        public static void Set(string shortcutPath, string targetPath, string appId)
        {
            var link = (IShellLinkW)new CShellLink();
            link.SetPath(targetPath);

            var store = (IPropertyStore)link;
            var key = new PropertyKey("9F4C2855-9F79-4B39-A8D0-E1D42DE1D5F3", 5); // PKEY_AppUserModel_ID
            using (var pv = new PropVariant(appId))
            {
                store.SetValue(ref key, pv);
                store.Commit();
            }

            var persistFile = (IPersistFile)link;
            persistFile.Save(shortcutPath, true);
        }
    }
}
'@

[AumidHelper.ShortcutAumid]::Set($ShortcutPath, $TargetPath, $AppId)
