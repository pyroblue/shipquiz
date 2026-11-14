import { CopyButton } from "@/components/copy-button";

/** A labeled form. The button copies the typed text. */
export function CodeBlock({
  code,
  label,
}: {
  code: string;
  label?: string;
}) {
  return (
    <div className="max-w-full min-w-0 overflow-hidden border border-line bg-card text-ink">
      {label ? (
        <div className="bg-ink px-4 py-2">
          <span className="text-[11px] font-bold tracking-[0.07em] text-white uppercase">{label}</span>
        </div>
      ) : null}
      <div className="flex items-start gap-3 px-4 py-3">
        <pre className="min-w-0 flex-1 overflow-x-auto font-mono text-sm leading-relaxed">
          <code>{code}</code>
        </pre>
        <CopyButton text={code} />
      </div>
    </div>
  );
}
