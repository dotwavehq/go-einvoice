#!/usr/bin/env python3
"""Validate a CII e-invoice against EN 16931 / XRechnung schematron (compiled XSLT).

Usage: validate.py <invoice.xml> <stylesheet.xslt> [<stylesheet.xslt> ...]

Runs each stylesheet via Saxon (saxonche, XSLT 2.0) and reports SVRL failed
asserts. Exits non-zero if any assert has flag "fatal" or "error".
"""
import re
import sys

from saxonche import PySaxonProcessor

FATAL = {"fatal", "error", ""}


def failed_asserts(svrl):
    for m in re.finditer(r"<svrl:failed-assert\b([^>]*)>(.*?)</svrl:failed-assert>", svrl, re.S):
        attrs, body = m.group(1), m.group(2)
        flag = (re.search(r'flag="([^"]*)"', attrs) or [None, ""])[1]
        text = re.search(r"<svrl:text>(.*?)</svrl:text>", body, re.S)
        text = re.sub(r"\s+", " ", text.group(1)).strip() if text else ""
        yield flag, text


def main():
    xml, stylesheets = sys.argv[1], sys.argv[2:]
    blocking = 0
    with PySaxonProcessor(license=False) as proc:
        xp = proc.new_xslt30_processor()
        for sheet in stylesheets:
            exe = xp.compile_stylesheet(stylesheet_file=sheet)
            svrl = exe.transform_to_string(source_file=xml)
            for flag, text in failed_asserts(svrl):
                marker = "FAIL" if flag in FATAL else "warn"
                if flag in FATAL:
                    blocking += 1
                print(f"[{marker}:{flag or 'error'}] {text}")
    if blocking:
        print(f"\n{blocking} blocking violation(s)", file=sys.stderr)
        return 1
    print("\nOK: no blocking violations")
    return 0


if __name__ == "__main__":
    sys.exit(main())
