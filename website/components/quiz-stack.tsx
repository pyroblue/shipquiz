"use client";

import { useEffect, useState } from "react";
import { ShipWheel } from "@/components/ship-wheel";

const deck = [
  {
    question: "What sits between the browser and a route?",
    choices: ["The database", "The request handler", "The page component"],
    correct: 2,
  },
  {
    question: "Who decides whether a request is signed in?",
    choices: ["The request handler", "The route", "The browser"],
    correct: 1,
  },
  {
    question: "How do public pages and signed-in routes fit together?",
    choices: [
      "Every page uses the same gate",
      "Routes decide after rendering",
      "Only signed-in routes cross the auth boundary",
    ],
    correct: 3,
  },
];

/** How long the stamp stays up before the next card flips in. */
const holdMs = 2800;
const chooseMs = 1100;
/** Delay after the choice lands before the stamp appears. */
const markMs = 500;

/** Wait, and remember the timer so it can be cleared. */
function wait(ms: number, timers: number[]) {
  return new Promise<void>((resolve) => {
    timers.push(window.setTimeout(resolve, ms));
  });
}

/** A white check, drawn heavy enough to read inside the stamp. */
function CheckMark() {
  return (
    <svg viewBox="0 0 16 16" className="h-4 w-4" aria-hidden="true">
      <path
        d="M3 8.2 6.2 11.5 13 4.5"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

/** The stamp that lands on the right choice. */
function CorrectMark() {
  return (
    <p className="inline-flex items-center gap-2 text-sm font-bold tracking-[0.08em] text-stamp uppercase">
      <span className="inline-flex h-7 w-7 items-center justify-center bg-stamp text-paper">
        <CheckMark />
      </span>
      Correct
    </p>
  );
}

/** One card: the question, three choices, and the selected answer. */
function QuizCard({
  example,
  index,
  selected,
  marked,
}: {
  example: (typeof deck)[number];
  index: number;
  selected: number | null;
  marked: boolean;
}) {
  return (
    <div className="quiz-card h-full p-6 text-ink">
      <p className="text-xs tracking-[0.07em] text-muted">
        Question {index + 1}/{deck.length}
      </p>
      <p className="mt-3 min-h-[4.5rem] font-sans text-2xl leading-snug">{example.question}</p>
      <div className="mt-5 space-y-2 font-mono text-sm">
        {example.choices.map((choice, choiceIndex) => {
          const number = choiceIndex + 1;
          const active = selected === number;
          return (
            <p
              key={choice}
              className={`flex min-h-16 items-center rounded-sm border-2 px-3 py-2 transition-colors ${
                active ? "border-stamp bg-sky font-bold text-ink" : "border-line text-muted"
              }`}
            >
              <span className={`mr-2 ${active ? "text-stamp" : "text-mark"}`}>{number}</span>
              {choice}
            </p>
          );
        })}
      </div>
      <div className="mt-4">
        <div className={marked ? "quiz-stamp-in" : "invisible"}>
          <CorrectMark />
        </div>
      </div>
    </div>
  );
}

/** A paper stack that selects an answer, shows Correct!, then flips. */
export function QuizStack() {
  const [reduced, setReduced] = useState(false);
  const [index, setIndex] = useState(0);
  const [selected, setSelected] = useState<number | null>(null);
  const [marked, setMarked] = useState(false);

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
        setIndex(card);
        setSelected(null);
        setMarked(false);
        await wait(chooseMs, timers);
        if (cancelled) return;
        setSelected(deck[card].correct);
        await wait(markMs, timers);
        if (cancelled) return;
        setMarked(true);
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
            selected={example.correct}
            marked
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
                  selected={active ? selected : example.correct}
                  marked={active && marked}
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
