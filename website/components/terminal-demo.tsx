"use client";

import { useEffect, useRef, useState, type RefObject } from "react";

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
  { text: "   1  The database", pause: 120 },
  { text: "   2  The request handler", pause: 120 },
  { text: "   3  The page component", pause: 300 },
  { text: "   press 1, 2, or 3 · ? reveals the answer", pause: 350 },
  { text: "   > 2", typed: true, pause: 350 },
  { text: "   ✓ correct", pause: 250 },
  { text: "   The handler owns the session boundary.", pause: 500 },
  { text: "", pause: 200 },
  { text: "⚓ PASS  cleared to ship! 🎉  2.0 / 2", pause: 500 },
  { text: "", pause: 180 },
  { text: "To github.com:acme/app.git", pause: 160 },
  { text: "   a1b2c3d..e4f5g6h  main -> main", pause: 2200 },
];

type ChatLine = {
  from: "you" | "agent";
  text: string;
  /** Type this message one character at a time. */
  typed?: boolean;
  /** Wait after the message is on screen, in milliseconds. */
  pause?: number;
  /** The quiz passed, so the bubble uses the stamp. */
  pass?: boolean;
  /** Numbered choices inside an agent question. */
  choices?: string[];
  /** The 1-based choice that gets clicked. */
  choice?: number;
};

/** A chat where the agent quizzes before it calls the change done. */
const skillChat: ChatLine[] = [
  { from: "you", text: "Ship this change.", typed: true, pause: 500 },
  { from: "agent", text: "One question before I call it done.", pause: 700 },
  {
    from: "agent",
    text: "What sits between the browser and a route?",
    choices: ["The database", "The request handler", "The page component"],
    choice: 2,
    pause: 700,
  },
  { from: "agent", text: "Correct. The handler owns the session boundary.", pause: 700 },
  { from: "agent", text: "Cleared to ship.", pass: true, pause: 400 },
];

const typeMs = 28;
/** How long the finished transcript stays before the loop starts over. */
const holdMs = 5000;

/** Commands and successful quiz lines use the accent. */
function lineColor(line: string) {
  if (line.startsWith("$") || line.startsWith("⚓") || line.includes("✓")) return "text-stamp";
  return undefined;
}

/** Wait, and remember the timer so it can be cleared. */
function wait(ms: number, timers: number[]) {
  return new Promise<void>((resolve) => {
    timers.push(window.setTimeout(resolve, ms));
  });
}

/** Whether the user asked for less motion. */
function useReducedMotion() {
  const [reduced, setReduced] = useState(false);

  useEffect(() => {
    const media = window.matchMedia("(prefers-reduced-motion: reduce)");
    const apply = () => setReduced(media.matches);
    apply();
    media.addEventListener("change", apply);
    return () => media.removeEventListener("change", apply);
  }, []);

  return reduced;
}

/** True while the element is far enough into the viewport to play. */
function useInView(ref: RefObject<HTMLElement | null>) {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const node = ref.current;
    if (!node) return;
    const observer = new IntersectionObserver(
      ([entry]) => setVisible(entry.isIntersecting),
      { threshold: 0.35 },
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, [ref]);

  return visible;
}

