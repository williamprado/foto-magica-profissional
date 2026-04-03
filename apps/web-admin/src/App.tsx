import { AppShell, Badge, Button, Card } from "@foto-magica/ui";

const tenantRows = [
  { name: "Demo Studio", plan: "growth", jobs: 28, credits: 87, status: "active" },
  { name: "Retratos Sul", plan: "starter", jobs: 6, credits: 4, status: "attention" },
  { name: "Equipe Prime", plan: "scale", jobs: 112, credits: 190, status: "active" }
];

const providerRows = [
  { name: "Google GenAI", scope: "analysis + prompt", health: "ok" },
  { name: "Vertex Imagen", scope: "image generation", health: "ok" },
  { name: "MinIO / S3", scope: "storage", health: "ok" },
  { name: "Stripe / Asaas / PayPal", scope: "billing", health: "degraded" }
];

export default function App() {
  return (
    <AppShell
      title="Superadmin Control Center"
      subtitle="Operação multi-tenant, billing, providers, filas e observabilidade da Foto Magica Profissional."
      actions={<Button variant="accent">Criar tenant manual</Button>}
    >
      <div className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <Card>
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-2xl font-black">Tenants ativos</h2>
              <p className="mt-2 text-slate-500">Gestão com isolamento, saldo e uso do pipeline.</p>
            </div>
            <Badge tone="accent">3 tenants</Badge>
          </div>
          <div className="mt-6 overflow-hidden rounded-[24px] border border-line">
            <table className="min-w-full divide-y divide-line text-left text-sm">
              <thead className="bg-fog">
                <tr>
                  <th className="px-4 py-3">Tenant</th>
                  <th className="px-4 py-3">Plano</th>
                  <th className="px-4 py-3">Jobs</th>
                  <th className="px-4 py-3">Créditos</th>
                  <th className="px-4 py-3">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-line">
                {tenantRows.map((tenant) => (
                  <tr key={tenant.name}>
                    <td className="px-4 py-4 font-semibold">{tenant.name}</td>
                    <td className="px-4 py-4">{tenant.plan}</td>
                    <td className="px-4 py-4">{tenant.jobs}</td>
                    <td className="px-4 py-4">{tenant.credits}</td>
                    <td className="px-4 py-4">
                      <Badge tone={tenant.status === "attention" ? "danger" : "accent"}>{tenant.status}</Badge>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>

        <div className="space-y-6">
          <Card>
            <h2 className="text-2xl font-black">Saúde dos providers</h2>
            <div className="mt-6 space-y-3">
              {providerRows.map((provider) => (
                <div key={provider.name} className="flex items-center justify-between rounded-[22px] bg-fog px-4 py-3">
                  <div>
                    <p className="font-semibold">{provider.name}</p>
                    <p className="text-sm text-slate-500">{provider.scope}</p>
                  </div>
                  <Badge tone={provider.health === "ok" ? "accent" : "danger"}>{provider.health}</Badge>
                </div>
              ))}
            </div>
          </Card>

          <Card>
            <h2 className="text-2xl font-black">Filas e auditoria</h2>
            <div className="mt-6 grid gap-4 md:grid-cols-2">
              <div className="rounded-[22px] bg-ink p-5 text-white">
                <p className="text-sm text-slate-300">Jobs processando</p>
                <p className="mt-2 text-4xl font-black">12</p>
              </div>
              <div className="rounded-[22px] bg-accent p-5 text-ink">
                <p className="text-sm font-semibold">Webhooks hoje</p>
                <p className="mt-2 text-4xl font-black">34</p>
              </div>
              <div className="rounded-[22px] bg-fog p-5">
                <p className="text-sm text-slate-500">Falhas com retry</p>
                <p className="mt-2 text-4xl font-black">2</p>
              </div>
              <div className="rounded-[22px] bg-fog p-5">
                <p className="text-sm text-slate-500">Receita do mês</p>
                <p className="mt-2 text-4xl font-black">R$ 1.490</p>
              </div>
            </div>
          </Card>
        </div>
      </div>
    </AppShell>
  );
}

