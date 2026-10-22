## Quick Start

1. Clone the repo:

```bash
git clone git@github.com:luis-octavius/akrasia.git && cd akrasia
```

2. Install it (the binary goes to `$(go env GOPATH)/bin`, which should be on your `PATH`):

```bash
go install .
```

3. Initialize the database:

```bash
akrasia init
```

4. Open the interactive interface:

```bash
akrasia
```

Press `a` to add your first task and `?` to see every key. Prefer commands? `akrasia add "Morning run" --daily` (or `--weekly --on fri`, `--monthly --on 5`) works too, and `akrasia help` lists all of them.

5. (Optional) Create a shorter alias:

```bash
echo "alias akr='akrasia'" >> ~/.zshrc # or ~/.bashrc
```

### Running with Docker

```bash
docker build -t akrasia .
docker run -it --rm -v akrasia-data:/data akrasia init
docker run -it --rm -v akrasia-data:/data akrasia
```

The database lives in the `akrasia-data` volume (`DB_PATH=/data/akrasia.db`). The image sets `TZ=America/Sao_Paulo`; change it with `-e TZ=<your/zone>` so "today" matches your day.
