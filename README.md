# Akrasia  

<img src="https://i.imgur.com/tWFhpXs.gif" />
    
_Akrasía_ is a Greek word meaning "incontinence" or lack of self-control. This app helps you manage tasks you need to complete but tend to procrastinate on.   
As Plato wrote in _Laws_, humans are engaged in a never-ending internal war within their own souls — **a battle against pleasure-seeking**. Today, as we scroll through endless feeds of videos and posts, we chase instant gratification while neglecting the meaningful goals we should pursue.    
This app aims to help you regain self-control in your daily life.  

## Motivation
Yes, there are many applications for task management out there, but in my experience using them, I found that none truly met my needs. I have tried numerous productivity apps, each with different approaches to organizing tasks and sending reminders. Yet, none retained my engagement for more than a short period.

This gap between available tools and my personal workflow led me to develop my own application. I wanted a tool that was intuitive, aligned with how I think about productivity, and motivating enough to use consistently.

## Requirements 
- [Go 1.25 or later](https://go.dev/doc/install)
- SQLite3
- [Goose](https://pressly.github.io/goose/installation/)
- A terminal

## Quick Start
1. Clone the repo
```bash
git clone git@github.com:luis-octavius/akrasia.git && cd akrasia
```

2. Install it:
   
```bash
go install .
```

3. Initialize the database with the command:
   
```bash
akrasia init 
```

4. (Optional) Create an alias:
```bash
echo "akr='akrasia'" >> ~/.zshrc # or .bashrc 
```

## Commands 
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
  init             initialize akrasia app                                  
  update-status    update concluded status to true 

Flags:
  -h, --help   help for akrasia

Use "akrasia [command] --help" for more information about a command.
```

## Usage 
```bash
# add command
akrasia add --name Stendhal --desc "Finish the book The Red and The Black" 13 
2026/01/04 11:38:31 Todo Stendhal created successfully!

akrasia update-status --name stendhal # case-insensitive, mark Todo as done
Stendhal | Finish the book The Red and The Black |
13 Feb 26 00:00 UTC | Done

akrasia get-all # 
Todos:

Stendhal | Finish the book The Red and The Black |
13 Feb 26 00:00 UTC | Done

akrasia delete-concluded # self-explanatory
Concluded Todos deleted successfully!
```

## Contributing
Contributions are welcome!
If you're willing to contribute, just fork the project and open a pull request at main.

## License
This project is licensed under the MIT License - see the [LICENSE](docs/LICENSE.md) file for details.

## Acknowledgments
- Inspired by ancient Greek philosophy on self-control
- Built with Go

[PT-BR Version](docs/README-pt.md)
