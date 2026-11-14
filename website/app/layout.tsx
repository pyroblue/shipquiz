import type { Metadata } from "next";
import { Newsreader, Special_Elite } from "next/font/google";
import localFont from "next/font/local";
import { NauticalScatter } from "@/components/nautical-scatter";
import { SiteFooter } from "@/components/site-footer";
import { SiteHeader } from "@/components/site-header";
import "./globals.css";

const body = Newsreader({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
  variable: "--font-body",
});

const display = localFont({
  src: "./fonts/Galley.ttf",
  weight: "500",
  variable: "--font-display",
});

const typewriter = Special_Elite({
  subsets: ["latin"],
  weight: "400",
  variable: "--font-typewriter",
});

export const metadata: Metadata = {
  metadataBase: new URL("https://shipquiz.com"),
  title: {
    default: "shipquiz — ship only what you can explain",
    template: "%s · shipquiz",
  },
  description:
    "A local pre-push quiz and GitHub Action that ask you to explain the architecture of a change before it ships. Bring your own model key. Diffs stay off shipquiz servers.",
};

/** A hand-drawn serif for titles, a text serif, and a typewriter for commands. */
export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="en"
      className={`${body.variable} ${display.variable} ${typewriter.variable} h-full`}
    >
      <body className="flex min-h-full flex-col">
        <SiteHeader />
        <div className="relative min-w-0 flex-1 overflow-x-clip">
          <NauticalScatter />
          <div className="relative z-10">{children}</div>
        </div>
        <SiteFooter />
      </body>
    </html>
  );
}
