# circular-dependency-checker

Detects cyclic dependencies in Godot projects that cause "bad address index" errors.

## Quick start

```bash
python circular_dependency_checker.py <path-to-godot-project>
```

## Why this exists

In GDScript, `preload()` calls at the top of a script are resolved at parse time. If script A preloads script B and B preloads A (directly or transitively), Godot raises a cryptic **"bad address index"** error at runtime — not a clear circular-import message. This tool walks all `.gd` files, builds a directed dependency graph from `class_name`, `extends`, and type annotations, then reports any cycles before you hit that error in the engine.

## Sample output

```
Checking /path/to/my-godot-project...
Cycle detected: PlayerState -> StateMachine -> PlayerState
Cycle detected: Enemy -> EnemyAI -> BaseAI -> Enemy
2 cycle(s) found. Fix preload order or use load() with lazy initialization.
```

If no cycles are found:
```
No circular dependencies detected.
```

## License

Apache-2.0 — see [LICENSE](LICENSE).
