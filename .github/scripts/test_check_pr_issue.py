"""Table-driven tests for the PR issues policy validation."""

import unittest

from check_pr_issue import highest_bump, linked_issue_numbers, validate


class ValidateTest(unittest.TestCase):
    def test_cases(self):
        cases = [
            {
                "name": "one linked versioned issue passes",
                "author": "Shimigui",
                "body": "## Summary\n\nCloses #112\n",
                "linked": [112],
                "labels": {112: ["Feature", "version:patch"]},
                "want": [],
            },
            {
                "name": "several linked issues pass",
                "author": "Shimigui",
                "body": "Closes #12\nCloses #13\n",
                "linked": [12, 13],
                "labels": {12: ["version:minor"], 13: ["version:patch"]},
                "want": [],
            },
            {
                "name": "no linked issue fails",
                "author": "Shimigui",
                "body": "## Summary\n\nNo issue here.\n",
                "linked": [],
                "labels": {},
                "want": ["at least one linked issue", "at least one 'Closes #N'"],
            },
            {
                "name": "linear-style ref does not satisfy the policy",
                "author": "Shimigui",
                "body": "Refs: GHP-123\n",
                "linked": [],
                "labels": {},
                "want": ["at least one linked issue", "at least one 'Closes #N'"],
            },
            {
                "name": "body closes a different set than GitHub links",
                "author": "Shimigui",
                "body": "Closes #7\n",
                "linked": [9],
                "labels": {9: ["version:patch"]},
                "want": ["closes [7] but GitHub links [9]"],
            },
            {
                "name": "linked issue without a version label fails",
                "author": "Shimigui",
                "body": "Closes #12\n",
                "linked": [12],
                "labels": {12: ["Feature"]},
                "want": ["issue #12 must have exactly one version:* label; found 0"],
            },
            {
                "name": "one of several linked issues lacks a version label",
                "author": "Shimigui",
                "body": "Closes #12\nCloses #13\n",
                "linked": [12, 13],
                "labels": {12: ["version:minor"], 13: ["Feature"]},
                "want": ["issue #13 must have exactly one version:* label; found 0"],
            },
            {
                "name": "linked issue with two version labels fails",
                "author": "Shimigui",
                "body": "Closes #12\n",
                "linked": [12],
                "labels": {12: ["version:major", "version:patch"]},
                "want": ["issue #12 must have exactly one version:* label; found 2"],
            },
            {
                "name": "dependabot is exempt",
                "author": "app/dependabot",
                "body": "Bumps golang.org/x/term.\n",
                "linked": [],
                "labels": {},
                "want": [],
            },
            {
                "name": "dependabot[bot] spelling is exempt",
                "author": "dependabot[bot]",
                "body": "",
                "linked": [],
                "labels": {},
                "want": [],
            },
        ]

        for case in cases:
            with self.subTest(case["name"]):
                errors = validate(
                    case["author"], case["body"], case["linked"], case["labels"]
                )
                if case["want"]:
                    self.assertEqual(len(errors), len(case["want"]))
                    for expected, error in zip(case["want"], errors):
                        self.assertIn(expected, error)
                else:
                    self.assertEqual(errors, [])


class LinkedIssueNumbersTest(unittest.TestCase):
    def test_close_keywords(self):
        body = "Closes #12\nfixes #13\nresolves #14\nRefs #15\n"
        self.assertEqual(linked_issue_numbers(body), [12, 13, 14])

    def test_quoted_close_line(self):
        self.assertEqual(linked_issue_numbers("> Closes #21\n"), [21])


class HighestBumpTest(unittest.TestCase):
    def test_cases(self):
        cases = [
            ({"12": ["version:patch"]}, "patch"),
            (
                {12: ["version:minor"], 13: ["version:patch"]},
                "minor",
            ),
            (
                {12: ["version:patch"], 13: ["version:major"], 14: ["version:minor"]},
                "major",
            ),
            ({12: ["Feature", "version:minor"]}, "minor"),
            ({12: ["Feature"]}, None),
        ]
        for labels, want in cases:
            with self.subTest(labels=labels):
                # The first case uses string keys only to prove the helper
                # ignores the key type; all real callers pass integers.
                self.assertEqual(highest_bump(labels), want)


if __name__ == "__main__":
    unittest.main()
