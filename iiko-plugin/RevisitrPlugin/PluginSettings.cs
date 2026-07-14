using System;
using System.IO;
using System.Reflection;
using System.Web.Script.Serialization;
using Resto.Front.Api;

namespace Revisitr.IikoPlugin
{
    /// <summary>
    /// Plugin configuration, read once at startup from revisitr.plugin.config.json
    /// located next to the plugin DLL. Keys are matched case-insensitively.
    /// </summary>
    public sealed class PluginSettings
    {
        public string BaseUrl { get; set; }
        public string ApiKey { get; set; }
        public int TimeoutSeconds { get; set; } = 6;

        private const string FileName = "revisitr.plugin.config.json";

        public static PluginSettings Load()
        {
            var dir = Path.GetDirectoryName(Assembly.GetExecutingAssembly().Location) ?? ".";
            var path = Path.Combine(dir, FileName);

            if (!File.Exists(path))
            {
                PluginContext.Log.Warn($"Revisitr: config not found at {path}. API calls will fail until it is created.");
                return new PluginSettings();
            }

            try
            {
                var json = File.ReadAllText(path);
                var s = new JavaScriptSerializer().Deserialize<PluginSettings>(json) ?? new PluginSettings();
                if (s.TimeoutSeconds <= 0)
                    s.TimeoutSeconds = 6;
                return s;
            }
            catch (Exception ex)
            {
                PluginContext.Log.Error($"Revisitr: failed to read config {path}: {ex.Message}");
                return new PluginSettings();
            }
        }
    }
}
