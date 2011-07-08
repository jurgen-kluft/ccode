using System.Runtime.InteropServices;

namespace xcode
{
    internal static class SafeNativeMethods
    {
        [DllImport("user32.dll", SetLastError = true)]
        internal static extern int GetSystemMetrics(SystemMetric smIndex);
    }
}
