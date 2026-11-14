import Link from "next/link";
import { CodeBlock } from "@/components/code-block";
import { QuizStack } from "@/components/quiz-stack";
import { ShipLoop } from "@/components/ship-loop";
import { ShipTitle } from "@/components/ship-title";
import { SkillDemo, TerminalDemo } from "@/components/terminal-demo";
import {
  commands,
  installCommand,
  pathExport,
  sourceBuild,
} from "@/lib/content";

/** Marketing page: install the hook, the Action, or the agent skill. */
export default function HomePage() {
  return (
    <main>
      <section className="relative w-full">
        <div className="ship-fade">
          <ShipLoop />
        </div>
        <div className="pointer-events-none absolute inset-0 flex items-center">
          <div className="ship-well mx-auto flex w-full max-w-6xl justify-center px-5">
            <ShipTitle />
          </div>
        </div>
      </section>
      <section className="mx-auto grid min-w-0 max-w-6xl gap-12 px-5 py-14 lg:grid-cols-[1.1fr_0.9fr] lg:items-center lg:py-20">
        <div className="max-w-xl">
          <h1 className="font-serif leading-tight text-ink">
            <span className="block text-4xl tracking-tighter sm:text-5xl sm:whitespace-nowrap uppercase">
              You&apos;re the captain.
            </span>
            <span className="mt-2 block text-4xl tracking-tighter sm:text-[2.4rem] sm:whitespace-nowrap normal-case">
              Ship only what you can{" "}
              <span className="text-stamp">explain.</span>
            </span>
          </h1>
          <p className="mt-6 max-w-xl text-lg leading-relaxed">
            shipquiz asks a few questions about your change and blocks{" "}
            <code className="font-mono text-stamp">git push</code> until the
            answers show you understand it. You use your own API key and model.
            The diff never goes to a shipquiz server.
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
              Run the path line if you want to stay in the same terminal. A new
              one already knows where shipquiz is. Have the repo checked out?
              Build it from there instead.
            </p>
          </div>
          <div className="space-y-4">
            <CodeBlock label="Install" code={installCommand} />
            <CodeBlock label="This shell" code={pathExport} />
            <CodeBlock label="Build from source" code={sourceBuild} />
          </div>
        </div>
      </section>

      <section className="border-t border-line">
        <div className="mx-auto grid max-w-6xl gap-6 px-5 py-14 md:grid-cols-3">
          <article className="border border-line border-t-[3px] border-t-mark bg-card p-6">
            <h2 className="font-serif text-3xl">Before the push</h2>
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
            <h2 className="font-serif text-3xl">On the pull request</h2>
            <p className="mt-3 leading-relaxed">
              The Action posts the questions as a comment and opens a check
              named <code className="font-mono">shipquiz</code>. Reply with
              answers, and the check then passes or fails with an explanation of
              the correct answer.
            </p>
            <p className="command mt-4 border-t border-line pt-3 font-mono text-sm text-muted">
              shipquiz github-action --write
            </p>
          </article>
          <article className="border border-line border-t-[3px] border-t-mark bg-card p-6">
            <h2 className="font-serif text-3xl">In the editor</h2>
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
              Add the skill to the repo. Claude Code and Cursor both read{" "}
              <code className="font-mono text-stamp">.claude/skills</code>. The
              copy in{" "}
              <code className="font-mono text-stamp">.agents/skills</code> covers
              the other agents. It tells your agent to run{" "}
              <code className="font-mono text-stamp">shipquiz quiz</code> when a
              change is about to ship. You can also ask for it by name. It
              never sees your API key.
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
