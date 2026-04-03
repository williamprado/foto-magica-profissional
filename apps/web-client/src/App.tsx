import { useMemo, useState } from "react";
import { NavLink, Route, Routes } from "react-router-dom";
import { createApiClient } from "@foto-magica/api-client";
import type { GalleryItem, Plan, PromptSection } from "@foto-magica/types";
import { useAsyncAction } from "@foto-magica/hooks";
import { AppShell, Badge, Button, Card } from "@foto-magica/ui";

const api = createApiClient();

const defaultPrompt: PromptSection[] = [
  { key: "style", title: "STYLE", content: "Retrato corporativo premium com iluminação suave, visual limpo e composição confiante." },
  { key: "pose", title: "POSE", content: "Enquadramento vertical profissional, expressão segura, foco no rosto e tronco." },
  { key: "textures", title: "TEXTURES", content: "Blazer texturizado, fundo de estúdio com profundidade e acabamento refinado." }
];

function NavBar() {
  const links = [
    { to: "/", label: "Gerar Foto" },
    { to: "/galeria", label: "Galeria" },
    { to: "/creditos", label: "Créditos" },
    { to: "/perfil", label: "Perfil" }
  ];

  return (
    <header className="border-b border-line bg-white/95 backdrop-blur">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4 lg:px-10">
        <div className="text-3xl font-black tracking-tight">FotoMagica</div>
        <nav className="flex items-center gap-3">
          {links.map((link) => (
            <NavLink
              key={link.to}
              to={link.to}
              className={({ isActive }) =>
                `rounded-full px-4 py-2 text-sm font-semibold ${isActive ? "bg-ink text-white" : "text-slate-500 hover:text-ink"}`
              }
            >
              {link.label}
            </NavLink>
          ))}
        </nav>
        <div className="flex items-center gap-3">
          <Badge tone="accent">3 créditos</Badge>
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-fog font-semibold">W</div>
        </div>
      </div>
    </header>
  );
}

function AuthPage() {
  const [mode, setMode] = useState<"login" | "register">("register");
  const { execute, loading, error } = useAsyncAction(async () => {
    if (mode === "login") {
      return api.login({ email: "admin@fotomagica.app", password: "Passw0rd!", tenantSlug: "demo-studio" });
    }
    return api.register({
      companyName: "Demo Studio",
      companySlug: "demo-studio",
      fullName: "William Prado",
      email: "admin@fotomagica.app",
      password: "Passw0rd!"
    });
  });

  return (
    <div className="grid min-h-screen place-items-center bg-fog px-6 py-12">
      <Card className="w-full max-w-xl space-y-6 p-8">
        <div className="text-center">
          <p className="text-5xl font-black tracking-tight">FotoMagica</p>
          <p className="mt-4 text-slate-500">Plataforma multi-tenant para retratos profissionais com IA.</p>
        </div>
        <div className="grid grid-cols-2 gap-3 rounded-full bg-fog p-1">
          <button className={`rounded-full px-4 py-3 text-sm font-semibold ${mode === "register" ? "bg-white shadow-soft" : ""}`} onClick={() => setMode("register")}>Criar conta</button>
          <button className={`rounded-full px-4 py-3 text-sm font-semibold ${mode === "login" ? "bg-white shadow-soft" : ""}`} onClick={() => setMode("login")}>Entrar</button>
        </div>
        <div className="grid gap-4">
          <input className="rounded-full border-line" placeholder="Empresa / tenant slug" defaultValue="demo-studio" />
          <input className="rounded-full border-line" placeholder="Seu nome" defaultValue="William Prado" />
          <input className="rounded-full border-line" placeholder="Email" defaultValue="admin@fotomagica.app" />
          <input className="rounded-full border-line" type="password" placeholder="Senha" defaultValue="Passw0rd!" />
        </div>
        {error ? <p className="text-sm text-danger">{error}</p> : null}
        <Button className="w-full" onClick={() => void execute()} disabled={loading}>
          {loading ? "Processando..." : mode === "login" ? "Entrar na plataforma" : "Criar ambiente"}
        </Button>
      </Card>
    </div>
  );
}

