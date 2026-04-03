insert into plans (code, name, description, credit_amount, price_cents, currency, payment_provider, active)
values
  ('starter', 'Starter', 'Pacote inicial para pequenos volumes', 20, 990, 'BRL', 'mock', true),
  ('growth', 'Growth', 'Plano principal para uso recorrente', 50, 1990, 'BRL', 'mock', true),
  ('scale', 'Scale', 'Pacote para maior volume e times', 120, 3990, 'BRL', 'mock', true)
on conflict (code) do update set
  name = excluded.name,
  description = excluded.description,
  credit_amount = excluded.credit_amount,
  price_cents = excluded.price_cents,
  currency = excluded.currency,
  payment_provider = excluded.payment_provider,
  active = excluded.active;

