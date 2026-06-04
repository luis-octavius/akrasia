# Akrasia

_Akrasía_ é uma palavra grega que significa "incontinência" ou falta de autocontrole. Este aplicativo ajuda você a gerenciar tarefas que precisa completar mas tende a procrastinar.
Como Platão escreveu em _Leis_, os humanos estão engajados em uma guerra interna interminável dentro de suas próprias almas — uma batalha contra a busca por prazer. Hoje, enquanto rolamos feeds infinitos de vídeos e posts, perseguimos gratificação instantânea enquanto negligenciamos os objetivos significativos que deveríamos buscar.
Este aplicativo visa ajudá-lo a recuperar o autocontrole em sua vida diária.

### Motivação

Sim, há inúmeros aplicativos para controlar tarefas, mas em minha experiência usando esses aplicativos, descobri que nenhum deles realmente sanava todas as minhas necessidades. Eu tentei vários aplicativos de produtividade, cada um com diferentes métodos para organizar tarefas e enviar lembretes. Ainda assim, nenhum deles manteve meu engajamento por mais de um curto período de tempo.

Esse abismo entre as ferramentas disponíveis e o meu fluxo de trabalho pessoal me levou a desenvolver meu próprio aplicativo. Eu queria uma ferramenta que seria intuitiva, alinhada ao modo como penso produtividade, e que motivasse o bastante para uso consistente.

### Requisitos

- Go 1.25 ou superior
- SQLite3
- Goose
- Um terminal

### Instalação

1. Clone o repositório

```bash
git clone git@github.com:luis-octavius/akrasia.git && cd akrasia
```

2. Instale

```bash
go install .
```

3. Inicialize o banco de dados

```bash
akrasia init
```

4. Crie o cron job para automaticamente atualizar tarefas diárias:

```bash
akrasia create-cron # ou cc
```

5. (Opcional) Crie um alias:

```
echo "akr='akrasia'" >> ~/.zshrc # ou .bashrc

```

### Comandos

```text
Uso:
  akrasia [comando]

Comandos disponíveis:
  add               cria uma nova tarefa com descrição, prioridade e data opcionais
  backfill-history  preenche histórico ausente de tarefas diárias
  check-expired     exibe tarefas que já passaram da data de expiração
  check-expiring    verifica tarefas que expiram em 5 dias
  completion        gera script de autocompletar para o shell especificado
  create-cron       cria o cron job para atualizar tarefas diárias
  delete-by-name    exclui um todo pelo nome
  delete-concluded  exclui todas as tarefas concluídas
  focus             exibe de 1 a 3 tarefas prioritárias para foco imediato
  get-all           retorna todas as tarefas salvas no armazenamento
  get-by-name       busca uma tarefa pelo nome (case-insensitive e fuzzy)
  get-daily         exibe todas as tarefas diárias
  help              ajuda sobre qualquer comando
  history           exibe a história de ofensivas da tarefa buscada
  init              inicializa o aplicativo
  streak            exibe a ofensiva atual da tarefa buscada
  today             exibe um painel de foco diário (vencidas, hoje, diárias pendentes e em breve)
  update-daily      atualiza tarefas diárias (usado por cron, não é destinado ao uso manual)
  update-status     marca tarefa como concluída e registra no histórico com notas opcionais

Flags:
  -h, --help   ajuda para akrasia

Use "akrasia [comando] --help" para mais informações sobre um comando.


```

### Uso simples

```bash
# comando add
akrasia add --name Stendhal --desc "Terminar o livro O Vermelho e o Negro" --date 13,02
2026/01/04 11:38:31 Tarefa Stendhal criada com sucesso!

# adiciona uma tarefa diária
akrasia add --name "Corrida matinal" --daily --priority high
2026/03/08 08:15:20 Tarefa Corrida matinal criada com sucesso!

akrasia update-status --name stendhal # case-insensitive, marca tarefa como concluída
Stendhal | Terminar o livro O Vermelho e o Negro |
13 Fev 26 00:00 UTC | Concluída

akrasia get-all --priority high #
Todos:

Stendhal | Terminar o livro O Vermelho e o Negro |
13 Fev 26 00:00 UTC | Concluída

akrasia delete-concluded --yes # autoexplicativo (exige confirmação)
Tarefas concluídas excluídas com sucesso!

# preenche o histórico para os últimos 30 dias (padrão)
akrasia backfill-history
Preenchido 90 entradas de histórico para 3 tarefa(s) diária(s)

# preenche o histórico para uma tarefa específica e 60 dias atrás
akrasia backfill-history --task "Corrida matinal" --days 60
Preenchido 60 entradas de histórico para a tarefa 'Corrida matinal'

# vê o que precisa de atenção hoje
akrasia today --only overdue --limit 5 --priority high
TODAY FOCUS

OVERDUE (1)
Pagar conta de luz | ...

DUE TODAY (2)
Preparar relatório semanal | ...

DAILY PENDING (1)
Corrida matinal | ...

# escolhe os itens principais para executar agora
akrasia focus --limit 3 --priority high
FOCUS (3)
Pagar conta de luz | ...
Preparar relatório semanal | ...
Corrida matinal | ...

# obtém a ofensiva atual para uma tarefa diária
akrasia streak --name "Corrida matinal"
A sua ofensiva atual com Corrida matinal é: 5

# obtém o histórico de ofensivas
akrasia history --name "Corrida matinal"
1. Data de Início: 2026-02-15 | Data de Término: 2026-02-20 | Total de Dias: 6
2. Data de Início: 2026-01-10 | Data de Término: 2026-01-15 | Total de Dias: 6

```

### Contribuindo

Contribuições são bem-vindas! Se gostaria de contribuir, apenas faça o fork do projeto e abra uma pull request na main.

### Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENÇA](LICENSE.md) para detalhes.

### Reconhecimentos

- Inspirado pela filosofia grega antiga sobre autocontrole
- Construído com Go
