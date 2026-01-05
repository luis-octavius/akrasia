[PT-BR Version](docs/README-pt.md)

### Akrasia
_Akrasía_ is a Greek word meaning "incontinence" or lack of self-control. This app helps you manage tasks you need to complete but tend to procrastinate on.  
As Plato wrote in _Laws_, humans are engaged in a never-ending internal war within their own souls — **a battle against pleasure-seeking**. Today, as we scroll through endless feeds of videos and posts, we chase instant gratification while neglecting the meaningful goals we should pursue.  
This app aims to help you regain self-control in your daily life.  

### Requirements 
- [Go 1.25 or later](https://go.dev/doc/install)
- SQLite3
- [Goose](https://pressly.github.io/goose/installation/)
- A terminal

### Installation 
1. Clone the repo
```bash
git clone git@github.com:luis-octavius/akrasia.git && cd akrasia
```

2. Create a sqlite3 database file and create a .env with a `DB_PATH` variable pointing to the file:
```bash
# I recommend to create the file in the root of app
touch akrasia.db
cat > .env << 'EOF'                                                                                                      
DB_PATH="./akrasia.db"
EOF
```

3. Do the up migrations with the `migrations_up.sh`:

```bash
chmod +x migrations_up.sh
./migrations_up.sh
```

4. Install it:
   
```bash
go install .
```

5. Use it:
   
```bash
akrasia add "Do homework" "Semantics - Philosophy of Language" "12 02"
```

6. (Optional) Create an alias:
```bash
echo "akr='akrasia'" >> ~/.zshrc # or .bashrc 
```

### Commands 
```
Usage:
  akrasia [command]

Available Commands:
  add              adds a todo in storage, description is optional
  check-expired    check expired todos
  check-expiring   check todos that are expiring in 5 days
  completion       Generate the autocompletion script for the specified shell
  delete-concluded delete all concluded todos
  get-all          returns all the todos saved in storage
  get-by-name      gets a todo by name
  help             Help about any command
  update-status    update concluded status to true

Flags:
  -h, --help   help for akrasia

Use "akrasia [command] --help" for more information about a command.
```

### Simple usage 
```bash
# add command
akrasia add "Stendhal" "Finish the book The Red and The Black" "13 02" # Finish in 13-Feb
akrasia add "Stendhal" "Finish the book The Red and The Black" "13" # Finish in 13 of the actual month
akrasia add "Stendhal" "Finish the book The Red and The Black" "" # Daily Todo
akrasia add "Stendhal" "Finish the book The Red and The Black" "13 02 20:00:00" # Finish in 13-Feb at 20 o'clock
2026/01/04 11:38:31 ✅ Todo Stendhal created successfully!

akrasia update-status "stendhal" # case-insensitive, mark Todo as done
📋 Stendhal | Finish the book The Red and The Black | 13 Feb 26 00:00 UTC | Done

akrasia get-all # 
🔔 Todos:

📋 Stendhal | Finish the book The Red and The Black | 13 Feb 26 00:00 UTC | Done

akrasia delete-concluded # self-explanatory
✅ Concluded Todos deleted successfully!
```

## Contributing
Contributions are welcome!.

## License
This project is licensed under the MIT License - see the [LICENSE](docs/LICENSE.md) file for details.

## Acknowledgments
- Inspired by ancient Greek philosophy on self-control
- Built with Go
