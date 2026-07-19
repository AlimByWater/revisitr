using System;
using System.Threading;
using System.Web.Script.Serialization;
using Resto.Front.Api;
using Resto.Front.Api.Data.Cheques;
using Resto.Front.Api.Data.Orders;
using Resto.Front.Api.Data.Organization;
using Resto.Front.Api.Data.Payments;
using Resto.Front.Api.Data.Security;
using Resto.Front.Api.Data.View;
using Resto.Front.Api.Exceptions;
using Resto.Front.Api.UI;

namespace Revisitr.IikoPlugin
{
    /// <summary>
    /// "Revisitr Бонусы" payment type.
    ///
    /// Flow:
    ///   CollectData  → cashier enters/scans the guest's one-time code → identify()
    ///                  → store session + available-to-redeem in rollback data.
    ///   OnPaymentAdded → cap the payable-with-bonuses sum to available.
    ///   Pay          → redeem(sum) [if sum &gt; 0], then accrue on the money part.
    ///
    /// The plugin never blocks the cashier: redeem failure aborts THIS payment
    /// (bonuses untouched); accrue failure is logged and swallowed (order already paid).
    /// </summary>
    public sealed class RevisitrPaymentProcessor : IPaymentProcessor, IDisposable
    {
        private const string Key = "revisitr";
        private const string Name = "Revisitr Бонусы";

        // V8 rollback data is a plain string, so SessionData is carried as JSON.
        private static readonly JavaScriptSerializer Json = new JavaScriptSerializer();

        private readonly RevisitrApiClient api;

        public RevisitrPaymentProcessor(RevisitrApiClient api)
        {
            this.api = api;
        }

        public string PaymentSystemKey => Key;
        public string PaymentSystemName => Name;

        // Executes when the plugin payment item is about to be added to the order.
        public void CollectData(
            Guid orderId,
            Guid paymentTypeId,
            IUser cashier,
            IReceiptPrinter printer,
            IViewManager viewManager,
            IPaymentDataContext context)
        {
            var order = PluginContext.Operations.GetOrderById(orderId);
            var orderTotal = order != null ? order.ResultSum : 0m;

            // One dialog covers both identification channels: type the word OR scan its QR.
            var input = viewManager.ShowExtendedKeyboardDialog(
                "Код гостя Revisitr — введите слово или отсканируйте QR", enableBarcode: true);
            if (input == null)
                throw new PaymentActionCancelledException();

            string code = null;
            if (input is StringInputDialogResult typed)
                code = typed.Result;
            else if (input is BarcodeInputDialogResult scanned)
                code = scanned.Barcode;

            code = code?.Trim();
            if (string.IsNullOrEmpty(code))
                throw new PaymentActionFailedException("Код не введён.");

            viewManager.ChangeProgressBarMessage("Проверяем код…");

            IdentifyResult id;
            try
            {
                // No retry on identify: the code is one-time and consumed server-side.
                id = api.Identify(code, orderTotal);
            }
            catch (ApiException ex)
            {
                throw new PaymentActionFailedException(Friendly(ex));
            }
            catch (Exception ex)
            {
                PluginContext.Log.Error($"Revisitr identify failed: {ex}");
                throw new PaymentActionFailedException("Сервис лояльности недоступен. Закройте заказ обычным способом.");
            }

            context.SetRollbackData(Json.Serialize(new SessionData
            {
                Session = id.Session,
                Available = id.AvailableToRedeem,
                Balance = id.Balance,
                Currency = id.Currency,
                ClientName = id.Name,
                OrderTotal = orderTotal
            }));

            viewManager.ShowOkPopup("Revisitr",
                $"Гость: {id.Name}\n" +
                $"Баланс: {id.Balance:0.##} {id.Currency}\n" +
                $"Доступно к списанию: {id.AvailableToRedeem:0.##} {id.Currency}");
        }

        // Executes after the plugin payment item is added.
        // MVP: no client-side sum capping — the backend rejects redeem > available
        // (400), and the cashier already sees "available to redeem" in the popup.
        // Capping the numpad range is a UX follow-up (needs the V8 sum-limit API).
        public void OnPaymentAdded(
            IOrder order,
            IPaymentItem paymentItem,
            IUser cashier,
            IOperationService operations,
            IReceiptPrinter printer,
            IViewManager viewManager,
            IPaymentDataContext context)
        {
        }

        public bool OnPreliminaryPaymentEditing(
            IOrder order,
            IPaymentItem paymentItem,
            IUser cashier,
            IOperationService operationService,
            IReceiptPrinter printer,
            IViewManager viewManager,
            IPaymentDataContext context)
        {
            // Allow the standard numpad for editing the preliminary payment sum.
            return true;
        }

