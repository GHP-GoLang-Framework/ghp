#!/usr/bin/env python3
"""Enforce the PR-issues policy on a GitHub pull request.

Every PR must close at least one issue; closing several is allowed, and the
release takes the highest version bump among them. The pure functions are
kept separate from the ``gh`` calls so they can be tested with table-driven
fixtures without network access. Dependabot is the only exemption: its
dependency PRs have no issue and are classified as a patch bump.
"""

import json
import re
import subprocess
import sys
from typing import Iterable

# GitHub reports bot logins either as "dependabot[bot]" or "app/dependabot"
# depending on the endpoint; a substring check covers both forms.
DEPENDABOT = "dependabot"

# Only close keywords create a GitHub issue link. "Refs" is a mention and
# must not satisfy the policy.
CLOSE_KEYWORDS = r"closes|close|closed|fixes|fixed|resolves|resolved"
CLOSE_LINE = re.compile(
    rf"(?im)^[ \t>]*(?:{CLOSE_KEYWORDS})[ \t]+#(\d+)\b"
)

# Precedence used by both this check and the release job: the highest label
# wins, it is never a sum of bumps.
BUMP_RANK = {"major": 3, "minor": 2, "patch": 1}


def linked_issue_numbers(body: str) -> list[int]:
    """Return issue numbers closed by explicit keywords in a PR body."""
    return [int(number) for number in CLOSE_LINE.findall(body or "")]


def version_bumps(labels: Iterable[str]) -> list[str]:
    """Extract version:* labels, keeping only the three supported names."""
    return [
        label.split(":", 1)[1]
        for label in labels
        if label.startswith("version:") and label.split(":", 1)[1] in BUMP_RANK
    ]


def highest_bump(version_labels: dict[int, list[str]]) -> str | None:
    """Return the single bump a release should apply, or None if unversioned."""
    bumps = [bump for labels in version_labels.values() for bump in version_bumps(labels)]
    if not bumps:
        return None
    return max(bumps, key=lambda bump: BUMP_RANK[bump])


def validate(
    author: str,
    body: str,
    linked: Iterable[int],
    version_labels: dict[int, list[str]],
) -> list[str]:
    """Return policy violations for a PR; an empty list means it passes."""
    if DEPENDABOT in author.lower():
        return []

    linked = sorted(set(linked))
    errors: list[str] = []

    if not linked:
        errors.append("expected at least one linked issue, GitHub reports none")

    explicit = set(linked_issue_numbers(body))
    if not explicit:
        errors.append("the PR body must contain at least one 'Closes #N' line")
    elif explicit != set(linked):
        errors.append(
            f"the PR body closes {sorted(explicit)} but GitHub links {linked}"
        )

    for number in linked:
        bumps = version_bumps(version_labels.get(number, []))
        if len(bumps) != 1:
            errors.append(
                f"issue #{number} must have exactly one version:* label; "
                f"found {len(bumps)}"
            )

    return errors


def gh_json(*args: str):
    """Run ``gh`` with args and decode its JSON output."""
    result = subprocess.run(
        ["gh", *args], check=True, capture_output=True, text=True
    )
    return json.loads(result.stdout)


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: check_pr_issue.py <pull-request-number>", file=sys.stderr)
        return 2

    pr = gh_json(
        "pr",
        "view",
        sys.argv[1],
        "--json",
        "author,body,closingIssuesReferences",
    )
    author = (pr.get("author") or {}).get("login") or ""
    linked = [issue["number"] for issue in pr.get("closingIssuesReferences") or []]

    version_labels: dict[int, list[str]] = {}
    for number in linked:
        issue = gh_json("issue", "view", str(number), "--json", "labels")
        version_labels[number] = [
            label["name"] for label in issue.get("labels") or []
        ]

    errors = validate(author, pr.get("body") or "", linked, version_labels)
    if errors:
        print("::error::PR issues policy failed:")
        for error in errors:
            print(f"- {error}")
        print(
            "Close at least one issue with 'Closes #N', give every linked "
            "issue exactly one version:* label, or use a Dependabot PR for "
            "dependency updates."
        )
        return 1

    if "dependabot" in author.lower():
        print("Dependabot PR: exempt from the issue-link rule; bump = patch.")
    else:
        bump = highest_bump(version_labels)
        print(
            f"PR closes {linked}; highest version bump = {bump}."
        )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
