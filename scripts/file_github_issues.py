#!/usr/bin/env python3
"""Open the repo's good-first-issue files on GitHub.

Requires `gh auth login` and a GitHub remote named origin. The script creates
the labels, then runs `gh issue create` once per file in
docs/contributing/issues.
"""

from __future__ import annotations

import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
ISSUE_DIR = ROOT / "docs" / "contributing" / "issues"

LABELS = {
    "help wanted": ("008672", "Extra attention is needed"),
    "good first issue": ("7057ff", "Good for newcomers"),
}


def run(args: list[str], check: bool = True) -> subprocess.CompletedProcess[str]:
    return subprocess.run(args, cwd=ROOT, text=True, capture_output=True, check=check)


def parse_issue(path: Path) -> tuple[str, list[str], str]:
    text = path.read_text()
    if not text.startswith("---\n"):
        sys.exit(f"{path} is missing frontmatter")
    end = text.find("\n---\n", 4)
    if end < 0:
        sys.exit(f"{path} frontmatter does not close")
    header = text[4:end]
    body = text[end + 5 :].strip() + "\n"
    title = ""
    labels: list[str] = []
    in_labels = False
    for line in header.splitlines():
        if line.startswith("title:"):
            title = line.split(":", 1)[1].strip().strip('"')
            in_labels = False
        elif line.startswith("labels:"):
            in_labels = True
        elif in_labels and line.strip().startswith("- "):
            labels.append(line.strip()[2:].strip())
        else:
            in_labels = False
    if not title or not labels:
        sys.exit(f"{path} needs a title and labels")
    return title, labels, body


def main() -> int:
    auth = run(["gh", "auth", "status"], check=False)
    if auth.returncode != 0:
        sys.stderr.write(auth.stderr or auth.stdout)
        sys.stderr.write(
            "\nGitHub auth is not usable. Run `gh auth login`, then rerun this script.\n"
        )
        return 1

    for name, (color, description) in LABELS.items():
        exists = run(["gh", "label", "list", "--search", name], check=False)
        if exists.returncode != 0:
            sys.stderr.write(exists.stderr)
            return exists.returncode
        if name not in exists.stdout.splitlines() and not any(
            line.split("\t", 1)[0] == name for line in exists.stdout.splitlines() if line
        ):
            created = run(
                [
                    "gh",
                    "label",
                    "create",
                    name,
                    "--color",
                    color,
                    "--description",
                    description,
                ],
                check=False,
            )
            if created.returncode != 0 and "already exists" not in created.stderr:
                sys.stderr.write(created.stderr)
                return created.returncode

    files = sorted(ISSUE_DIR.glob("*.md"))
    if not files:
        sys.exit(f"no issues in {ISSUE_DIR}")
    for path in files:
        title, labels, body = parse_issue(path)
        cmd = ["gh", "issue", "create", "--title", title, "--body", body]
        for label in labels:
            cmd.extend(["--label", label])
        created = run(cmd, check=False)
        if created.returncode != 0:
            sys.stderr.write(created.stderr)
            return created.returncode
        sys.stdout.write(created.stdout)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
