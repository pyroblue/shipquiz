"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const links = [
  { href: "/#install", label: "Install" },
  { href: "/#hook", label: "Hook" },
  { href: "/#action", label: "Github Action" },
  { href: "/#skill", label: "Skill" },
  { href: "/docs", label: "Docs" },
];

/** Site header. The wordmark is omitted on the home page, where the ship carries it. */
export function SiteHeader() {
  const pathname = usePathname();
  const home = pathname === "/";

  return (
    <header className="sticky top-0 z-20 border-b border-sky/40 bg-ink text-paper">
      <div
        className={`mx-auto flex w-full max-w-6xl flex-wrap items-center gap-x-6 gap-y-3 px-5 py-4 ${home ? "justify-end" : "justify-between"}`}
      >
        {home ? null : (
          <Link href="/" className="display-face font-wordmark text-2xl leading-none">
            shipquiz
          </Link>
        )}
        <nav className="flex max-w-full flex-wrap gap-x-5 gap-y-2 text-sm uppercase tracking-[0.06em]">
          {links.map((link) => {
            const active = link.href === "/docs" && pathname.startsWith("/docs");
            return (
              <Link
                key={link.href}
                href={link.href}
                className={
                  active
                    ? "text-paper underline decoration-sky underline-offset-4"
                    : "text-sky hover:text-paper"
                }
                aria-current={active ? "page" : undefined}
              >
                {link.label}
              </Link>
            );
          })}
        </nav>
      </div>
    </header>
  );
}
