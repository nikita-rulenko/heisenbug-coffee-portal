# CLAUDE.md

Этот репозиторий использует открытый формат AGENTS.md для описания
контекста проекта. Claude Code по умолчанию подхватывает CLAUDE.md,
поэтому здесь — тонкий мост.

Основной контекст проекта: @AGENTS.md

## Архитектура контекста (5-way split)

- `AGENTS.md` — кросс-подходные правила для агента (стиль ответа,
  дисциплина «источник не истина», reply discipline).
- `docs/*.md` — факты о самом проекте
  (`project_overview.md`, `test-context.md`, `test-index.md`,
  `test-patterns.md`, `known_issues.md`).
- `.cursor/rules/*.mdc` — правила работы с конкретными инструментами
  (Mem0, Helixir, GitHub, архитектура, тестирование).
- `.claude/skills/<name>/` — роль и методология. Основной skill
  онбординга: `.claude/skills/qa-onboarding/SKILL.md` + варианты в
  `references/<подход>.md`.
- `prompts/show_difference_v2/*.md` — запуск конкретной задачи:
  роль, задача, формат отчёта, привязка к skill.

## Claude-специфичные дельты

- Четыре промта онбординга лежат в `prompts/show_difference_v2/`:
  `md_onb.md`, `mem0_onb.md`, `helixir_onb.md`, `git_iss_onb.md`.
  Каждый — тонкая обёртка, делегирующая методологию в skill
  `qa-onboarding`.
- Claude Code автоматически подхватывает skills из `.claude/skills/`
  по YAML-фронтматтеру. Промту достаточно назвать skill по имени —
  reference-файл (`references/<подход>.md`) читается агентом под
  задачу.
- Правила работы с конкретными MCP-серверами:
  - Mem0 — `.cursor/rules/mem0.mdc`
  - Helixir — `.cursor/rules/helixir.mdc`
  - GitHub — `.cursor/rules/github.mdc`
  - Архитектура и тестирование — `.cursor/rules/architecture.mdc`,
    `.cursor/rules/testing.mdc`
