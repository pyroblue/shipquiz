"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState, type MouseEvent } from "react";

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
  const [open, setOpen] = useState(false);

  useEffect(() => {
    setOpen(false);
  }, [pathname]);

  /** Close the phone menu after a choice. */
  function closeMenu() {
    setOpen(false);
  }

  return (
    <header
      className={
        home
          ? `fixed inset-x-0 top-0 z-30 text-paper lg:sticky lg:z-20 lg:border-b lg:border-sky/40 lg:bg-ink ${open ? "bg-ink" : "bg-transparent"}`
          : "sticky top-0 z-20 border-b border-sky/40 bg-ink text-paper"
      }
    >
      <div
        className={`mx-auto flex w-full max-w-6xl items-center gap-x-6 px-5 py-4 ${home ? "justify-end" : "justify-between"}`}
      >
        {home ? null : (
          <Link
            href="/"
            className="display-face font-wordmark text-2xl leading-none uppercase"
            onClick={closeMenu}
          >
            SHIPQUIZ
          </Link>
        )}
        <button
          type="button"
          className={`inline-flex h-10 w-10 items-center justify-center lg:hidden ${open ? "text-paper" : "text-white drop-shadow-[0_1px_1px_rgba(26,49,72,0.9)]"}`}
          aria-expanded={open}
          aria-controls="site-menu"
          aria-label={open ? "Close menu" : "Open menu"}
          onClick={() => setOpen((value) => !value)}
        >
          {open ? <CloseIcon /> : <MenuIcon />}
        </button>
        <nav className="hidden gap-x-5 text-sm uppercase tracking-[0.06em] lg:flex">
          {links.map((link) => (
            <NavLink key={link.href} href={link.href} pathname={pathname}>
              {link.label}
            </NavLink>
          ))}
        </nav>
      </div>
      {open ? (
        <nav
          id="site-menu"
          className="bg-ink px-5 pb-4 lg:hidden"
        >
          {links.map((link) => (
            <NavLink
              key={link.href}
              href={link.href}
              pathname={pathname}
              onNavigate={closeMenu}
              stacked
            >
              {link.label}
            </NavLink>
          ))}
        </nav>
      ) : null}
    </header>
  );
}

/** One header link. Docs stays underlined while that page is open. */
function NavLink({
  href,
  pathname,
  onNavigate,
  stacked = false,
  children,
}: {
  href: string;
  pathname: string;
  onNavigate?: () => void;
  stacked?: boolean;
  children: string;
}) {
  const active = href === "/docs" && pathname.startsWith("/docs");

  /** Ease to a homepage section. Leave the page when the section is not here. */
  function onClick(event: MouseEvent<HTMLAnchorElement>) {
    const id = href.split("#")[1];
    const target = id && pathname === "/" ? document.getElementById(id) : null;
    onNavigate?.();
    if (!target) return;
    event.preventDefault();
    const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    requestAnimationFrame(() => {
      target.scrollIntoView({
        behavior: reduce ? "auto" : "smooth",
        block: "start",
      });
      window.history.pushState(null, "", `/#${id}`);
    });
  }

  return (
    <Link
      href={href}
      onClick={onClick}
      className={
        stacked
          ? `block py-3 text-base uppercase tracking-[0.06em] ${active ? "text-paper" : "text-sky"}`
          : active
            ? "text-paper underline decoration-sky underline-offset-4"
            : "text-sky hover:text-paper"
      }
      aria-current={active ? "page" : undefined}
    >
      {children}
    </Link>
  );
}

/** Three bars for the closed phone menu. */
function MenuIcon() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      className="h-6 w-6"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
    >
      <path d="M4 7h16M4 12h16M4 17h16" />
    </svg>
  );
}

/** An X for the open phone menu. */
function CloseIcon() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      className="h-6 w-6"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
    >
      <path d="M6 6l12 12M18 6L6 18" />
    </svg>
  );
}
