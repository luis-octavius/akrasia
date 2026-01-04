# Akrasia

_Akrasía_ é uma palavra grega que significa "incontinência" ou falta de autocontrole. Este aplicativo ajuda você a gerenciar tarefas que precisa completar mas tende a procrastinar.
Como Platão escreveu em _Leis_, os humanos estão engajados em uma guerra interna interminável dentro de suas próprias almas — uma batalha contra a busca por prazer. Hoje, enquanto rolamos feeds infinitos de vídeos e posts, perseguimos gratificação instantânea enquanto negligenciamos os objetivos significativos que deveríamos buscar.
Este aplicativo visa ajudá-lo a recuperar o autocontrole em sua vida diária.

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

2. Crie um arquivo de banco de dados sqlite3 e um arquivo .env com a variável DB_PATH apontando para o arquivo:

```bash
# Recomendo criar o arquivo na raiz do aplicativo
touch akrasia.db
cat > .env << 'EOF'
DB_PATH="./akrasia.db"
EOF
```

3. Execute as migrações com migrations_up.sh:

```bash
chmod +x migrations_up.sh
./migrations_up.sh

```

4. Instale o aplicativo:

```
go install .
```

5. Use-o:

```bash
akrasia add "Fazer lição" "Semântica - Filosofia da Linguagem" "12 02"

```

6. (Opcional) Crie um alias:

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
  update-status    atualiza status de concluído para verdadeiro

Flags:
  -h, --help   ajuda para akrasia

Use "akrasia [comando] --help" para mais informações sobre um comando.


```

### Uso simples

```bash
# comando add
akrasia add "Stendhal" "Terminar o livro O Vermelho e o Negro" "13 02" # Terminar até 13-Fev
akrasia add "Stendhal" "Terminar o livro O Vermelho e o Negro" "13" # Terminar até dia 13 do mês atual
akrasia add "Stendhal" "Terminar o livro O Vermelho e o Negro" "" # Tarefa diária (sem prazo)
akrasia add "Stendhal" "Terminar o livro O Vermelho e o Negro" "13 02 20:00:00" # Terminar até 13-Fev às 20:00
2026/01/04 11:38:31 ✅ Tarefa Stendhal criada com sucesso!

akrasia update-status "stendhal" # case-insensitive, marca tarefa como concluída
📋 Stendhal | Terminar o livro O Vermelho e o Negro | 13 Fev 26 00:00 UTC | Concluída

akrasia get-all # 
🔔 Tarefas:

📋 Stendhal | Terminar o livro O Vermelho e o Negro | 13 Fev 26 00:00 UTC | Concluída

akrasia delete-concluded # autoexplicativo
✅ Tarefas concluídas excluídas com sucesso!

```  

### Contribuindo
Contribuições são bem-vindas!

### Licença
Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENÇA](LICENSE.md) para detalhes.

