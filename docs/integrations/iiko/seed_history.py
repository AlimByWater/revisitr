#!/usr/bin/env python3
"""Generate ~3 months of realistic café history for the Revisitr iiko integration.

iiko Cloud cannot backdate orders, so this seeds Revisitr's DB directly:
  - bot_clients (loyalty members, bot 6 / org 3)
  - external_orders (integration 2), dated 2026-03-24 .. 2026-06-23
  - integration_aggregates recomputed from external_orders (daily, for charts)
  - bot_clients RFM/visit fields recomputed from their orders

Narrative: a specialty coffee café that launched its Revisitr loyalty program
~3 months ago. Order volume grows over time; weekends/Friday peak; morning coffee
rush + lunch bowls + evening desserts; product mix shifts to cold drinks toward
June; a few events (May holidays, a promo weekend, a rainy slow week). Loyalty
adoption rises, so the matched/loyalty share grows over the period.

Output: SQL on stdout. Reproducible (fixed seed). Re-runnable: deletes prior
demo3m rows first. Real iiko-synced orders (non-demo3m external_ids) are left
untouched; today's bucket stays as the real synced data.

Usage:
  python3 seed_history.py | docker exec -i infra-postgres-1 psql -U revisitr -d revisitr
"""
import json
import random
from datetime import date, timedelta

random.seed(20260624)

INTEGRATION_ID = 2
BOT_ID = 6                # Baratie (org 3)
TG_BASE = 9_000_000_000   # synthetic telegram_ids, no collision
START = date(2026, 3, 24)
END = date(2026, 6, 23)   # inclusive; today (06-24) left as real iiko data
DAYS = (END - START).days + 1

# Real products from the stand: (id, name, price, kind)
# kind: hotcoffee | coldcoffee | pastry | food | drink_hot | drink_cold
P = [
    ("7a35c826-7314-9d1c-019e-5ec4c505f71e", "Revisitr Beans",      450, "retail"),
    ("7a35c826-7314-9d1c-019e-5ec4c505f765", "Revisitr Drip",       320, "hotcoffee"),
    ("7a35c826-7314-9d1c-019e-5ec4c505f767", "Revisitr Cookie",     150, "pastry"),
    ("7a35c826-7314-9d1c-019e-5ec4c505fc1c", "Revisitr Espresso",   190, "hotcoffee"),
    ("7a35c826-7314-9d1c-019e-5ec4c505fc1e", "Revisitr Cappuccino", 260, "hotcoffee"),
    ("7a35c826-7314-9d1c-019e-5ec4c505fc20", "Revisitr Raf",        310, "hotcoffee"),
    ("7a35c826-7314-9d1c-019e-5ec4c505fc22", "Revisitr Cold Brew",  280, "coldcoffee"),
    ("7a35c826-7314-9d1c-019e-5ec4c505fc24", "Revisitr Croissant",  220, "pastry"),
    ("7a35c826-7314-9d1c-019e-5ec4c505fc28", "Revisitr Brownie",    240, "pastry"),
    ("7a35c826-7314-9d1c-019e-5ec4c505fc2a", "Revisitr Cheesecake", 290, "pastry"),
    ("7a35c826-7314-9d1c-019e-5ec4c505fc2c", "Revisitr Granola",    330, "food"),
    ("7a35c826-7314-9d1c-019e-5ec4c505fc2f", "Revisitr Salad",      360, "food"),
    ("7a35c826-7314-9d1c-019e-5ec4c505fc31", "Revisitr Chicken Bowl", 420, "food"),
    ("7a35c826-7314-9d1c-019e-5ec4c505fc33", "Revisitr Soup",       270, "food"),
    ("7a35c826-7314-9d1c-019e-5ec4c505fc35", "Revisitr Lemonade",   210, "drink_cold"),
    ("7a35c826-7314-9d1c-019e-5ec4c505fc37", "Revisitr Matcha",     300, "drink_cold"),
    ("7a35c826-7314-9d1c-019e-5ec4c505fc39", "Revisitr Tea",        170, "drink_hot"),
]
BY_KIND = {}
for pid, name, price, kind in P:
    BY_KIND.setdefault(kind, []).append((pid, name, price))
CAT = {"hotcoffee": "Coffee", "coldcoffee": "Coffee", "drink_hot": "Drinks",
       "drink_cold": "Drinks", "pastry": "Bakery", "food": "Kitchen", "retail": "Retail"}


def season(d):
    """0.0 at START (cold weather) -> 1.0 at END (warm). Shifts hot<->cold mix."""
    return (d - START).days / max(DAYS - 1, 1)


def pick(kind):
    return random.choice(BY_KIND[kind])


def coffee(d):
    # warmer -> more cold coffee
    return pick("coldcoffee") if random.random() < 0.15 + 0.45 * season(d) else pick("hotcoffee")


def drink(d):
    if random.random() < 0.2 + 0.5 * season(d):
        return pick("drink_cold")
    return pick("drink_hot")


