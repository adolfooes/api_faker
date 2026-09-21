# Convenções de Branch e PR

## Fonte

`gh` não disponível neste ambiente. Convenções inferidas do histórico de commits via `git log`.

## Rulesets ativos

⚠️ gh não disponível ou falhou — rulesets não detectados via API GitHub.

## Fluxo de PR inferido

Commits seguem Conventional Commits (`feat:`, `fix:`, `chore:`, `feat(scope):`, `fix(scope):`).
Único branch observado no histórico: `main`. Sem branches de feature no log local.

```
baseBranch: main
branchTemplate: {KEY}-{slug}
prTarget: main
```

## Campo git no manifest

```json
{
  "baseBranch": "main",
  "branchTemplate": "{KEY}-{slug}",
  "prTarget": "main"
}
```
