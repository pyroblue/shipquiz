"use client";

import { useEffect, useRef, useState } from "react";

const art = [
  "/nautical/compass.png",
  "/nautical/wheel.png",
  "/nautical/gulls.png",
  "/nautical/anchor.png",
  "/nautical/whale.png",
  "/nautical/wave.png",
];

const slots = [
  { top: "6%", left: "-6%", size: 520, fade: 0.11, depth: 0.18 },
  { top: "16%", left: "68%", size: 460, fade: 0.1, depth: 0.28 },
  { top: "30%", left: "-8%", size: 640, fade: 0.1, depth: 0.14 },
  { top: "42%", left: "72%", size: 420, fade: 0.12, depth: 0.32 },
  { top: "54%", left: "-10%", size: 560, fade: 0.1, depth: 0.2 },
  { top: "66%", left: "70%", size: 480, fade: 0.11, depth: 0.26 },
  { top: "78%", left: "-4%", size: 400, fade: 0.1, depth: 0.16 },
  { top: "86%", left: "66%", size: 440, fade: 0.11, depth: 0.3 },
];

type Mark = (typeof slots)[number] & { src: string; rotate: number };

/** Shuffle a copy so the same list can be dealt in a new order. */
function shuffle<T>(items: T[]) {
  const deck = [...items];
  for (let i = deck.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    const current = deck[i];
    deck[i] = deck[j];
    deck[j] = current;
  }
  return deck;
}

/** Fill each slot with a drawing and a new tilt. */
function deal(): Mark[] {
  const deck = shuffle([...art, ...art]);
  return slots.map((slot, index) => ({
    ...slot,
    src: deck[index],
    rotate: Math.round(Math.random() * 80 - 40),
  }));
}

/** Shift each drawing a little slower than the page, so it sits further back. */
function placeMarks(root: HTMLElement, scrollY: number, placed: Mark[]) {
  const nodes = root.children;
  placed.forEach((mark, index) => {
    const el = nodes[index];
    if (!(el instanceof HTMLElement)) return;
    const shift = scrollY * mark.depth;
    el.style.transform = `translate3d(0, ${shift}px, 0) rotate(${mark.rotate}deg)`;
  });
}

/** Faded navy drawings scattered down the page, behind the type. */
export function NauticalScatter() {
  const rootRef = useRef<HTMLDivElement>(null);
  const [marks, setMarks] = useState<Mark[]>([]);

  useEffect(() => {
    setMarks(deal());
  }, []);

  useEffect(() => {
    const root = rootRef.current;
    if (!root || marks.length === 0) return;

    const media = window.matchMedia("(prefers-reduced-motion: reduce)");
    let frame = 0;

    /** Read the scroll position and lag the drawings behind the page. */
    function apply() {
      if (!root) return;
      placeMarks(root, media.matches ? 0 : window.scrollY, marks);
    }

    /** Coalesce scroll events to one update per frame. */
    function onScroll() {
      cancelAnimationFrame(frame);
      frame = requestAnimationFrame(apply);
    }

    apply();
    window.addEventListener("scroll", onScroll, { passive: true });
    media.addEventListener("change", onScroll);
    return () => {
      cancelAnimationFrame(frame);
      window.removeEventListener("scroll", onScroll);
      media.removeEventListener("change", onScroll);
    };
  }, [marks]);

  return (
    <div
      ref={rootRef}
      className="pointer-events-none absolute inset-0 z-0 overflow-hidden"
      aria-hidden="true"
    >
      {marks.map((mark, index) => (
        <div
          key={`${mark.top}-${index}`}
          className="absolute"
          style={{
            top: mark.top,
            left: mark.left,
            width: mark.size,
            height: mark.size,
            transform: `rotate(${mark.rotate}deg)`,
            opacity: mark.fade,
            backgroundColor: "#1a3148",
            WebkitMaskImage: `url(${mark.src})`,
            maskImage: `url(${mark.src})`,
            WebkitMaskRepeat: "no-repeat",
            maskRepeat: "no-repeat",
            WebkitMaskSize: "contain",
            maskSize: "contain",
            WebkitMaskPosition: "center",
            maskPosition: "center",
            maskMode: "alpha",
          }}
        />
      ))}
    </div>
  );
}
