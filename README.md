# Akrasía

Akrasía is a simple CLI tool to keep track of your tasks. The application was designed to help combat [Akrasía](https://en.wikipedia.org/wiki/Akrasia) - the lack of self-control that leads to acting against one's better judgment.

## Requirements
- [Go](https://go.dev/dl/) installed

## Installation

1. Clone the repository:
```bash
git clone git@github.com:luis-octavius/todo-terminal.git
cd akrasia
```

2. Build and install:
```bash
go install
```

3. Run the application:
```bash
akrasia
```

## Features

| Command | Usage | Description |
|---------|-------|-------------|
| `add` | `add "study for algorithms proof"` | Add a new todo with the specified description |
| `get-all` | `get-all` | Display all todos in the database |
| `get` | `get 1` | Show details of the todo with ID 1 |
| `delete` | `delete 1` | Remove the todo with ID 1 |
| `help` | `help` | Display available commands |
| `exit` | `exit` | Close the application |

## Future improvements
### Data Persistence
- Add database support to persist data between sessions
- Implement data export/import functionality
- Add automatic backup system

### User Experience
- Enhance UI readability with colors and better structure
- Add keyboard shortcuts for frequent actions
- Implement command autocompletion

### Task Management
- Support for different task types: daily, weekly, monthly
- Add due dates and reminders for tasks
- Implement task prioritization system
- Add task categories/tags

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

## License
Copyright (c) 2024 Luis Octavius

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
