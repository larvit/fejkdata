#!/usr/bin/env python3
"""Rewrite data files to the one spelling each shape has.

    reshape.py data/sv_SE/person.json ...

A one-item choice becomes its item; an object holding only a format becomes that
string; weight 1 and repeat 1 are dropped; a repeated choice item becomes one
item with a weight.
"""
import json
import sys

OPTIONS = ("format", "weight", "repeat", "separator")
WIDTH = 110


def reshape(n):
	if isinstance(n, list):
		items = [reshape(x) for x in n]
		if len(items) == 1:
			return items[0]
		out, at = [], {}
		for x in items:
			key = json.dumps(x, sort_keys=True)
			if key in at:
				i = at[key]
				if isinstance(out[i], str):
					out[i] = {"format": out[i], "weight": 2}
				elif isinstance(out[i], dict):
					out[i]["weight"] = out[i].get("weight", 1) + 1
				else:
					out.append(x)
				continue
			at[key] = len(out)
			out.append(x)
		return out
	if isinstance(n, dict):
		d = {k: (v if k in OPTIONS else reshape(v)) for k, v in n.items()}
		if d.get("weight") == 1:
			del d["weight"]
		if d.get("repeat") == 1:
			del d["repeat"]
		if list(d) == ["format"]:
			return d["format"]
		return d
	return n


def scalar(v):
	return json.dumps(v, ensure_ascii=False)


def one_line(n):
	if isinstance(n, dict):
		return "{ " + ", ".join(scalar(k) + ": " + one_line(v) for k, v in n.items()) + " }"
	if isinstance(n, list):
		return "[" + ", ".join(one_line(x) for x in n) + "]"
	return scalar(n)


def dump(n, indent=0):
	pad = "  " * indent
	if isinstance(n, dict):
		line = one_line(n)
		if len(pad) + len(line) <= WIDTH and not any(isinstance(v, dict) for v in n.values()):
			return line
		body = ",\n".join(pad + "  " + scalar(k) + ": " + dump(v, indent + 1) for k, v in n.items())
		return "{\n" + body + "\n" + pad + "}"
	if isinstance(n, list):
		if all(isinstance(x, str) for x in n):
			return one_line(n)
		line = one_line(n)
		if len(pad) + len(line) <= WIDTH:
			return line
		body = ",\n".join(pad + "  " + dump(x, indent + 1) for x in n)
		return "[\n" + body + "\n" + pad + "]"
	return scalar(n)


if __name__ == "__main__":
	for path in sys.argv[1:]:
		with open(path) as fh:
			before = json.load(fh)
		after = dump(reshape(before)) + "\n"
		with open(path, "w") as fh:
			fh.write(after)
		print("reshaped", path)
