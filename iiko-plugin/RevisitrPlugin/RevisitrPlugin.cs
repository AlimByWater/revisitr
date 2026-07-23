using System;
using System.Collections.Generic;
using Resto.Front.Api;
using Resto.Front.Api.Attributes;
using Resto.Front.Api.Attributes.JetBrains;
using Resto.Front.Api.Exceptions;

namespace Revisitr.IikoPlugin
{
    /// <summary>
    /// Revisitr loyalty plugin for iikoFront. Registers a "Revisitr Бонусы"
    /// payment type: the cashier identifies the guest by a one-time word-code
    /// (typed or scanned as QR), redeems bonuses as payment, and accrues points
    /// on the remaining (money) part of the check.
    ///
    /// Deploy the build output to:
    ///   C:\Program Files\iiko\iikoRMS\Front.Net\Plugins\Resto.Front.Api.RevisitrPlugin.V8\
    /// alongside manifest.xml and revisitr.plugin.config.json (see the .example file).
    /// </summary>
    [UsedImplicitly]
    [PluginLicenseModuleId(21016318)] // iikoAPIPayment — loyalty systems (free ModuleID)
    public sealed class RevisitrPlugin : IFrontPlugin
    {
        private readonly List<IDisposable> disposables = new List<IDisposable>();
        private readonly RevisitrApiClient apiClient;

        public RevisitrPlugin()
        {
            var settings = PluginSettings.Load();
            apiClient = new RevisitrApiClient(settings);

            var processor = new RevisitrPaymentProcessor(apiClient);
            disposables.Add(processor);
            disposables.Add(new ClosedOrderReporter(apiClient));

            try
            {
                disposables.Add(PluginContext.Operations.RegisterPaymentSystem(processor));
            }
            catch (LicenseRestrictionException ex)
            {
                PluginContext.Log.Warn($"Revisitr: license restriction, payment system not registered: {ex.Message}");
                return;
            }
            catch (PaymentSystemRegistrationException ex)
            {
                PluginContext.Log.Warn($"Revisitr: payment system '{processor.PaymentSystemKey}' not registered: {ex.Message}");
                return;
            }

            PluginContext.Log.Info($"Revisitr payment system '{processor.PaymentSystemKey}' registered. Backend: {settings.BaseUrl}");
        }

        public void Dispose()
        {
            // Dispose in reverse order: unregister the payment system before the processor.
            for (var i = disposables.Count - 1; i >= 0; i--)
            {
                try { disposables[i].Dispose(); }
                catch (Exception ex) { PluginContext.Log.Warn($"Revisitr: dispose error: {ex.Message}"); }
            }
            apiClient?.Dispose();
        }
    }
}