def basket(daypart, d):
    """Return list of (pid,name,price,qty) for a realistic order."""
    items = []
    if daypart == "morning":
        items.append(coffee(d))
        if random.random() < 0.6:
            items.append(pick("pastry"))
        if random.random() < 0.12:
            items.append(coffee(d))  # second coffee (to-go pair)
    elif daypart == "lunch":
        if random.random() < 0.8:
            items.append(pick("food"))
        items.append(drink(d) if random.random() < 0.5 else coffee(d))
        if random.random() < 0.2:
            items.append(pick("pastry"))
    else:  # evening
        if random.random() < 0.55:
            items.append(coffee(d))
        else:
            items.append(drink(d))
        if random.random() < 0.5:
            items.append(pick("pastry"))
        if random.random() < 0.08:
            items.append(("7a35c826-7314-9d1c-019e-5ec4c505f71e", "Revisitr Beans", 450))  # retail bag
    if not items:
        items.append(coffee(d))
    # collapse to qty
    agg = {}
    for it in items:
        pid, name, price = it[0], it[1], it[2]
        if pid in agg:
            agg[pid][3] += 1
        else:
            agg[pid] = [pid, name, price, 1]
    return list(agg.values())


# ---- Customers ----
FIRST = ["Анна", "Дмитрий", "Мария", "Сергей", "Елена", "Алексей", "Ольга", "Иван",
         "Наталья", "Павел", "Екатерина", "Андрей", "Татьяна", "Михаил", "Юлия",
         "Артём", "Светлана", "Никита", "Дарья", "Роман", "Ксения", "Владимир",
         "Полина", "Денис", "Алина", "Максим", "Виктория", "Игорь", "София", "Глеб",
         "Вероника", "Кирилл", "Альбина", "Степан", "Марина", "Тимур", "Жанна",
         "Олег", "Лидия", "Григорий"]
LAST = ["Смирнов", "Иванова", "Кузнецов", "Попова", "Соколов", "Лебедева", "Новиков",
        "Морозова", "Волков", "Орлова", "Васильев", "Павлова", "Семёнов", "Голубева",
        "Виноградов", "Зайцева", "Петров", "Беляева", "Комаров", "Киселёва"]


def phone(i):
    return f"+7916{2000000 + i:07d}"


customers = []
n_vip, n_reg, n_occ = 8, 16, 20
for i in range(n_vip + n_reg + n_occ):
    fn = FIRST[i % len(FIRST)]
    ln = random.choice(LAST)
    if i < n_vip:
        seg, vpw, reg = "vip", random.uniform(2.5, 4.0), True
        join = START + timedelta(days=random.randint(0, 20))
    elif i < n_vip + n_reg:
        seg, vpw, reg = "regular", random.uniform(0.7, 1.6), random.random() < 0.85
        join = START + timedelta(days=random.randint(0, 55))
    else:
        seg, vpw, reg = "occasional", random.uniform(0.1, 0.45), random.random() < 0.4
        join = START + timedelta(days=random.randint(0, DAYS - 5))
    customers.append(dict(idx=i, fn=fn, ln=ln, phone=phone(i), seg=seg,
                          vpw=vpw, reg=reg, join=join, tg=TG_BASE + i))

registered = [c for c in customers if c["reg"]]

WEEKDAY = {0: 0.75, 1: 0.9, 2: 0.95, 3: 1.0, 4: 1.35, 5: 1.4, 6: 1.05}  # Mon..Sun
EVENTS = {}
for dd in [date(2026, 5, 1), date(2026, 5, 2), date(2026, 5, 3)]:
    EVENTS[dd] = 1.25                      # May holidays — strollers
EVENTS[date(2026, 5, 9)] = 1.6             # Victory Day spike
for dd in [date(2026, 5, 16), date(2026, 5, 17)]:
    EVENTS[dd] = 1.5                        # "free cookie" promo weekend
for k in range(7):                          # rainy slow week in April
    EVENTS[date(2026, 4, 7) + timedelta(days=k)] = 0.7


def day_volume(d, t):
    base = 7 + 24 * (t / (DAYS - 1)) ** 1.15        # growth trend
    f = WEEKDAY[d.weekday()] * EVENTS.get(d, 1.0)
    f *= random.gauss(1.0, 0.12)
    return max(2, round(base * f))


print("BEGIN;")
print("SET TIME ZONE 'UTC';")
print(f"DELETE FROM external_orders WHERE integration_id={INTEGRATION_ID} AND external_id LIKE 'demo3m-%';")
print(f"DELETE FROM bot_clients WHERE bot_id={BOT_ID} AND data->>'seed'='demo3m';")

# Insert registered clients (RFM fields recomputed later)
for c in registered:
    print(
        "INSERT INTO bot_clients (bot_id, telegram_id, username, first_name, last_name, "
        "phone, phone_normalized, registered_at, city, data) VALUES "
        f"({BOT_ID}, {c['tg']}, NULL, '{c['fn']}', '{c['ln']}', '{c['phone']}', "
        f"'{c['phone']}', '{c['join']} 08:00:00+00', 'Москва', '{{\"seed\":\"demo3m\"}}') "
        "ON CONFLICT (bot_id, telegram_id) DO NOTHING;"
    )

