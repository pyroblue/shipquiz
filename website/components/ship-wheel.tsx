/** The ship's wheel, floated on the corner of the quiz cards. */
export function ShipWheel() {
  return (
    <div
      aria-hidden="true"
      className="pointer-events-none absolute -right-1 -bottom-2 z-20 h-24 w-24 rotate-12 text-ink"
      style={{
        backgroundColor: "currentColor",
        WebkitMaskImage: "url(/ship-wheel.svg)",
        maskImage: "url(/ship-wheel.svg)",
        WebkitMaskRepeat: "no-repeat",
        maskRepeat: "no-repeat",
        WebkitMaskSize: "contain",
        maskSize: "contain",
        WebkitMaskPosition: "center",
        maskPosition: "center",
        maskMode: "alpha",
      }}
    />
  );
}
