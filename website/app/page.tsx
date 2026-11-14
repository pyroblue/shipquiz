import Link from "next/link";
import { CodeBlock } from "@/components/code-block";
import { QuizStack } from "@/components/quiz-stack";
import { ShipLoop } from "@/components/ship-loop";
import { ShipTitle } from "@/components/ship-title";
import { SkillDemo, TerminalDemo } from "@/components/terminal-demo";
import { commands, installCommand, sourceBuild } from "@/lib/content";

/** Marketing page: install the hook, the Action, or the agent skill. */
export default function HomePage() {
  return (
    <main>
      <section className="relative w-full">
        <div className="ship-fade">
          <ShipLoop />
        </div>
        <div className="pointer-events-none absolute inset-0 flex items-center">
          <div className="ship-well mx-auto flex w-full max-w-6xl justify-center px-2 sm:px-5">
            <ShipTitle />
          </div>
        </div>
      </section>
      <section className="mx-auto grid min-w-0 max-w-6xl gap-12 px-5 py-14 lg:grid-cols-[1.1fr_0.9fr] lg:items-center lg:py-20">
        <div className="max-w-xl">
          <h1 className="font-serif leading-tight text-ink">
            <span className="block text-[1.85rem] tracking-tighter sm:text-5xl sm:whitespace-nowrap uppercase">
              You&apos;re the captain.
            </span>
            <span className="mt-2 block text-[1.5rem] tracking-tighter sm:text-[2.4rem] sm:whitespace-nowrap normal-case">
              Ship only what you can{" "}
              <span className="text-stamp">explain.</span>
            </span>
          </h1>
          <p className="mt-6 max-w-xl text-lg leading-relaxed">
            shipquiz asks a few questions about your change. It can block{" "}
            <code className="font-mono text-stamp">git push</code>, fail a CI
            check, or run on its own as a skill. You might not have written the
            code. You still have to understand the architecture. You use your
            own API key and model. The diff never goes to a shipquiz server.
          </p>
        </div>
        <QuizStack />
      </section>

      <section id="install" className="border-t border-line">
        <div className="mx-auto grid max-w-6xl gap-8 px-5 py-14 lg:grid-cols-[0.8fr_1.2fr]">
          <div>
            <h2 className="font-serif text-3xl">Install</h2>
            <p className="mt-4 leading-relaxed text-muted">
              Paste the install command on a Mac or Linux machine. It downloads
              shipquiz and adds it to your shell. Setup then asks for your API
              key, a model, and whether to quiz this repo before a push.
            </p>
            <p className="mt-4 leading-relaxed text-muted">
              Open a new terminal to run it. Have the repo checked out? Build
              it from there instead.
            </p>
          </div>
          <div className="space-y-4">
            <CodeBlock label="Install" code={installCommand} />
            <CodeBlock label="Build from source" code={sourceBuild} />
          </div>
        </div>
      </section>

      <section className="border-t border-line">
        <div className="mx-auto grid max-w-6xl gap-6 px-5 py-14 md:grid-cols-3">
          <article className="border border-line border-t-[3px] border-t-mark bg-card p-6">
            <p className="flex items-center gap-1.5 text-xs tracking-[0.08em] text-muted uppercase">
              <span className="text-stamp">
                <AnchorIcon />
              </span>
              pre-push hook
            </p>
            <h2 className="mt-1 font-serif text-3xl">Before the push</h2>
            <p className="mt-3 leading-relaxed">
              <code className="font-mono">shipquiz install</code> adds a
              pre-push hook. Git sends it the commits you are about to push. If
              you pass, the push goes through. If you don&apos;t, it stops.
            </p>
          </article>
          <article
            id="action"
            className="border border-line border-t-[3px] border-t-mark bg-card p-6"
          >
            <p className="flex items-center gap-1.5 text-xs tracking-[0.08em] text-muted uppercase">
              <span className="text-stamp">
                <GitHubIcon />
              </span>
              github action
            </p>
            <h2 className="mt-1 font-serif text-3xl">On the pull request</h2>
            <p className="mt-3 leading-relaxed">
              The Action posts the questions as a comment and opens a check
              named <code className="font-mono">shipquiz</code>. Reply with the
              selected numbers, and the check then passes or fails with an
              explanation of the correct answer.
            </p>
            <p className="command mt-4 border-t border-line pt-3 font-mono text-sm text-muted">
              shipquiz github-action
            </p>
          </article>
          <article className="border border-line border-t-[3px] border-t-mark bg-card p-6">
            <p className="flex items-center gap-1.5 text-xs tracking-[0.08em] text-muted uppercase">
              <span className="text-stamp">
                <SkillIcon />
              </span>
              shipquiz skill
            </p>
            <h2 className="mt-1 font-serif text-3xl">In the editor</h2>
            <p className="mt-3 leading-relaxed">
              The skill tells your agent to run shipquiz when a change is about
              to ship. You can also ask for it by name. It never sees your API
              key, and it does not grade the diff itself.
            </p>
          </article>
        </div>
      </section>

      <section id="skill" className="border-t border-line">
        <div className="mx-auto grid max-w-6xl items-start gap-10 px-5 py-14 lg:grid-cols-[0.75fr_1.25fr]">
          <div>
            <h2 className="font-serif text-3xl">Your agent asks first.</h2>
            <p className="mt-4 leading-relaxed text-muted">
              The skill tells your agent to quiz you before it calls a change
              done. You can ask for it by name. It never sees your API key.
            </p>
            <p className="mt-4">
              <Link
                href="/docs#skill"
                className="text-stamp underline decoration-stamp/40 underline-offset-4 hover:decoration-stamp"
              >
                How to add the skill
              </Link>
            </p>
          </div>
          <SkillDemo />
        </div>
      </section>

      <section id="hook" className="border-t border-line">
        <div className="mx-auto grid max-w-6xl items-start gap-10 px-5 py-14 lg:grid-cols-[0.75fr_1.25fr]">
          <div>
            <h2 className="font-serif text-3xl">You push. It asks.</h2>
            <p className="mt-4 leading-relaxed text-muted">
              <code className="font-mono text-stamp">git push</code> stops for a
              question about the change. Answer it, and the push finishes.
            </p>
          </div>
          <TerminalDemo />
        </div>
      </section>

      <section id="commands" className="border-t border-line">
        <div className="mx-auto max-w-6xl px-5 py-14">
          <h2 className="font-serif text-3xl">Commands</h2>
          <dl className="mt-8 divide-y divide-line border-y border-line">
            {commands.map((item) => (
              <div
                key={item.command}
                className="grid gap-2 py-4 md:grid-cols-[22rem_1fr]"
              >
                <dt className="font-mono text-sm">{item.command}</dt>
                <dd className="text-muted">{item.detail}</dd>
              </div>
            ))}
          </dl>
          <Link
            href="/docs"
            className="mt-8 inline-flex items-center gap-2.5 bg-ink px-5 py-2.5 text-base font-bold tracking-[0.06em] text-paper uppercase hover:bg-stamp"
          >
            <CompassMark />
            Full docs
          </Link>
        </div>
      </section>
    </main>
  );
}