DAYPARTS = [("morning", 0.45, 5, 3), ("lunch", 0.30, 9, 3), ("evening", 0.25, 14, 4)]
# (name, prob, H0, HSPAN): hour = H0 + rand(0..HSPAN), UTC → ≈ 08:00-21:00 MSK

oid = 0
for t in range(DAYS):
    d = START + timedelta(days=t)
    n = day_volume(d, t)
    p_loyal = 0.32 + 0.33 * (t / (DAYS - 1))     # loyalty adoption grows
    active = [c for c in registered if c["join"] <= d]
    weights = [c["vpw"] for c in active]
    for _ in range(n):
        r = random.random()
        dp = DAYPARTS[0][0]
        acc = 0.0
        for name, p, h0, hspan in DAYPARTS:
            acc += p
            if r <= acc:
                dp, H0, HSPAN = name, h0, hspan
                break
        hour = H0 + random.randint(0, HSPAN)
        minute = random.randint(0, 59)
        ts = f"{d} {hour:02d}:{minute:02d}:{random.randint(0,59):02d}+00"

        items = basket(dp, d)
        total = sum(price * qty for _, _, price, qty in items)
        items_json = json.dumps(
            [{"external_id": pid, "name": nm, "quantity": qty, "price": price,
              "category": CAT.get(next(k for k, v in BY_KIND.items()
                                       if any(x[0] == pid for x in v)), "Coffee")}
             for pid, nm, price, qty in items],
            ensure_ascii=False,
        ).replace("'", "''")

        oid += 1
        ext = f"demo3m-{oid:05d}"
        cust = None
        if active and random.random() < p_loyal:
            cust = random.choices(active, weights=weights, k=1)[0]
            ph = f"'{cust['phone']}'"
            cid = (f"(SELECT id FROM bot_clients WHERE bot_id={BOT_ID} "
                   f"AND phone='{cust['phone']}' LIMIT 1)")
        elif random.random() < 0.7:
            # walk-in with a one-off phone (unmatched)
            ph = f"'+7916{8000000 + random.randint(0, 999999):07d}'"
            cid = "NULL"
        else:
            ph = "NULL"      # anonymous walk-in
            cid = "NULL"
        print(
            "INSERT INTO external_orders (integration_id, external_id, client_id, "
            "customer_phone, items, total, ordered_at, synced_at) VALUES "
            f"({INTEGRATION_ID}, '{ext}', {cid}, {ph}, '{items_json}'::jsonb, "
            f"{total}, '{ts}', NOW());"
        )

# Recompute daily aggregates from ALL of this integration's orders (charts).
print(f"DELETE FROM integration_aggregates WHERE integration_id={INTEGRATION_ID};")
print(f"""INSERT INTO integration_aggregates
  (integration_id, date, revenue, avg_check, tx_count, guest_count, matched_count, synced_at)
SELECT {INTEGRATION_ID}, date(ordered_at), SUM(total),
       CASE WHEN COUNT(*)>0 THEN ROUND(SUM(total)/COUNT(*),2) ELSE 0 END,
       COUNT(*), COUNT(DISTINCT customer_phone),
       COUNT(*) FILTER (WHERE client_id IS NOT NULL), NOW()
FROM external_orders WHERE integration_id={INTEGRATION_ID}
GROUP BY date(ordered_at);""")

# Recompute client visit/RFM fields from their linked orders.
print(f"""UPDATE bot_clients bc SET
  total_visits_lifetime = s.cnt,
  frequency_count = s.cnt,
  monetary_sum = s.spend,
  last_visit_date = s.last_day,
  recency_days = (DATE '2026-06-24' - s.last_day),
  m_score = LEAST(5, GREATEST(1, (s.spend/2000)::int + 1)),
  f_score = LEAST(5, GREATEST(1, (s.cnt/6)::int + 1)),
  r_score = CASE WHEN (DATE '2026-06-24' - s.last_day) <= 7 THEN 5
                 WHEN (DATE '2026-06-24' - s.last_day) <= 21 THEN 4
                 WHEN (DATE '2026-06-24' - s.last_day) <= 45 THEN 3
                 WHEN (DATE '2026-06-24' - s.last_day) <= 70 THEN 2 ELSE 1 END,
  rfm_segment = CASE
     WHEN s.cnt >= 30 AND (DATE '2026-06-24' - s.last_day) <= 14 THEN 'Champions'
     WHEN s.cnt >= 12 AND (DATE '2026-06-24' - s.last_day) <= 21 THEN 'Loyal'
     WHEN (DATE '2026-06-24' - s.last_day) <= 14 THEN 'Recent'
     WHEN (DATE '2026-06-24' - s.last_day) >= 45 THEN 'At Risk'
     ELSE 'Regular' END,
  rfm_updated_at = NOW()
FROM (
  SELECT client_id, COUNT(*) cnt, SUM(total) spend, MAX(date(ordered_at)) last_day
  FROM external_orders WHERE integration_id={INTEGRATION_ID} AND client_id IS NOT NULL
  GROUP BY client_id
) s WHERE bc.id = s.client_id;""")

print("COMMIT;")