function GenerationPage() {
  const [promptSections, setPromptSections] = useState(defaultPrompt);
  const [jobs, setJobs] = useState([
    { id: "job-001", status: "processing", progress: 62, costCredits: 1, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString(), promptSections }
  ]);

  const updatePrompt = (key: string, value: string) => {
    setPromptSections((current) => current.map((section) => (section.key === key ? { ...section, content: value } : section)));
  };

  return (
    <AppShell
      title="Envie a referência e gere sua foto"
      subtitle="Fluxo pensado para análise visual, prompt estruturado, fila de geração e entrega final com créditos."
      actions={<Button variant="accent">Nova sessão</Button>}
    >
      <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
        <Card className="space-y-6">
          <div>
            <h2 className="text-3xl font-black">1. Referência visual</h2>
            <p className="mt-2 text-slate-500">
              Escolha a imagem que define estilo, iluminação e composição para as próximas gerações.
            </p>
          </div>
          <div className="grid gap-6 lg:grid-cols-[0.95fr_1.05fr]">
            <label className="grid min-h-72 place-items-center rounded-[28px] border border-dashed border-slate-300 bg-fog text-center">
              <div>
                <p className="text-lg font-semibold">Arraste ou selecione sua referência</p>
                <p className="mt-2 text-sm text-slate-500">PNG, JPG ou WEBP até 8MB</p>
              </div>
            </label>
            <div className="rounded-[28px] bg-[radial-gradient(circle_at_top,#3d4a5e_0,#111827_65%,#05070a_100%)] p-6 text-white">
              <p className="text-xs uppercase tracking-[0.3em] text-slate-300">Análise IA</p>
              <p className="mt-4 text-2xl font-black">Direção visual executiva</p>
              <p className="mt-3 text-sm leading-6 text-slate-200">
                Fundo escuro com textura suave, blazer estruturado, pose frontal confiante e enquadramento de alto impacto.
              </p>
              <div className="mt-6 h-2 rounded-full bg-white/10">
                <div className="h-2 w-3/4 rounded-full bg-accent" />
              </div>
            </div>
          </div>
        </Card>

        <Card className="space-y-5">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-2xl font-black">2. Prompt Gerado</h2>
              <p className="mt-1 text-sm text-slate-500">Edite antes de enviar a sua foto.</p>
            </div>
            <Badge tone="accent">1 crédito</Badge>
          </div>
          {promptSections.map((section) => (
            <div key={section.key} className="rounded-[24px] border border-line p-4">
              <p className="text-xs font-black uppercase tracking-[0.3em] text-slate-500">{section.title}</p>
              <textarea
                className="mt-3 min-h-24 w-full rounded-2xl border-line bg-fog text-sm"
                value={section.content}
                onChange={(event) => updatePrompt(section.key, event.target.value)}
              />
            </div>
          ))}
          <Button className="w-full">Enviar foto do usuário e gerar</Button>
        </Card>
      </div>

      <div className="mt-8 grid gap-6 lg:grid-cols-[0.85fr_1.15fr]">
        <Card>
          <h3 className="text-2xl font-black">3. Upload da foto do usuário</h3>
          <div className="mt-5 grid min-h-80 place-items-center rounded-[28px] border border-dashed border-slate-300 bg-fog text-center">
            <div>
              <p className="text-lg font-semibold">Selfie ou retrato original</p>
              <p className="mt-2 max-w-sm text-sm text-slate-500">
                O sistema combina a sua foto com a referência e o prompt estruturado para criar o resultado final.
              </p>
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-2xl font-black">4. Pipeline em andamento</h3>
              <p className="mt-1 text-sm text-slate-500">Status da fila, retries e ativo final.</p>
            </div>
            <Badge tone="default">Worker + Queue</Badge>
          </div>
          <div className="mt-5 space-y-4">
            {jobs.map((job) => (
              <div key={job.id} className="rounded-[24px] border border-line p-5">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-semibold">{job.id}</p>
                    <p className="text-sm text-slate-500">Status: {job.status}</p>
                  </div>
                  <Badge tone={job.status === "failed" ? "danger" : "accent"}>{job.progress}%</Badge>
                </div>
                <div className="mt-4 h-3 rounded-full bg-fog">
                  <div className="h-3 rounded-full bg-ink" style={{ width: `${job.progress}%` }} />
                </div>
              </div>
            ))}
            <div className="grid gap-4 md:grid-cols-2">
              <div className="rounded-[24px] bg-fog p-5">
                <p className="text-sm text-slate-500">Créditos consumidos</p>
                <p className="mt-2 text-3xl font-black">1</p>
              </div>
              <div className="rounded-[24px] bg-[linear-gradient(135deg,#0f172a,#1e293b)] p-5 text-white">
                <p className="text-sm text-slate-300">Saída final</p>
                <p className="mt-2 text-3xl font-black">Disponível na galeria</p>
              </div>
            </div>
          </div>
        </Card>
      </div>
    </AppShell>
  );
}