/** A looping terminal. `label` is the bar; `script` is what gets typed. */
function TerminalPlay({ script, label }: { script: Line[]; label: string }) {
  const rootRef = useRef<HTMLDivElement>(null);
  const [played, setLines] = useState<string[] | null>(null);
  const reduced = useReducedMotion();
  const visible = useInView(rootRef);
  const lines = reduced ? script.map((line) => line.text) : (played ?? []);

  useEffect(() => {
    if (reduced || !visible) return;

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
        await wait(holdMs, timers);
      }
    }

    void play();
    return () => {
      cancelled = true;
      for (const id of timers) window.clearTimeout(id);
    };
  }, [reduced, script, visible]);

  return (
    <div
      ref={rootRef}
      className="border border-line border-t-[3px] border-t-mark bg-card"
      aria-hidden="true"
    >
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

/** One message in the skill chat. `typing` draws the cursor on a partial line. */
function ChatBubble({ message, typing }: { message: ChatLine; typing: boolean }) {
  const you = message.from === "you";
  const shell = you
    ? "bg-ink text-paper"
    : message.pass
      ? "border border-stamp bg-sky font-medium text-stamp"
      : "border border-line bg-paper text-ink";

  return (
    <div className={`flex flex-col gap-1 ${you ? "items-end" : "items-start"}`}>
      <p className="text-[11px] tracking-[0.08em] text-muted uppercase">{you ? "You" : "Agent"}</p>
      <div className={`max-w-[85%] rounded-2xl px-3 py-2 text-sm leading-relaxed ${shell}`}>
        <p>
          {message.text}
          {typing ? <span className="term-cursor" /> : null}
        </p>
        {message.choices ? (
          <div className="mt-2 space-y-1.5">
            {message.choices.map((choice, index) => {
              const picked = message.choice === index + 1;
              return (
                <p
                  key={choice}
                  className={`rounded-xl border-2 px-3 py-1.5 ${
                    picked
                      ? "border-stamp bg-sky font-bold text-ink"
                      : "border-line bg-card text-ink"
                  }`}
                >
                  <span className={`mr-2 ${picked ? "text-stamp" : "text-mark"}`}>{index + 1}</span>
                  {choice}
                </p>
              );
            })}
          </div>
        ) : null}
      </div>
    </div>
  );
}

/** A looping chat: you ask to ship, the agent quizzes, then clears it. */
export function SkillDemo() {
  const rootRef = useRef<HTMLDivElement>(null);
  const logRef = useRef<HTMLDivElement>(null);
  const [played, setMessages] = useState<ChatLine[] | null>(null);
  const reduced = useReducedMotion();
  const visible = useInView(rootRef);
  const messages = reduced ? skillChat : (played ?? []);

  useEffect(() => {
    const log = logRef.current;
    if (!log) return;
    log.scrollTop = log.scrollHeight;
  }, [messages]);

  useEffect(() => {
    if (reduced || !visible) return;

    let cancelled = false;
    const timers: number[] = [];

    async function play() {
      while (!cancelled) {
        const shown: ChatLine[] = [];
        setMessages([]);
        await wait(300, timers);
        for (const message of skillChat) {
          if (cancelled) return;
          if (message.typed) {
            for (let i = 1; i <= message.text.length; i++) {
              if (cancelled) return;
              setMessages([...shown, { ...message, text: message.text.slice(0, i) }]);
              await wait(typeMs, timers);
            }
          }
          if (cancelled) return;
          if (message.choices && message.choice) {
            shown.push({ ...message, choice: undefined });
            setMessages([...shown]);
            await wait(900, timers);
            if (cancelled) return;
            shown[shown.length - 1] = message;
          } else {
            shown.push(message);
          }
          setMessages([...shown]);
          await wait(message.pause ?? 400, timers);
        }
        await wait(holdMs, timers);
      }
    }

    void play();
    return () => {
      cancelled = true;
      for (const id of timers) window.clearTimeout(id);
    };
  }, [reduced, visible]);

  return (
    <div
      ref={rootRef}
      className="border border-line border-t-[3px] border-t-mark bg-card"
      aria-hidden="true"
    >
      <div className="bg-ink px-4 py-2 text-[11px] font-bold tracking-[0.07em] text-white uppercase">
        chat
      </div>
      <div ref={logRef} className="flex h-[28rem] flex-col gap-3 overflow-y-auto px-4 py-4">
        {messages.map((message, index) => (
          <ChatBubble
            key={`${message.from}-${index}`}
            message={message}
            typing={
              !reduced &&
              index === messages.length - 1 &&
              message.text !== skillChat[index]?.text
            }
          />
        ))}
      </div>
    </div>
  );
}
