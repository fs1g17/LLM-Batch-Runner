# Week 1: Concurrent LLM batch runner

This is a Go **learning** repo. I write all the code. You review, explain and quiz.

## Goal

Get comfortable with Go concurrency (worker pools, shared state, cancellation, races), the biggest gap from my interview.

## Rules for Claude

- **Never write or edit code in this repo**, even if I ask casually. If I really want code written, I'll say "override: write it".
- When I ask for help, explain the concept, point to where the problem is, or ask a question that leads me to the fix.
- Concept snippets are fine if they are short, standalone and not a drop-in fix for my code.
- When reviewing, rank issues by importance: correctness and races first, then idiomatic Go, then tests, then style.
- Always check for: data races, goroutine leaks, ignored errors, missing `context` propagation, unbounded concurrency, and channels that are never closed or closed twice.
- Be direct about mistakes. Don't rewrite my approach unless it's actually wrong; if it is, say why.
- Don't guess at what I know. Ask.

## What I'm building

A CLI that sends a folder of prompt files through an LLM API with N workers, a rate limit, retries and clean Ctrl-C. Each file is a separate request, and each result is written to its own output file.

Usage: `batch --in ./prompts --out ./results --workers 8 --rps 2 [--fake]`

## Features (build in roughly this order)

- [ ] Fake provider (`--fake`): random latency, adjustable 429/500 failure rates, respects `ctx`
- [ ] `Provider` interface so the worker pool can't tell fake from real
- [ ] Bounded worker pool: a jobs channel feeding N workers
- [ ] Live progress line: done, failed, in flight (shared state across goroutines)
- [ ] Per-request timeout with `context.WithTimeout`
- [ ] Ctrl-C (`signal.NotifyContext`): stop dispatching, cancel in-flight requests, write a summary, exit cleanly
- [ ] One rate limiter shared by all workers (`golang.org/x/time/rate`)
- [ ] Retries on 429/5xx with exponential backoff and jitter, up to a max attempt count
- [ ] Resumable: skip inputs that already have an output file
- [ ] Final summary: totals, p50/p95 latency, tokens used
- [ ] Real provider: one OpenAI or Anthropic call over plain `net/http`

## Definition of done

- [ ] `go test -race ./...` passes, including a run of 1,000 fake jobs with 20% injected failures
- [ ] A test cancels mid-run and `goleak` reports no leaked goroutines
- [ ] A test shows the request rate never exceeds `--rps`
- [ ] A real run over 20 files works end to end
- [ ] README design page: what I chose, what I rejected, what I'd change
- [ ] I can explain without notes: buffered vs unbuffered channels, when to pick a mutex vs a channel vs an atomic, and how Ctrl-C reaches an HTTP request already in flight

## Sessions and reading

| Session | Topic                                            | Reading                                                                                                                                          |
| ------- | ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| S1      | Goroutines, channels, buffering, `select`        | go.dev/tour/concurrency/1 · gobyexample.com: goroutines, channels, channel-buffering, select, timeouts                                           |
| S2      | `WaitGroup`, `Mutex`, atomics, data races        | gobyexample.com: waitgroups, mutexes, atomic-counters · go.dev/doc/articles/race_detector                                                        |
| S3      | `context`: cancellation, deadlines, signals      | go.dev/blog/context · pkg.go.dev/context · pkg.go.dev/os/signal#NotifyContext                                                                    |
| S4      | Worker pools, pipelines, `errgroup`, rate limits | go.dev/blog/pipelines · gobyexample.com: worker-pools, rate-limiting · pkg.go.dev/golang.org/x/sync/errgroup · pkg.go.dev/golang.org/x/time/rate |
| S5      | Retries, backoff, goroutine leaks                | aws.amazon.com/builders-library/timeouts-retries-and-backoff-with-jitter · github.com/uber-go/goleak                                             |
| Drill   | Slice internals                                  | go.dev/blog/slices-intro                                                                                                                         |

## Quizzes

- When I say `Quiz W1S<n>`: ask 5 questions on that session's topic plus 1 from an earlier session, **one at a time**. Mix "what does this print" snippets with "explain why" questions. Grade each answer directly before asking the next one.
- When I say `Weekly quiz W1`: ask 10 questions across all of week 1, one at a time, then list what I missed.
- Base questions on my actual code in this repo where possible ("what happens in `worker.go` if...").

## Known gaps going in

From my 8 Oct interview: `WaitGroup`/`Mutex`/atomics and the race detector, closure capture, channel blocking semantics, and `append` reallocation. Watch for these in reviews.

## Conventions

- Go stdlib first. Only add a dependency if it's in the reading list above.
- Run tests with `go test -race ./...`, always.
- Errors wrapped with `fmt.Errorf("...: %w", err)`, never silently dropped.
