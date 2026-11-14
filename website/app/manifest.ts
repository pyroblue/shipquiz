import type { MetadataRoute } from "next";

/** Home-screen icons, plus the navy used for the mobile browser chrome. */
export default function manifest(): MetadataRoute.Manifest {
  return {
    name: "shipquiz",
    short_name: "shipquiz",
    description: "Ship only what you can explain.",
    start_url: "/",
    display: "browser",
    background_color: "#f3f0e6",
    theme_color: "#1a3148",
    icons: [
      {
        src: "/icon-192.png",
        sizes: "192x192",
        type: "image/png",
        purpose: "any",
      },
      {
        src: "/icon-512.png",
        sizes: "512x512",
        type: "image/png",
        purpose: "any",
      },
    ],
  };
}
