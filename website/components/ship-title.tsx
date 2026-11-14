"use client";

import { useEffect, useLayoutEffect, useRef } from "react";

/** Runs before paint in the browser, and after render on the server. */
const useIsoLayoutEffect = typeof window === "undefined" ? useEffect : useLayoutEffect;

/** The hero word fills most of the content well. */
const heroScale = 0.85;

/** Set the font size so the glyphs fill most of the well's inner width. */
function fitTitle(el: HTMLParagraphElement, well: HTMLElement) {
  const style = getComputedStyle(well);
  const padding = parseFloat(style.paddingLeft) + parseFloat(style.paddingRight);
  const target = (well.clientWidth - padding) * heroScale;
  const current = parseFloat(getComputedStyle(el).fontSize);
  const glyphs = el.scrollWidth;
  if (target <= 0 || current <= 0 || glyphs <= 0) return;
  el.style.fontSize = `${(current * target) / glyphs}px`;
}

/** The ship word, centered in the content well and a little narrower than it. */
export function ShipTitle() {
  const ref = useRef<HTMLParagraphElement>(null);

  useIsoLayoutEffect(() => {
    const el = ref.current;
    const well = el?.parentElement;
    if (!el || !well) return;

    const fit = () => fitTitle(el, well);
    fit();
    const observer = new ResizeObserver(fit);
    observer.observe(well);
    document.fonts.ready.then(fit).catch(() => undefined);
    return () => observer.disconnect();
  }, []);

  return (
    <>
      <svg className="absolute h-0 w-0" aria-hidden="true">
        <filter
          id="hero-distress"
          x="-8%"
          y="-25%"
          width="116%"
          height="160%"
          colorInterpolationFilters="sRGB"
        >
          <feTurbulence
            type="fractalNoise"
            baseFrequency="0.55"
            numOctaves="3"
            seed="11"
            result="grit"
          />
          <feColorMatrix
            in="grit"
            type="matrix"
            values="1 0 0 0 0  0 1 0 0 0  0 0 1 0 0  1 0 0 0 -0.45"
            result="specks"
          />
          <feComposite in="SourceGraphic" in2="specks" operator="out" />
        </filter>
      </svg>
      <p
        ref={ref}
        className="ship-title display-face w-max font-wordmark leading-[0.8] whitespace-nowrap text-white uppercase"
      >
        shipquiz
      </p>
    </>
  );
}
