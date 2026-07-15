using System;

namespace Revisitr.IikoPlugin
{
    /// <summary>
    /// Data collected in CollectData and carried to Pay via the payment context's
    /// rollback data. Must be [Serializable] (iikoFront persists it between the
    /// payment-item collect and the actual pay step).
    /// </summary>
    [Serializable]
    public sealed class SessionData
    {
        public string Session;
        public decimal Available;
        public decimal Balance;
        public string Currency;
        public string ClientName;
        public decimal OrderTotal;
    }
}
