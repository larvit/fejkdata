#!/usr/bin/env python3
"""Rewrite data files from the class-char format grammar to the literal-text one.

    convert.py data/sv_SE/person.json ...

Old: 0 1 A a are character classes, # escapes.  New: text is literal; a run of
0/A/a becomes {digits(n)}/{upper(n)}/{lower(n)}, 1 followed by n zeros becomes
{int(10^n, 10^(n+1)-1)}, a literal brace becomes {{ or }}.
"""
import re
import sys

RUN = {"0": "{{digits({})}}", "A": "{{upper({})}}", "a": "{{lower({})}}"}


def literal(c):
	return {"{": "{{", "}": "}}"}.get(c, c)


def convert_format(f):
	out, i, n = [], 0, len(f)
	while i < n:
		c = f[i]
		if c == "#":
			i += 1
			if i < n:
				out.append(literal(f[i]))
				i += 1
			else:
				out.append("#")
		elif c == "{":
			j = f.find("}", i)
			if j < 0:
				out.append(f[i:])
				break
			out.append(f[i:j + 1])
			i = j + 1
		elif c in RUN:
			k = 0
			while i < n and f[i] == c:
				k += 1
				i += 1
			out.append(RUN[c].format(k))
		elif c == "1":
			i += 1
			k = 0
			while i < n and f[i] == "0":
				k += 1
				i += 1
			out.append("{{int({},{})}}".format(10 ** k, 10 ** (k + 1) - 1))
		else:
			out.append(literal(c))
			i += 1
	return "".join(out)


FORMAT_VALUE = re.compile(r'("format":\s*")((?:[^"\\]|\\.)*)(")')


def convert_text(text):
	return FORMAT_VALUE.sub(lambda m: m.group(1) + convert_format(m.group(2)) + m.group(3), text)


if __name__ == "__main__":
	for path in sys.argv[1:]:
		with open(path) as fh:
			before = fh.read()
		after = convert_text(before)
		if after != before:
			with open(path, "w") as fh:
				fh.write(after)
			print("converted", path)
