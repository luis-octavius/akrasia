# Akrasia - Roadmap de Issues e Melhorias

## Overview
Documento para rastrear e organizar as issues e melhorias necessárias no projeto.

---

## Issues Principais

### 1. Adicionar Parâmetros Posicionais ao Comando `add`
**Status:** ✅ Concluído  
**Descrição:** Simplificar a criação de tarefas permitindo argumentos posicionais  
**Motivação:** Melhorar UX ao criar tarefas - usuários podem digitar `akrasia add "Task name"` em vez de `akrasia add --name "Task name"`

**Subtarefas:**
- [x] Analisar assinatura atual do comando `add` em [commands.go](internal/commands/commands.go)
- [x] Implementar suporte a nome como primeiro argumento posicional
- [x] Implementar suporte a description como segundo argumento posicional (opcional)
- [x] Atualizar exemplos na documentação
- [x] Testar compatibilidade com flags existentes

**Notas:**
- O comando atual aceita até 3 args (`Args: cobra.MaximumNArgs(3)`) mas NÃO usa nenhum
- Flags atuais: `--name` (obrigatória), `--desc`, `--date`, `--daily`, `--priority`
- Proposta: argumentos posicionais têm prioridade sobre flags para backward compatibility
- Exemplos desejados:
  - `akrasia add "Morning run"` - nome posicional
  - `akrasia add "Morning run" "Do 5km" --daily --priority high` - nome + desc
  - `akrasia add --name "Morning run" --daily` - modo antigo (continua funcionando)

**Implementação:**
- ✅ Alterado `Use` para `add [<name> [<description>]]`
- ✅ Reduzido `Args` para `cobra.MaximumNArgs(2)` (apenas 2 args)
- ✅ Removido `MarkFlagRequired("name")` - agora nome pode vir de posicional ou flag
- ✅ Adicionada lógica que prioriza args posicionais sobre flags
- ✅ Adicionada validação que garante nome seja fornecido (de um jeito ou outro)

---

### 2. Melhorar Acessibilidade de Quotes Longos
**Status:** ✅ Concluído  
**Descrição:** Quotes muito longos quebram de forma estranha no terminal  
**Motivação:** Melhorar legibilidade e acessibilidade em diferentes tamanhos de terminal

**Subtarefas:**
- [x] Identificar comportamento de quebra atual em [quotes.go](internal/tasks/quotes.go)
- [x] Implementar word-wrapping apropriado para o terminal
- [x] Adicionar sistema de detecção de largura do terminal
- [x] Testar em diferentes tamanhos de terminal (80, 120, 160 colunas)
- [x] Validar que o formatação funciona com a função `MsgQuote` do pacote color

**Notas:**
- Quotes gerados em: [generateRandomQuote()](internal/tasks/quotes.go#L80)
- Função de renderização: [color.MsgQuote()](pkg/color/color.go#L24)
- Usar biblioteca `golang.org/x/term` para detectar tamanho do terminal

**Implementação:**
- ✅ Adicionada função `getTerminalWidth()` para detectar largura do terminal
- ✅ Adicionada função `wrapText()` que quebra texto em palavras inteiras (não no meio)
- ✅ Implementado fallback para 80 colunas se detectar falhar
- ✅ Deixado margem de 4 colunas para padding/formatação
- ✅ Integrado word-wrapping na função `generateRandomQuote()`
- ✅ Testado com sucesso - quotes agora quebram adequadamente

---

### 3. Sistema de Temas de Cores com Alto Contraste
**Status:** ✅ Concluído  
**Descrição:** Transformar colors simples em sistema configurável com temas (incluindo alto contraste)  
**Motivação:** Melhor acessibilidade e customização visual

**Subtarefas:**
- [x] Analisar estrutura atual do pacote color em [pkg/color/color.go](pkg/color/color.go)
- [x] Desenhar arquitetura do sistema de temas:
  - [x] Definir tipos/structs para representar temas
  - [x] Decidir onde armazenar configuração de tema (arquivo JSON em `~/.config/akrasia/`)
- [x] Implementar temas:
  - [x] Tema `default` (cores atuais)
  - [x] Tema `high-contrast` (para acessibilidade)
- [x] Criar comando `config` para gerenciar temas
  - [x] `akrasia config theme <name>` - definir tema
  - [x] `akrasia config theme list` - listar temas disponíveis
  - [x] `akrasia config theme show` - exibir tema atual
- [x] Atualizar todas as chamadas de color functions com o novo sistema
- [x] Testar com screen readers/ferramentas de acessibilidade

**Notas:**
- Funções atuais: `MsgError()`, `MsgWarning()`, `MsgSuccess()`, `MsgQuote()`
- Biblioteca usada: `github.com/fatih/color`
- Local de armazenamento: `~/.config/akrasia/theme.json` (respeita `XDG_CONFIG_HOME`)

**Implementação - Arquitetura:**
- ✅ Criado arquivo [pkg/color/themes.go](pkg/color/themes.go) com estruturas de temas
- ✅ Definidos tipos: `Theme` e `ColorStyle`
- ✅ Implementados 2 temas: `default` e `high-contrast`
- ✅ Funções: `SaveTheme()`, `GetCurrentTheme()`, `GetAvailableThemes()`
- ✅ Sistema de carregamento automático de tema na inicialização
- ✅ Atualizado [pkg/color/color.go](pkg/color/color.go) para usar temas
- ✅ Temas sendo aplicados em `MsgError()`, `MsgWarning()`, `MsgSuccess()`, `MsgQuote()`

**Implementação - Comando Config:**
- ✅ Criado comando `config` em [internal/commands/commands.go](internal/commands/commands.go)
- ✅ Criado subcomando `config theme` com suporte a:
  - ✅ `akrasia config theme list` - lista temas disponíveis
  - ✅ `akrasia config theme show` - mostra tema atual
  - ✅ `akrasia config theme <name>` - define novo tema
- ✅ Tema persiste em `~/.config/akrasia/theme.json`
- ✅ Testado com sucesso - alternância entre temas funciona

---

## Template para Novas Issues

```markdown
### N. Descrição Breve
**Status:** ⏳ Não iniciado  
**Descrição:** [Descrição detalhada]  
**Motivação:** [Por que fazer isso]

**Subtarefas:**
- [ ] Tarefa 1
- [ ] Tarefa 2

**Notas:**
- Nota 1
- Nota 2
```

---

## Legenda de Status
- ⏳ Não iniciado
- 🚧 Em progresso
- ✅ Concluído
- 🔧 Em revisão
- ⚠️ Bloqueado

---

## Histórico de Alterações

| Data | Alteração |
|------|-----------|
| 2026-06-04 | ✅ Concluídas todas as 3 issues principais! |
| 2026-06-04 | ✅ Issue 1: Argumentos posicionais para comando `add` |
| 2026-06-04 | ✅ Issue 2: Word-wrapping para quotes (com detecção de terminal) |
| 2026-06-04 | ✅ Issue 3: Sistema de temas com alto contraste + comando `config` |
