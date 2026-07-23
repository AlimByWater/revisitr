using System;
using System.Collections.Generic;
using System.Globalization;
using System.Net;
using System.Net.Http;
using System.Text;
using System.Web.Script.Serialization;

namespace Revisitr.IikoPlugin
{
    /// <summary>
    /// Thrown for non-2xx backend responses. Carries the HTTP status so the
    /// payment processor can map it to a cashier-facing message.
    /// </summary>
    public sealed class ApiException : Exception
    {
        public int StatusCode { get; }

        public ApiException(int statusCode, string message) : base(message)
        {
            StatusCode = statusCode;
        }
    }

    public sealed class IdentifyResult
    {
        public string Session;
        public int ClientId;
        public string Name;
        public decimal Balance;
        public string Currency;
        public decimal AvailableToRedeem;
    }

    public sealed class OpResult
    {
        public decimal BalanceAfter;
        public decimal Accrued;
    }

    /// <summary>
    /// HTTP client for the Revisitr POS-plugin API (see
    /// docs/integrations/iiko/PLUGIN_CONTRACT.md). Uses only framework types —
    /// HttpClient for transport and JavaScriptSerializer for JSON — so nothing
    /// extra is shipped alongside the plugin.
    /// </summary>
    public sealed class RevisitrApiClient : IDisposable
    {
        private readonly HttpClient http;
        private readonly JavaScriptSerializer json = new JavaScriptSerializer();
        // The payment processor and closed-order reporter use one client concurrently.
        private readonly object jsonSync = new object();

        public RevisitrApiClient(PluginSettings settings)
        {
            // Older .NET Framework defaults may not enable TLS 1.2.
            ServicePointManager.SecurityProtocol |= SecurityProtocolType.Tls12;

            http = new HttpClient
            {
                Timeout = TimeSpan.FromSeconds(settings.TimeoutSeconds > 0 ? settings.TimeoutSeconds : 6)
            };
            if (!string.IsNullOrEmpty(settings.BaseUrl))
                http.BaseAddress = new Uri(settings.BaseUrl.TrimEnd('/') + "/");
            if (!string.IsNullOrEmpty(settings.ApiKey))
                http.DefaultRequestHeaders.Add("X-API-Key", settings.ApiKey);
        }

        /// <summary>Consumes a one-time code and opens a checkout session.</summary>
        public IdentifyResult Identify(string code, decimal orderTotal)
        {
            var resp = Post("api/v1/pos-plugin/identify", new Dictionary<string, object>
            {
                ["code"] = code,
                ["order_total"] = orderTotal
            });

            var client = resp.ContainsKey("client") ? resp["client"] as Dictionary<string, object> : null;
            return new IdentifyResult
            {
                Session = Str(resp, "session"),
                ClientId = client != null ? Int(client, "id") : 0,
                Name = client != null ? Str(client, "name") : "",
                Balance = client != null ? Dec(client, "balance") : 0m,
                Currency = client != null ? Str(client, "currency") : "",
                AvailableToRedeem = client != null ? Dec(client, "available_to_redeem") : 0m
            };
        }

        /// <summary>Spends bonuses. Idempotent per (integration, order_id).</summary>
        public OpResult Redeem(string session, string orderId, decimal amount)
        {
            var resp = Post("api/v1/pos-plugin/redeem", new Dictionary<string, object>
            {
                ["session"] = session,
                ["order_id"] = orderId,
                ["amount"] = amount
            });
            return new OpResult { BalanceAfter = Dec(resp, "balance_after"), Accrued = Dec(resp, "accrued") };
        }

        /// <summary>Accrues points on a check amount. Idempotent per (integration, order_id).</summary>
        public OpResult Accrue(string session, string orderId, decimal amount)
        {
            var resp = Post("api/v1/pos-plugin/accrue", new Dictionary<string, object>
            {
                ["session"] = session,
                ["order_id"] = orderId,
                ["amount"] = amount
            });
            return new OpResult { BalanceAfter = Dec(resp, "balance_after"), Accrued = Dec(resp, "accrued") };
        }

        /// <summary>Best-effort report of an order closed in iikoFront.</summary>
        public void ReportClosedOrder(Dictionary<string, object> order)
        {
            Post("api/v1/pos-plugin/order", order);
        }

        private Dictionary<string, object> Post(string path, Dictionary<string, object> body)
        {
            string payload;
            lock (jsonSync)
                payload = json.Serialize(body);
            using (var content = new StringContent(payload, Encoding.UTF8, "application/json"))
            using (var resp = http.PostAsync(path, content).GetAwaiter().GetResult())
            {
                var text = resp.Content.ReadAsStringAsync().GetAwaiter().GetResult();
                if (!resp.IsSuccessStatusCode)
                    throw new ApiException((int)resp.StatusCode, ExtractError(text));
                return Parse(text);
            }
        }

        private Dictionary<string, object> Parse(string text)
        {
            if (string.IsNullOrEmpty(text))
                return new Dictionary<string, object>();
            lock (jsonSync)
                return json.Deserialize<Dictionary<string, object>>(text) ?? new Dictionary<string, object>();
        }

        private string ExtractError(string text)
        {
            try
            {
                var d = Parse(text);
                if (d.ContainsKey("error") && d["error"] != null)
                    return Convert.ToString(d["error"], CultureInfo.InvariantCulture);
            }
            catch
            {
                // fall through to raw body
            }
            return string.IsNullOrEmpty(text) ? "unknown error" : text;
        }

        private static string Str(Dictionary<string, object> d, string k)
            => d.ContainsKey(k) && d[k] != null ? Convert.ToString(d[k], CultureInfo.InvariantCulture) : "";

        private static int Int(Dictionary<string, object> d, string k)
            => d.ContainsKey(k) && d[k] != null ? Convert.ToInt32(d[k], CultureInfo.InvariantCulture) : 0;

        private static decimal Dec(Dictionary<string, object> d, string k)
            => d.ContainsKey(k) && d[k] != null ? Convert.ToDecimal(d[k], CultureInfo.InvariantCulture) : 0m;

        public void Dispose() => http?.Dispose();
    }
}
