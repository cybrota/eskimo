# AGENTS Guidelines

This repository follows these guidelines for contributions by AI or human agents:

1. **Commit Messages**: Use [Conventional Commits](https://www.conventionalcommits.org/) format. Examples include:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `test:` for test-related changes
   - `chore:` for maintenance tasks

2. **Simplicity First**: Prefer simpler implementations over overly complex solutions.

3. **Run Tests**: Always run tests before committing to ensure functionality and catch regressions. Use `go test ./...` for Go modules.

4. **Uniform Structure**: Maintain a consistent code structure across modules so files and packages are easy to navigate.

5. **Explain Why**: Add comments explaining *why* something is done if it is not obvious from code alone.

6. **Copyright Header**: Add the following header at the beginning of every new code file:

```
Copyright (c) 2025 Naren Yellavula & Cybrota contributors
Apache License, Version 2.0

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
```

Follow these rules to keep the codebase easy to maintain.
