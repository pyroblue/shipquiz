"use client";

import { useEffect, useState } from "react";
import { ShipWheel } from "@/components/ship-wheel";

const deck = [
  {
    question: "What sits between the browser and a route?",
    answer: "the handler reads the session, then hands the request to the route",
  },
  {
    question: "Who decides whether a request is signed in?",
    answer: "the handler. a route receives whatever the handler already loaded",
  },
  {
    question: "How do public pages and signed-in routes fit together?",
    answer: "public pages stay outside the handler. signed-in routes go through it",
  },
];

const typeMs = 22;

/** How long Correct! stays up before the next card flips in. */
const holdMs = 2800;

/** Wait, and remember the timer so it can be cleared. */
function wait(ms: number, timers: number[]) {
  return new Promise<void>((resolve) => {
    timers.push(window.setTimeout(resolve, ms));
  });
}

/** A green check and the word Correct! */
function CorrectMark() {
  return (
    <p className="quiz-correct inline-flex items-center gap-1.5 text-sm font-medium">
      <svg viewBox="0 0 16 16" className="h-4 w-4" aria-hidden="true">
        <path
          d="M3 8.2 6.2 11.5 13 4.5"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.8"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
      Correct!
    </p>
  );
}

/** One card: the question, the typed answer, and Correct! once it lands. */
function QuizCard({
  example,
  index,
  answer,
  showVerdict,
  cursor,
}: {
  example: (typeof deck)[number];
  index: number;
  answer: string;
  showVerdict: boolean;
  cursor: boolean;
}) {
  return (
    <div className="quiz-card h-full p-6 text-ink">
      <p className="text-xs tracking-[0.07em] text-muted">
        Question {index + 1}/{deck.length}
      </p>
      <p className="mt-3 font-sans text-2xl leading-snug">{example.question}</p>
      <p className="mt-5 grid border-t border-line pt-3 font-mono text-sm">
        <span className="invisible col-start-1 row-start-1">&gt; {example.answer}</span>
        <span className="col-start-1 row-start-1">
          &gt; {answer}
          {cursor ? <span className="term-cursor text-mark" /> : null}
        </span>
      </p>
      <div className={`mt-4 ${showVerdict ? "" : "invisible"}`}>
        <CorrectMark />
      </div>
    </div>
  );
}

/** A paper stack that types an answer, shows Correct!, then flips to the next. */
export function QuizStack() {
  const [reduced, setReduced] = useState(false);
  const [index, setIndex] = useState(0);
  const [typed, setTyped] = useState("");
  const [phase, setPhase] = useState<"type" | "hold">("type");

  useEffect(() => {
    const media = window.matchMedia("(prefers-reduced-motion: reduce)");
    const update = () => setReduced(media.matches);
    update();
    media.addEventListener("change", update);
    return () => media.removeEventListener("change", update);
  }, []);

  useEffect(() => {
    if (reduced) return;

    let cancelled = false;
    const timers: number[] = [];

    async function play() {
      let card = 0;
      while (!cancelled) {
        const answer = deck[card].answer;
        setIndex(card);
        setPhase("type");
        setTyped("");
        for (let i = 1; i <= answer.length; i++) {
          if (cancelled) return;
          setTyped(answer.slice(0, i));
          await wait(typeMs, timers);
        }
        if (cancelled) return;
        setPhase("hold");
        await wait(holdMs, timers);
        if (cancelled) return;
        card = (card + 1) % deck.length;
      }
    }

    void play();
    return () => {
      cancelled = true;
      for (const id of timers) window.clearTimeout(id);
    };
  }, [reduced]);

  if (reduced) {
    return (
      <div className="relative space-y-4" aria-hidden="true">
        {deck.map((example, card) => (
          <QuizCard
            key={example.question}
            example={example}
            index={card}
            answer={example.answer}
            showVerdict
            cursor={false}
          />
        ))}
        <ShipWheel />
      </div>
    );
  }

  return (
    <div className="quiz-deck relative px-4 py-5" aria-hidden="true">
      <div className="relative">
        <div className="quiz-sheet quiz-sheet-a" />
        <div className="quiz-sheet quiz-sheet-b" />
        <div className="relative z-10 grid">
          {deck.map((example, card) => {
            const active = card === index;
            return (
              <div
                key={example.question}
                className={`col-start-1 row-start-1 ${active ? "quiz-flip-in" : "invisible"}`}
              >
                <QuizCard
                  example={example}
                  index={card}
                  answer={active ? typed : example.answer}
                  showVerdict={active ? phase === "hold" : true}
                  cursor={active && phase === "type"}
                />
              </div>
            );
          })}
        </div>
      </div>
      <ShipWheel />
    </div>
  );
}
