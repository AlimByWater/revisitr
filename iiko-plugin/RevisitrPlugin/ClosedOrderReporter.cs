using System;
using System.Collections.Generic;
using System.Globalization;
using System.Linq;
using System.Reactive.Linq;
using System.Threading;
using Resto.Front.Api;
using Resto.Front.Api.Data.Orders;

namespace Revisitr.IikoPlugin
{
    /// <summary>
    /// Reports orders once when they become closed on this cash terminal.
    /// Notification handling is deliberately limited to snapshotting and queuing:
    /// a backend outage must never delay the iikoFront cashier workflow.
    /// </summary>
    public sealed class ClosedOrderReporter : IDisposable
    {
        private readonly IDisposable subscription;
        private readonly RevisitrApiClient api;
        private readonly Guid terminalId;
        private readonly HashSet<Guid> reportedOrderIds;
        private readonly object sync = new object();

        public ClosedOrderReporter(RevisitrApiClient api)
        {
            this.api = api;
            terminalId = PluginContext.Operations.GetHostTerminal().Id;
            reportedOrderIds = new HashSet<Guid>(PluginContext.Operations.GetOrders(true)
                .Where(order => order.Status == OrderStatus.Closed)
                .Select(order => order.Id));

            subscription = PluginContext.Notifications.OrderChanged.Subscribe(notification =>
            {
                var order = notification.Entity;
                if (order == null || order.Status != OrderStatus.Closed || order.LastChangedTerminalId != terminalId)
                    return;

                try
                {
                    // Build the payload before queueing: notification entities must not cross threads.
                    var payload = Serialize(order);
                    lock (sync)
                    {
                        if (!reportedOrderIds.Add(order.Id))
                            return;
                    }
                    ThreadPool.QueueUserWorkItem(_ => Send(payload));
                }
                catch (Exception ex)
                {
                    PluginContext.Log.Warn($"Revisitr: failed to queue closed order {order.Id}: {ex.Message}");
                }
            });
        }

        private void Send(Dictionary<string, object> payload)
        {
            var orderId = Convert.ToString(payload["order_id"], CultureInfo.InvariantCulture);
            try
            {
                api.ReportClosedOrder(payload);
                PluginContext.Log.Info($"Revisitr: closed order {orderId} sent.");
            }
            catch (Exception ex)
            {
                // Delivery is best-effort. The order has already been closed in iikoFront.
                PluginContext.Log.Warn($"Revisitr: closed order {orderId} was not sent: {ex.Message}");
            }
        }

        private static Dictionary<string, object> Serialize(IOrder order)
        {
            return new Dictionary<string, object>
            {
                ["order_id"] = order.Id.ToString(),
                ["source"] = order is IDeliveryOrder ? "delivery" : "hall",
                ["ordered_at"] = order.CloseTime.HasValue
                    ? order.CloseTime.Value.ToUniversalTime().ToString("o", CultureInfo.InvariantCulture)
                    : "",
                ["total"] = order.ResultSum,
                ["table_num"] = string.Join(", ", order.Tables.Where(table => table != null).Select(table => table.Name)),
                ["waiter_name"] = order.Waiter != null ? order.Waiter.Name : "",
                ["items"] = order.Items
                    .Where(item => !item.Deleted)
                    .Select(SerializeItem)
                    .Where(item => item != null)
                    .ToList()
            };
        }

        private static Dictionary<string, object> SerializeItem(IOrderRootItem item)
        {
            var product = item as IOrderProductItem;
            if (product != null)
            {
                return new Dictionary<string, object>
                {
                    ["name"] = string.IsNullOrEmpty(product.ProductCustomName) ? product.Product.Name : product.ProductCustomName,
                    ["quantity"] = product.Amount,
                    ["price"] = product.Amount == 0m ? product.ResultSum : product.ResultSum / product.Amount
                };
            }

            var compound = item as IOrderCompoundItem;
            if (compound != null)
            {
                return new Dictionary<string, object>
                {
                    ["name"] = compound.PrimaryComponent.Product.Name,
                    ["quantity"] = 1m,
                    ["price"] = compound.PrimaryComponent.ResultSum
                };
            }

            return null;
        }

        public void Dispose()
        {
            subscription?.Dispose();
        }
    }
}
