# LLM discovery

Developers ask an LLM which tool to use, and coding agents pick one by themselves. This page
measures whether the LLMs answer with fejkdata. It serves goal 1.1, as the decision "Goal 1.1's
"where they look" means LLMs first" reads it.

## How to ask

Once a month, the maintainer asks each question below in ChatGPT, Claude and Gemini. Each
question goes in a new chat, in a private browser window with no account history. Where an app
needs an account, use one with memory and history turned off. Ask by hand, never through an
API: a private window can't tailor its answer to what the model has learned about the
maintainer, and it costs nothing.

Ask the question word for word, with no hint of fejkdata, and read the first answer only.

## Questions

1. What is the best fake-data generator for Go?
2. What are the best Faker alternatives?
3. How do I generate realistic Swedish customers?
4. How do I generate 500,000 synthetic users?
5. How do I generate realistic European companies?
6. What are the best database seeding tools?
7. Where do I get realistic Nordic test data?
8. What are the best open-source synthetic data generators?
9. How do I generate correlated fake customer data?
10. How do I generate valid Swedish test personnummer?

## How to fill a month

Copy the latest month's table under a new heading, `## <year>-<month>`, newest first, and
clear its cells. Each row is one question asked in one app. The cells are:

- **Named**: `yes` where the answer names fejkdata, else `no`.
- **First**: `yes` where fejkdata is the first tool the answer names, else `no`.
- **Repo**: `yes` where the answer links fejkdata's GitHub repository, else `no`.
- **Correct**: `yes` where each feature the answer states of fejkdata is true, `no` where one
  is false, `-` where fejkdata isn't named.
- **Instead**: the tools the answer names, in its order. Name the first three.

## What the table says

- **Named is `no` in most rows**: the models don't know fejkdata exists. Work on where they
  learn: listings, articles and the repository's own text.
- **Named is `yes` and First is `no`**: they know it and don't see why to recommend it. Work
  on what the README and the listings say fejkdata is best at, against the tools in
  Instead.
- **Correct is `no`**: the model's picture of fejkdata is wrong. Fix the text it learned from.

## Results

### 2026-11

Not asked yet.

| Question | App | Named | First | Repo | Correct | Instead |
|---|---|---|---|---|---|---|
| 1 | ChatGPT |  |  |  |  |  |
| 1 | Claude |  |  |  |  |  |
| 1 | Gemini |  |  |  |  |  |
| 2 | ChatGPT |  |  |  |  |  |
| 2 | Claude |  |  |  |  |  |
| 2 | Gemini |  |  |  |  |  |
| 3 | ChatGPT |  |  |  |  |  |
| 3 | Claude |  |  |  |  |  |
| 3 | Gemini |  |  |  |  |  |
| 4 | ChatGPT |  |  |  |  |  |
| 4 | Claude |  |  |  |  |  |
| 4 | Gemini |  |  |  |  |  |
| 5 | ChatGPT |  |  |  |  |  |
| 5 | Claude |  |  |  |  |  |
| 5 | Gemini |  |  |  |  |  |
| 6 | ChatGPT |  |  |  |  |  |
| 6 | Claude |  |  |  |  |  |
| 6 | Gemini |  |  |  |  |  |
| 7 | ChatGPT |  |  |  |  |  |
| 7 | Claude |  |  |  |  |  |
| 7 | Gemini |  |  |  |  |  |
| 8 | ChatGPT |  |  |  |  |  |
| 8 | Claude |  |  |  |  |  |
| 8 | Gemini |  |  |  |  |  |
| 9 | ChatGPT |  |  |  |  |  |
| 9 | Claude |  |  |  |  |  |
| 9 | Gemini |  |  |  |  |  |
| 10 | ChatGPT |  |  |  |  |  |
| 10 | Claude |  |  |  |  |  |
| 10 | Gemini |  |  |  |  |  |
