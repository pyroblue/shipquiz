import Link from "next/link";

/** Footer. Keys and diffs never land on shipquiz servers. */
export function SiteFooter() {
  return (
    <footer className="mt-20 border-t-2 border-mark text-ink">
      <div className="mx-auto flex max-w-6xl flex-col gap-3 px-5 py-8 text-sm sm:flex-row sm:items-center sm:justify-between">
        <p className="text-muted">
          macOS and Linux. Use your own API key. Diffs stay on your machine.
        </p>
        <div className="flex gap-5">
          <Link href="/docs" className="hover:underline">
            Docs
          </Link>
          <a href="https://github.com/pyroblue/shipquiz" className="hover:underline">
            GitHub
          </a>
        </div>
      </div>
    </footer>
  );
}