/** An anchor for the pre-push hook card. */
function AnchorIcon() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      className="h-4 w-4"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
    >
      <circle cx="12" cy="4.8" r="1.8" />
      <path d="M12 6.6v11" />
      <path d="M8.4 9.4h7.2" />
      <path d="M7 16.2c1.3 3.4 8.7 3.4 10 0" />
    </svg>
  );
}

/** The GitHub mark for the Action card. */
function GitHubIcon() {
  return (
    <svg aria-hidden="true" viewBox="0 0 24 24" className="h-4 w-4" fill="currentColor">
      <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
    </svg>
  );
}

/** A skill file for the editor card. */
function SkillIcon() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      className="h-4 w-4"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinejoin="round"
    >
      <path d="M7 3.5h7.2L19 8.2V20.5H7z" />
      <path d="M14 3.5V8.2h5" />
      <path d="M10 13h6M10 16.5h4" strokeLinecap="round" />
    </svg>
  );
}

/** Compass rose on the docs button. */
function CompassMark() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      className="h-6 w-6 shrink-0"
      fill="currentColor"
    >
      <circle
        cx="12"
        cy="12"
        r="9.5"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
      />
      <path d="M12 3.4 15.6 13.1 12 11.2 8.4 13.1Z" />
      <path d="M12 20.6 10.2 13.7 12 15.1 13.8 13.7Z" opacity="0.45" />
    </svg>
  );
}
