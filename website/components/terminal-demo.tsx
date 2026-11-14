"use client";

import { useEffect, useState } from "react";

type Line = {
  text: string;
  /** Type this line one character at a time. */
  typed?: boolean;
  /** Wait after the line is on screen, in milliseconds. */
  pause?: number;
};

const pushScript: Line[] = [
  { text: "$ git push", typed: true, pause: 500 },
  { text: "", pause: 350 },
  { text: "small · 1 question", pause: 280 },
  { text: "The session check moved into the handler.", pause: 550 },
  { text: "", pause: 180 },
  { text: "1  What sits between the browser and a route?", pause: 650 },
  { text: "", pause: 200 },
  { text: "   > the handler reads the session,", typed: true, pause: 180 },
  { text: "   > then hands the request to the route", typed: true, pause: 500 },
  { text: "   >", pause: 450 },
  { text: "", pause: 200 },
  { text: "pass  2.0 / 2", pause: 350 },
  { text: "You named the handler, and that it reads the session.", pause: 500 },
  { text: "", pause: 180 },
  { text: "To github.com:acme/app.git", pause: 160 },
  { text: "   a1b2c3d..e4f5g6h  main -> main", pause: 2200 },
];

/** Drop the skill where Claude Code and Cursor both look, then run the quiz. */
const skillScript: Line[] = [
  { text: "$ mkdir -p .claude/skills/shipquiz .agents/skills", typed: true, pause: 350 },
  { text: "$ curl -fsSL https://shipquiz.com/skill/SKILL.md \\", typed: true, pause: 80 },
  { text: "    -o .claude/skills/shipquiz/SKILL.md", typed: true, pause: 280 },
  { text: "$ cp -R .claude/skills/shipquiz .agents/skills/", typed: true, pause: 450 },
  { text: "", pause: 200 },
  { text: "skill ready", pause: 550 },
  { text: "", pause: 180 },
  { text: "$ shipquiz quiz", typed: true, pause: 400 },
  { text: "small · 1 question", pause: 280 },
  { text: "1  What sits between the browser and a route?", pause: 650 },
  { text: "", pause: 180 },
  { text: "   > the handler reads the session,", typed: true, pause: 160 },
  { text: "   > then hands the request to the route", typed: true, pause: 450 },
  { text: "", pause: 200 },
  { text: "pass  2.0 / 2", pause: 350 },
  { text: "The agent is told to run this before it calls the change done.", pause: 2200 },
];

const typeMs = 28;

/** The prompt and the pass line use the accent. Everything else stays ink. */
function lineColor(line: string) {
  if (line.startsWith("$") || line.startsWith("pass")) return "text-stamp";
  return undefined;
}

/** Wait, and remember the timer so it can be cleared. */
function wait(ms: number, timers: number[]) {
  return new Promise<void>((resolve) => {
    timers.push(window.setTimeout(resolve, ms));
  });
}

/** A looping terminal. `label` is the bar; `script` is what gets typed. */
function TerminalPlay({ script, label }: { script: Line[]; label: string }) {
  const [lines, setLines] = useState<string[]>([]);
  const [reduced, setReduced] = useState(false);

  useEffect(() => {
    const media = window.matchMedia("(prefers-reduced-motion: reduce)");
    const apply = () => setReduced(media.matches);
    apply();
    media.addEventListener("change", apply);
    return () => media.removeEventListener("change", apply);
  }, []);

  useEffect(() => {
    if (reduced) {
      setLines(script.map((line) => line.text));
      return;
    }

    let cancelled = false;
    const timers: number[] = [];

    async function play() {
      while (!cancelled) {
        const shown: string[] = [];
        setLines([]);
        await wait(300, timers);
        for (const line of script) {
          if (cancelled) return;
          if (line.typed) {
            for (let i = 1; i <= line.text.length; i++) {
              if (cancelled) return;
              setLines([...shown, line.text.slice(0, i)]);
              await wait(typeMs, timers);
            }
          }
          if (cancelled) return;
          shown.push(line.text);
          setLines([...shown]);
          await wait(line.pause ?? 160, timers);
        }
      }
    }

    void play();
    return () => {
      cancelled = true;
      for (const id of timers) window.clearTimeout(id);
    };
  }, [reduced, script]);

  return (
    <div className="border border-line border-t-[3px] border-t-mark bg-card" aria-hidden="true">
      <div className="bg-ink px-4 py-2 text-[11px] font-bold tracking-[0.07em] text-white uppercase">
        {label}
      </div>
      <pre className="h-[28rem] overflow-hidden px-4 py-4 font-mono text-sm leading-relaxed">
        {lines.map((line, index) => (
          <div key={index} className={lineColor(line)}>
            {line.length > 0 ? line : "\u00a0"}
            {!reduced && index === lines.length - 1 ? <span className="term-cursor text-mark" /> : null}
          </div>
        ))}
      </pre>
    </div>
  );
}

/** A looping terminal: git push, one question, then the push finishes. */
export function TerminalDemo() {
  return <TerminalPlay label="terminal" script={pushScript} />;
}

/** A looping terminal: add the skill, then the agent runs the quiz. */
export function SkillDemo() {
  return <TerminalPlay label="skill" script={skillScript} />;
}