        // Executes when the order is paid and contains our payment item.
        public void Pay(
            decimal sum,
            IOrder order,
            IPaymentItem paymentItem,
            Guid transactionId,
            IPointOfSale pointOfSale,
            IUser cashier,
            IOperationService operations,
            IReceiptPrinter printer,
            IViewManager viewManager,
            IPaymentDataContext context)
        {
            var raw = context.GetRollbackData();
            var sd = string.IsNullOrEmpty(raw) ? null : Json.Deserialize<SessionData>(raw);
            if (sd == null || string.IsNullOrEmpty(sd.Session))
                throw new PaymentActionFailedException("Сессия Revisitr не найдена. Удалите оплату и повторите ввод кода.");

            var orderId = order.Id.ToString();

            // 1) Redeem bonuses (skip if the cashier chose accrue-only with sum 0).
            if (sum > 0m)
            {
                viewManager?.ChangeProgressBarMessage("Списываем бонусы…");
                try
                {
                    WithRetry(() => api.Redeem(sd.Session, orderId, sum));
                }
                catch (ApiException ex)
                {
                    throw new PaymentActionFailedException(Friendly(ex));
                }
                catch (Exception ex)
                {
                    PluginContext.Log.Error($"Revisitr redeem failed: {ex}");
                    throw new PaymentActionFailedException("Сервис лояльности недоступен. Оплата бонусами невозможна.");
                }
            }

            // 2) Accrue on the money part (best-effort; never blocks the closed payment).
            var accrualBase = order.ResultSum - sum;
            if (accrualBase > 0m)
            {
                try
                {
                    var res = WithRetry(() => api.Accrue(sd.Session, orderId, accrualBase));
                    PluginContext.Log.Info($"Revisitr accrued {res.Accrued}, balance {res.BalanceAfter} (order {orderId})");
                }
                catch (Exception ex)
                {
                    PluginContext.Log.Warn($"Revisitr accrue failed (non-blocking, order {orderId}): {ex.Message}");
                }
            }
        }

        // --- Storno / cancel: MVP does NOT restore spent bonuses (see README "Known limitations"). ---

        public void ReturnPayment(
            decimal sum, Guid? orderId, Guid paymentTypeId, Guid transactionId,
            IPointOfSale pointOfSale, IUser cashier, IReceiptPrinter printer,
            IViewManager viewManager, IPaymentDataContext context)
        {
            PluginContext.Log.Warn($"Revisitr: ReturnPayment for order {orderId} — bonuses are NOT restored in MVP.");
        }

        public void ReturnPaymentSilently(
            decimal sum, Guid? orderId, Guid paymentTypeId, Guid transactionId,
            IPointOfSale pointOfSale, IUser cashier, IReceiptPrinter printer,
            IPaymentDataContext context)
        {
            PluginContext.Log.Warn($"Revisitr: ReturnPaymentSilently for order {orderId} — bonuses are NOT restored in MVP.");
        }

        public void EmergencyCancelPayment(
            decimal sum, Guid? orderId, Guid paymentTypeId, Guid transactionId,
            IPointOfSale pointOfSale, IUser cashier, IReceiptPrinter printer,
            IViewManager viewManager, IPaymentDataContext context)
        {
            PluginContext.Log.Warn($"Revisitr: EmergencyCancelPayment for order {orderId} — bonuses are NOT restored in MVP.");
        }

        public void EmergencyCancelPaymentSilently(
            decimal sum, Guid? orderId, Guid paymentTypeId, Guid transactionId,
            IPointOfSale pointOfSale, IUser cashier, IReceiptPrinter printer,
            IPaymentDataContext context)
        {
            PluginContext.Log.Warn($"Revisitr: EmergencyCancelPaymentSilently for order {orderId} — bonuses are NOT restored in MVP.");
        }

        public void ReturnPaymentWithoutOrder(
            decimal sum, Guid? orderId, Guid paymentTypeId,
            IPointOfSale pointOfSale, IUser cashier, IReceiptPrinter printer, IViewManager viewManager)
        {
            throw new PaymentActionFailedException("Возврат без заказа не поддерживается.");
        }

        // Revisitr always needs the identify dialog, so silent pay is not supported.
        public bool CanPaySilently(decimal sum, Guid? orderId, Guid paymentTypeId, IPaymentDataContext context) => false;

        public void PaySilently(
            decimal sum, IOrder order, IPaymentItem paymentItem, Guid transactionId,
            IPointOfSale pointOfSale, IUser cashier, IReceiptPrinter printer, IPaymentDataContext context)
        {
            throw new PaymentActionFailedException("Тихая оплата Revisitr не поддерживается.");
        }

        public void OnPaymentDeleting(
            IOrder order, IPaymentItem paymentItem, IUser cashier, IOperationService operationService,
            IReceiptPrinter printer, IViewManager viewManager, IPaymentDataContext context)
        {
            // Allow deletion of a not-yet-paid item; nothing to undo before Pay.
            PluginContext.Log.Info($"Revisitr: payment item removed from order {order.Id} before payment.");
        }

        private static T WithRetry<T>(Func<T> action)
        {
            Exception last = null;
            for (var attempt = 0; attempt < 2; attempt++)
            {
                try
                {
                    return action();
                }
                catch (ApiException)
                {
                    // Business errors (404/409/400/…) are deterministic — do not retry.
                    throw;
                }
                catch (Exception ex)
                {
                    // Network/timeout: retry is safe because redeem/accrue are idempotent per order_id.
                    last = ex;
                    Thread.Sleep(300);
                }
            }
            throw last;
        }

        private static string Friendly(ApiException ex)
        {
            switch (ex.StatusCode)
            {
                case 404: return "Код не найден или истёк. Попросите гостя обновить код.";
                case 401: return "Сессия истекла. Повторите ввод кода.";
                case 409: return "Недостаточно бонусов на балансе.";
                case 400: return "Некорректная сумма списания.";
                case 429: return "Слишком много попыток. Подождите минуту.";
                case 403: return "Плагин не настроен (неверный ключ). Обратитесь к администратору.";
                default: return "Сервис лояльности недоступен.";
            }
        }

        public void Dispose()
        {
            // No owned resources: the API client is owned by RevisitrPlugin.
        }
    }
}
