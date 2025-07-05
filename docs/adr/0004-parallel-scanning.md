# 4. Parallel Scanning Implementation

## Status
Accepted

## Context
Originally Eskimo cloned repositories and ran each configured scanner sequentially. This kept the code simple but made scans slow when organisations contain many repositories.

## Decision
We introduced concurrency to utilise all available CPU cores:

- Repositories are cloned in parallel using a worker pool limited by `runtime.NumCPU()`.
- After cloning completes, another pool scans repositories concurrently.
- Each repository run spawns goroutines for each enabled scanner.

This design keeps the code relatively small while significantly reducing overall scan time.

## Consequences
- Scan duration now scales with the number of available CPU cores.
- Output may interleave across goroutines, but failures are still logged per repository and scanner.
