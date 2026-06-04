## Quick Start

1. Clone the repo

```bash
git clone git@github.com:luis-octavius/akrasia.git && cd akrasia
```

2. Install it:

```bash
go install .
```

3. Initialize database:

```bash
akrasia init
```

4. Create cron job to update daily tasks automatically:

```bash
akrasia create-cron # or cc
```

5. (Optional) Create an alias:

```bash
echo "akr='akrasia'" >> ~/.zshrc # or .bashrc
```

