# Backlog — api_faker KYC mock path

## Entregas e ordem de execução

### Nível 0 — paralelas (sem dependências)
| ID | Título | Arquivos principais |
|----|--------|---------------------|
| d1 | Mover rota mock para fora do subrouter JWT | `router.go`, `mock.go` |
| d2 | Corrigir chave JWT no middleware | `middleware.go`, `cmd/main.go` |
| d8 | Makefile build/test + README | `Makefile`, `README.md` |

### Nível 1 — após d1
| ID | Título | Depende de |
|----|--------|------------|
| d3 | Path pattern matching com segmentos nomeados | d1 |
| d5 | POST/GET /api/scenario | d1 |

### Nível 2 — após d3
| ID | Título | Depende de |
|----|--------|------------|
| d4 | Interpolação de templates no modelo de resposta | d3 |

### Nível 3 — após d5
| ID | Título | Depende de |
|----|--------|------------|
| d6 | Seed do cenário de KYC (scripts + make seed-kyc) | d5 |

### Nível 4 — após tudo
| ID | Título | Depende de |
|----|--------|------------|
| d7 | Testes automatizados (primeira suite) | d1, d2, d3, d4, d5 |

---

## Grafo resumido

```
d1 ──┬── d3 ── d4 ──┐
     └── d5 ── d6   ├── d7
d2 ──────────────────┘
d8 (independente)
```

## Notas
- Repo não tem nenhum `*_test.go`. d7 é a primeira suite — toda mudança entra com teste.
- `go` não está instalado na máquina host. Build e test via container (d8 prepara isso).
- Seção 9 do documento (latência configurável, headers, matching por body, mudanças em infrapay/connector) está fora de escopo.
