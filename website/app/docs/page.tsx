import type { Metadata } from "next";
import { CodeBlock } from "@/components/code-block";
import {
  commands,
  globalConfig,
  installCommand,
  pathExport,
  repoConfig,
  skillInstall,
  sourceBuild,
  workflowYaml,
} from "@/lib/content";

export const metadata: Metadata = {
  title: "Docs",
  description:
    "Setup, config, the pre-push hook, the GitHub Action, and the agent skill for shipquiz.",
};

const toc = [
  ["setup", "Setup"],
  ["commands", "Commands"],
  ["config", "Config"],
  ["hook", "Hook"],
  ["action", "GitHub Action"],
  ["skill", "Skill"],
  ["scoring", "Scoring"],
  ["source", "From source"],
];

/** Reference for installing and operating shipquiz. */
export default function DocsPage() {
  return (
    <main className="mx-auto max-w-6xl px-5 py-10">
      <div className="grid min-w-0 gap-10 lg:grid-cols-[14rem_1fr]">
      <aside className="lg:sticky lg:top-24 lg:self-start">
        <p className="text-xs tracking-[0.08em] text-stamp">ON THIS PAGE</p>
        <nav className="mt-4 flex flex-wrap gap-x-4 gap-y-2 text-sm lg:flex-col">
          {toc.map(([id, label]) => (
            <a key={id} href={`#${id}`} className="text-muted hover:text-ink">
              {label}
            </a>
          ))}
        </nav>
      </aside>

      <div className="min-w-0 space-y-16">
        <header>
          <h1 className="font-serif text-4xl">Docs</h1>
          <p className="mt-4 max-w-2xl leading-relaxed text-muted">
            Use a pre-push hook, a GitHub Action, or an agent skill. All three
            run the same command-line tool, which calls your model provider.
            shipquiz does not store API keys or diffs.
          </p>
        </header>

        <section id="setup" className="space-y-4">
          <h2 className="font-serif text-3xl">Setup</h2>
          <p className="leading-relaxed text-muted">
            The installer detects macOS or Linux and your architecture, downloads
            the release into <code className="font-mono text-ink">~/.shipquiz/bin</code>,
            and appends that directory to <code className="font-mono text-ink">.zshrc</code> or{" "}
            <code className="font-mono text-ink">.bashrc</code> if it is not already there.
            Re-running it is safe.
          </p>
          <CodeBlock label="install" code={`${installCommand}\n${pathExport}`} />
          <ol className="list-decimal space-y-2 pl-5 leading-relaxed text-muted">
            <li>If a config already exists, setup asks before replacing it.</li>
            <li>Pick anthropic, openai, gemini, or xai. Gemini and xAI use the OpenAI API.</li>
            <li>The API key is read from the terminal and stored in the macOS Keychain, or in a file only you can read on Linux.</li>
            <li>shipquiz lists that provider’s models and lets you pick one. The default is a fast model for that provider.</li>
            <li>You can send a one-token test to check that the key and model work.</li>
            <li>Inside a git repo, setup offers to install the pre-push hook.</li>
          </ol>
          <p className="leading-relaxed text-muted">
            Set <code className="font-mono text-ink">base_url</code> when you use a host that speaks the OpenAI API.
            For Anthropic, <code className="font-mono text-ink">base_url</code> is the site origin, such as{" "}
            <code className="font-mono text-ink">https://api.anthropic.com</code>. For OpenAI, Gemini, and xAI it is the API root, and it already includes the version, such as <code className="font-mono text-ink">https://api.openai.com/v1</code> or <code className="font-mono text-ink">https://api.x.ai/v1</code>.
          </p>
        </section>

        <section id="commands" className="space-y-4">
          <h2 className="font-serif text-3xl">Commands</h2>
          <dl className="divide-y divide-line">
            {commands.map((item) => (
              <div key={item.command} className="grid gap-2 py-4 md:grid-cols-[22rem_1fr]">
                <dt className="font-mono text-sm">{item.command}</dt>
                <dd className="text-muted">{item.detail}</dd>
              </div>
            ))}
          </dl>
          <p className="text-sm leading-relaxed text-muted">
            The hook calls <code className="font-mono text-ink">shipquiz hook-run</code>.
            GitHub Actions calls <code className="font-mono text-ink">shipquiz action-run</code>.
            Neither is a command you type.
          </p>
        </section>

        <section id="config" className="space-y-4">
          <h2 className="font-serif text-3xl">Config</h2>
          <p className="leading-relaxed text-muted">
            Global config is <code className="font-mono text-ink">~/.config/shipquiz/config.toml</code>.
            A repo may add <code className="font-mono text-ink">.shipquiz.toml</code>. Repo values override the global file for the keys they set.
          </p>
          <CodeBlock label="~/.config/shipquiz/config.toml" code={globalConfig.trim()} />
          <CodeBlock label=".shipquiz.toml" code={repoConfig.trim()} />
          <p className="leading-relaxed text-muted">
            <code className="font-mono text-ink">question_style</code> is{" "}
            <code className="font-mono text-ink">default</code>,{" "}
            <code className="font-mono text-ink">failure-modes</code>, or{" "}
            <code className="font-mono text-ink">data-flow</code>.
            Default asks about the change as a whole. Failure-modes asks only what can break. Data-flow asks you to follow one value from one side of the diff to the other.
          </p>
          <p className="leading-relaxed text-muted">
            <code className="font-mono text-ink">shipquiz config set quiz.timeout_seconds 15</code> writes the global file.
            Keys you can set: <code className="font-mono text-ink">provider</code>,{" "}
            <code className="font-mono text-ink">model</code>,{" "}
            <code className="font-mono text-ink">base_url</code>,{" "}
            <code className="font-mono text-ink">question_style</code>,{" "}
            <code className="font-mono text-ink">quiz.min_questions</code>,{" "}
            <code className="font-mono text-ink">quiz.max_questions</code>,{" "}
            <code className="font-mono text-ink">quiz.passing_average</code>,{" "}
            <code className="font-mono text-ink">quiz.timeout_seconds</code>,{" "}
            <code className="font-mono text-ink">hooks.pre_push</code>.
          </p>
        </section>

        <section id="hook" className="space-y-4">
          <h2 className="font-serif text-3xl">Hook</h2>
          <p className="leading-relaxed text-muted">
            The hook is a short shell script. It runs the CLI with the commits git was about to push.
            If you already had a pre-push hook, shipquiz moves it to <code className="font-mono text-ink">pre-push.previous</code> and runs it after you pass.
            <code className="font-mono text-ink"> shipquiz uninstall</code> puts the old hook back.
          </p>
          <ul className="list-disc space-y-2 pl-5 leading-relaxed text-muted">
            <li>Empty diffs, lockfiles, and paths in <code className="font-mono text-ink">skip_paths</code> skip the quiz and let the push continue.</li>
            <li>Questions are saved by the two commits, the diff, the model, and the question style. A diff you already passed is not asked again.</li>
            <li>A failed quiz stops the push.</li>
            <li>If the model times out, the hook prints a warning and lets the push through.</li>
            <li>
              <code className="font-mono text-ink">SHIPQUIZ_SKIP=1</code> prints a skip line and lets the push through.{" "}
              <code className="font-mono text-ink">git push --no-verify</code> skips the hook entirely.
            </li>
            <li>Set <code className="font-mono text-ink">hooks.pre_push</code> to false to leave the hook installed but skip the quiz.</li>
          </ul>
          <p className="leading-relaxed text-muted">
            You answer in the terminal, one question at a time. A blank line submits the answer.
            If you fail and try the same diff again, shipquiz reuses the questions and only calls the model to grade.
          </p>
        </section>

        <section id="action" className="space-y-4">
          <h2 className="font-serif text-3xl">GitHub Action</h2>
          <p className="leading-relaxed text-muted">
            On a pull request, the Action writes the quiz, posts it as a comment, and opens a check named{" "}
            <code className="font-mono text-ink">shipquiz</code> with the result{" "}
            <code className="font-mono text-ink">action_required</code>. Require that check in branch protection if you want it to block merging.
            The diff is sent to your model provider, not to shipquiz.
          </p>
          <p className="leading-relaxed text-muted">
            Reply in the thread in your own words. The model matches that reply to each question.
            A comment that is not an answer is left alone.
          </p>
          <CodeBlock
            label="comment reply"
            code={`The token is checked in the handler, so a new route that skips it is public.\nA cache miss stamps the origin because the lock is process-local.`}
          />
          <p className="leading-relaxed text-muted">
            When the reply answers the quiz, the workflow posts what was missing and marks the check passed or failed.
            A failed quiz also fails the job. The repository secret is{" "}
            <code className="font-mono text-ink">SHIPQUIZ_LLM_API_KEY</code>, and its value is the model provider key.
          </p>
          <CodeBlock label=".github/workflows/shipquiz.yml" code={workflowYaml.trim()} />
          <p className="text-sm leading-relaxed text-muted">
            <code className="font-mono text-ink">shipquiz github-action</code> asks for the provider, model, and LLM API key, writes this file, and shows the secret to set.
            The key stays out of the file. <code className="font-mono text-ink">--print</code> prints the workflow without asking.
          </p>
        </section>

        <section id="skill" className="space-y-4">
          <h2 className="font-serif text-3xl">Skill</h2>
          <p className="leading-relaxed text-muted">
            Add the skill to a repo and the agent is told to run{" "}
            <code className="font-mono text-ink">shipquiz quiz</code> when a change is about to commit, push, or open a pull request.
            You can also invoke it by name. It is not a gate. The hook is what blocks the push.
            If the CLI is missing, the skill stops and shows the install commands.
          </p>
          <CodeBlock label="project skill" code={skillInstall} />
          <p className="leading-relaxed text-muted">
            The file is also at <a className="text-ink underline" href="/skill/SKILL.md">/skill/SKILL.md</a>.
            Commit both copies if the whole team should get it. Claude Code reads{" "}
            <code className="font-mono text-ink">.claude/skills</code>. Cursor reads that folder and{" "}
            <code className="font-mono text-ink">.agents/skills</code>.
          </p>
        </section>

        <section id="scoring" className="space-y-4">
          <h2 className="font-serif text-3xl">Scoring</h2>
          <p className="leading-relaxed text-muted">
            Each answer is 0 (wrong or missing), 1 (partial), or 2 (clear). The quiz passes only when the average is at least{" "}
            <code className="font-mono text-ink">passing_average</code> (default 1.5) and no question scored 0.
            shipquiz checks this itself after the model replies, so one wrong answer still fails the quiz.
          </p>
          <p className="leading-relaxed text-muted">
            Questions have to be about this diff. They are not questions about syntax or style. Even a trivial change gets one question.
            The quiz is skipped only when the diff is empty, when it only touches lockfiles, or when every file matches{" "}
            <code className="font-mono text-ink">skip_paths</code>.
          </p>
        </section>

        <section id="source" className="space-y-4">
          <h2 className="font-serif text-3xl">From source</h2>
          <p className="leading-relaxed text-muted">
            From a checkout of this repo, with Go installed:
          </p>
          <CodeBlock label="build" code={sourceBuild} />
          <p className="leading-relaxed text-muted">
            <code className="font-mono text-ink">go test ./...</code> covers config merging, skipped paths, grading, the hook installer, and comment parsing.
            Point Vercel at the <code className="font-mono text-ink">website</code> directory to host this site.
            <code className="font-mono text-ink"> SHIPQUIZ_LLM_API_KEY</code> overrides the saved key. That is the model provider key CI sends as a repository secret.
          </p>
        </section>
      </div>
      </div>
    </main>
  );
}
