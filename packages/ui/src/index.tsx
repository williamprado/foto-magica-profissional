import type { PropsWithChildren, ReactNode } from "react";

export function cn(...parts: Array<string | false | null | undefined>) {
  return parts.filter(Boolean).join(" ");
}

export function AppShell({
  title,
  subtitle,
  actions,
  children
}: PropsWithChildren<{
  title: string;
  subtitle?: string;
  actions?: ReactNode;
}>) {
  return (
    <div className="min-h-screen bg-white text-ink">
      <div className="mx-auto max-w-7xl px-6 py-8 lg:px-10">
        <div className="mb-8 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h1 className="text-4xl font-black tracking-tight">{title}</h1>
            {subtitle ? <p className="mt-3 max-w-2xl text-lg text-slate-500">{subtitle}</p> : null}
          </div>
          {actions}
        </div>
        {children}
      </div>
    </div>
  );
}

export function Card({
  className,
  children
}: PropsWithChildren<{ className?: string }>) {
  return (
    <section className={cn("rounded-[28px] border border-line bg-white p-6 shadow-soft", className)}>
      {children}
    </section>
  );
}

export function Button({
  children,
  variant = "primary",
  className,
  ...props
}: PropsWithChildren<
  React.ButtonHTMLAttributes<HTMLButtonElement> & {
    variant?: "primary" | "secondary" | "ghost" | "accent";
    className?: string;
  }
>) {
  return (
    <button
      className={cn(
        "inline-flex h-12 items-center justify-center rounded-full px-6 text-sm font-semibold transition",
        variant === "primary" && "bg-ink text-white hover:bg-slate-800",
        variant === "secondary" && "border border-ink bg-white text-ink hover:bg-slate-50",
        variant === "ghost" && "bg-transparent text-slate-500 hover:text-ink",
        variant === "accent" && "bg-accent text-ink hover:opacity-90",
        className
      )}
      {...props}
    >
      {children}
    </button>
  );
}

export function Badge({
  children,
  tone = "default"
}: PropsWithChildren<{ tone?: "default" | "accent" | "danger" }>) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold",
        tone === "default" && "bg-slate-100 text-slate-700",
        tone === "accent" && "bg-accentSoft text-ink",
        tone === "danger" && "bg-red-50 text-danger"
      )}
    >
      {children}
    </span>
  );
}

