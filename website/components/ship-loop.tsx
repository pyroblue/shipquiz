"use client";

import { useEffect, useState } from "react";

const posterSrc = "/ship-poster.jpg?v=6";

/**
 * Plays the ship loop at normal speed. The last two seconds fade into the
 * first two, so the loop has no jump. The first frame stays up until the
 * video can play, and is the still used when motion is reduced.
 */
export function ShipLoop() {
  const [reduced, setReduced] = useState(false);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const media = window.matchMedia("(prefers-reduced-motion: reduce)");
    const update = () => setReduced(media.matches);
    update();
    media.addEventListener("change", update);
    return () => media.removeEventListener("change", update);
  }, []);

  /**
   * Drops the still once this element already has a decoded frame.
   */
  const markReadyIfLoaded = (node: HTMLVideoElement | null) => {
    if (node && node.readyState >= HTMLMediaElement.HAVE_CURRENT_DATA) {
      setReady(true);
    }
  };

  if (reduced) {
    return (
      <img
        src={posterSrc}
        alt=""
        fetchPriority="high"
        decoding="async"
        className="block aspect-video w-full bg-paper object-cover"
      />
    );
  }

  return (
    <div className="relative">
      <video
        ref={markReadyIfLoaded}
        className="block aspect-video w-full bg-paper object-cover"
        autoPlay
        muted
        loop
        playsInline
        poster={posterSrc}
        aria-hidden="true"
        onLoadedData={() => setReady(true)}
      >
        <source src="/ship-loop.mp4?v=6" type="video/mp4" />
      </video>
      {!ready && (
        <img
          src={posterSrc}
          alt=""
          fetchPriority="high"
          decoding="async"
          className="absolute inset-0 h-full w-full object-cover"
        />
      )}
    </div>
  );
}
