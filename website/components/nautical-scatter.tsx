"use client";

import { useEffect, useRef } from "react";

const marks = [
  { src: "/nautical/compass.png", top: "6%", left: "-6%", size: 520, rotate: -16, fade: 0.11, depth: 0.18 },
  { src: "/nautical/wheel.png", top: "16%", left: "68%", size: 460, rotate: 18, fade: 0.1, depth: 0.28 },
  { src: "/nautical/gulls.png", top: "30%", left: "-8%", size: 640, rotate: -8, fade: 0.1, depth: 0.14 },
  { src: "/nautical/anchor.png", top: "42%", left: "72%", size: 420, rotate: 12, fade: 0.12, depth: 0.32 },
  { src: "/nautical/whale.png", top: "54%", left: "-10%", size: 560, rotate: -12, fade: 0.1, depth: 0.2 },
  { src: "/nautical/wave.png", top: "66%", left: "70%", size: 480, rotate: 9, fade: 0.11, depth: 0.26 },
  { src: "/nautical/compass.png", top: "78%", left: "-4%", size: 400, rotate: 22, fade: 0.1, depth: 0.16 },
  { src: "/nautical/wheel.png", top: "86%", left: "66%", size: 440, rotate: -14, fade: 0.11, depth: 0.3 },
];

/** Shift each drawing a little slower than the page, so it sits further back. */
function placeMarks(root: HTMLElement, scrollY: number) {
  const nodes = root.children;
  marks.forEach((mark, index) => {
    const el = nodes[index];
    if (!(el instanceof HTMLElement)) return;
    const shift = scrollY * mark.depth;
    el.style.transform = `translate3d(0, ${shift}px, 0) rotate(${mark.rotate}deg)`;
  });
}

/** Faded navy drawings scattered down the page, behind the type. */
export function NauticalScatter() {
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const root = rootRef.current;
    if (!root) return;

    const media = window.matchMedia("(prefers-reduced-motion: reduce)");
    let frame = 0;

    /** Read the scroll position and lag the drawings behind the page. */
    function apply() {
      if (!root) return;
      placeMarks(root, media.matches ? 0 : window.scrollY);
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
  }, []);

  return (
    <div
      ref={rootRef}
      className="pointer-events-none absolute inset-0 z-0 hidden overflow-hidden lg:block"
      aria-hidden="true"
    >
      {marks.map((mark) => (
        <div
          key={`${mark.src}-${mark.top}`}
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
