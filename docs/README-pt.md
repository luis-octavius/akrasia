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

4. (Opcional Linux) Crie um alias:

```
echo "akr='akrasia'" >> ~/.zshrc # ou .bashrc

```

### Comandos
```text
Uso:
  akrasia [comando]

Comandos disponíveis:
  add              adiciona uma tarefa no armazenamento, descrição é opcional
  check-expired    verifica tarefas expiradas
  completion       gera script de autocompletar para o shell especificado
  delete-concluded exclui todas as tarefas concluídas
  get-all          retorna todas as tarefas salvas no armazenamento
  get-by-name      obtém uma tarefa pelo nome
  help             ajuda sobre qualquer comando
  init             inicializa o aplicativo
  update-status    atualiza status de concluído para verdadeiro

Flags:
  -h, --help   ajuda para akrasia

Use "akrasia [comando] --help" para mais informações sobre um comando.


```

### Uso simples

```bash
# comando add
akrasia add --name Stendhal --desc "Terminar o livro O Vermelho e o Negro" --date 13,02
2026/01/04 11:38:31 Tarefa Stendhal criada com sucesso!

akrasia update-status --name stendhal # case-insensitive, marca tarefa como concluída
Stendhal | Terminar o livro O Vermelho e o Negro |
13 Fev 26 00:00 UTC | Done 

akrasia get-all # 
Todos:

Stendhal | Terminar o livro O Vermelho e o Negro |
13 Fev 26 00:00 UTC | Concluída

akrasia delete-concluded # autoexplicativo
Tarefas concluídas excluídas com sucesso!

```  

### Contribuindo
Contribuições são bem-vindas! Se gostaria de contribuir, apenas faça o fork do projeto e abra uma pull request na main.

### Licença
Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENÇA](LICENSE.md) para detalhes.

### Reconhecimentos
- Inspirado pela filosofia grega antiga sobre autocontrole 
- Construído com Go