function GalleryPage() {
  const items: GalleryItem[] = [
    {
      id: "asset-001",
      title: "Retrato executivo premium",
      previewUrl: "",
      favorite: true,
      createdAt: "2026-04-02T18:00:00Z"
    }
  ];

  return (
    <AppShell title="Sua galeria" subtitle="Todos os resultados processados pelo pipeline com busca, favoritos e histórico.">
      <div className="mb-6 flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <input className="w-full rounded-full border-line md:max-w-lg" placeholder="Buscar por prompt, status ou data" />
        <Button variant="accent">Nova Foto</Button>
      </div>
      <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
        {items.map((item) => (
          <Card key={item.id} className="overflow-hidden p-0">
            <div className="aspect-[4/5] bg-[radial-gradient(circle_at_top,#4b5563_0,#111827_65%,#05070a_100%)]" />
            <div className="p-5">
              <div className="flex items-center justify-between">
                <p className="font-bold">{item.title}</p>
                <Badge tone={item.favorite ? "accent" : "default"}>{item.favorite ? "Favorita" : "Nova"}</Badge>
              </div>
              <p className="mt-2 text-sm text-slate-500">{new Date(item.createdAt).toLocaleDateString("pt-BR")}</p>
            </div>
          </Card>
        ))}
      </div>
    </AppShell>
  );
}

function CreditsPage() {
  const plans: Plan[] = [
    { id: "starter", code: "starter", name: "Starter", creditAmount: 20, priceCents: 990, currency: "BRL", active: true },
    { id: "growth", code: "growth", name: "Growth", creditAmount: 50, priceCents: 1990, currency: "BRL", active: true },
    { id: "scale", code: "scale", name: "Scale", creditAmount: 120, priceCents: 3990, currency: "BRL", active: true }
  ];

  return (
    <AppShell title="Gerenciar créditos" subtitle="Planos recarregáveis com histórico, auditoria e billing multi-provider.">
      <div className="mb-8 flex items-center justify-center">
        <div className="rounded-full bg-accent px-8 py-4 text-ink shadow-soft">
          <p className="text-sm font-medium uppercase tracking-[0.3em]">Saldo atual</p>
          <p className="text-center text-4xl font-black">3 créditos</p>
        </div>
      </div>
      <div className="grid gap-5 lg:grid-cols-3">
        {plans.map((plan) => (
          <Card key={plan.id}>
            <Badge tone={plan.code === "growth" ? "accent" : "default"}>{plan.code === "growth" ? "Mais vendido" : "Plano"}</Badge>
            <p className="mt-4 text-2xl font-black">{plan.name}</p>
            <p className="mt-2 text-slate-500">{plan.creditAmount} créditos para geração com IA</p>
            <p className="mt-6 text-4xl font-black">
              {(plan.priceCents / 100).toLocaleString("pt-BR", { style: "currency", currency: plan.currency })}
            </p>
            <Button className="mt-6 w-full" variant={plan.code === "growth" ? "accent" : "primary"}>
              Comprar com checkout seguro
            </Button>
          </Card>
        ))}
      </div>
    </AppShell>
  );
}

function ProfilePage() {
  return (
    <AppShell title="Perfil do tenant" subtitle="Informações da conta, estatísticas, gestão da equipe e configurações seguras.">
      <div className="grid gap-6 lg:grid-cols-[1.15fr_0.85fr]">
        <Card className="space-y-5">
          <h2 className="text-2xl font-black">Informações pessoais</h2>
          <div className="grid gap-4 md:grid-cols-2">
            <input className="rounded-full border-line" defaultValue="William Prado" />
            <input className="rounded-full border-line" defaultValue="admin@fotomagica.app" />
            <input className="rounded-full border-line md:col-span-2" defaultValue="demo-studio" />
          </div>
          <Button>Salvar alterações</Button>
        </Card>
        <div className="space-y-6">
          <Card>
            <p className="text-sm text-slate-500">Estatísticas</p>
            <div className="mt-4 grid grid-cols-3 gap-4">
              <div>
                <p className="text-xs uppercase tracking-[0.2em] text-slate-400">Créditos</p>
                <p className="mt-2 text-3xl font-black">3</p>
              </div>
              <div>
                <p className="text-xs uppercase tracking-[0.2em] text-slate-400">Fotos</p>
                <p className="mt-2 text-3xl font-black">18</p>
              </div>
              <div>
                <p className="text-xs uppercase tracking-[0.2em] text-slate-400">Equipe</p>
                <p className="mt-2 text-3xl font-black">4</p>
              </div>
            </div>
          </Card>
          <Card className="border-red-100">
            <p className="text-sm font-semibold text-danger">Zona de risco</p>
            <p className="mt-3 text-sm text-slate-500">Encerramento de conta, exportação de dados e auditoria de acessos.</p>
            <Button variant="secondary" className="mt-5 border-danger text-danger hover:bg-red-50">Excluir tenant</Button>
          </Card>
        </div>
      </div>
    </AppShell>
  );
}

export default function App() {
  const [authenticated] = useState(true);
  const content = useMemo(
    () => (
      <>
        <NavBar />
        <Routes>
          <Route path="/" element={<GenerationPage />} />
          <Route path="/galeria" element={<GalleryPage />} />
          <Route path="/creditos" element={<CreditsPage />} />
          <Route path="/perfil" element={<ProfilePage />} />
        </Routes>
      </>
    ),
    []
  );

  return authenticated ? content : <AuthPage />;
}

