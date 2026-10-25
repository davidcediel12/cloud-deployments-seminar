# Cloud Deployments Seminar: starter repo

A tiny app that we take from code to the cloud during the seminar: container → registry → Azure → automation → operations.

```
web/   React (Vite) frontend → static files → Azure Blob Storage static website
api/   Go REST API          → container    → Azure Container Apps
```

This repo is teaching material. For the assignment you create **your own repo**.

## 1. Setup (before the seminar)

**Accounts:** **Docker Hub**, and a **cloud account**. The seminar shows Azure: **Azure for Students** is the easiest
(free, no credit card); a personal Azure account works too. For the assignment, any cloud is allowed.

> ⚠️ Want Azure for Students? Activate it **first**: activating the regular Azure free trial with your university
> email blocks Azure for Students afterwards.

**Option A: your machine (recommended).** Install Docker Desktop (or Docker Engine) and Git, then:

```bash
git clone https://github.com/davidcediel12/cloud-deployments-seminar.git
cd cloud-deployments-seminar
git switch step-0-hello-world
docker pull golang:1.26 gcr.io/distroless/static-debian13:nonroot   # base images for Hands-on 1 (large, download them now)
```

**Option B: GitHub Codespaces** (if you can't install Docker): a cloud machine with this repo, opened in VS Code in
the browser, with Docker, Go, Node and the Azure CLI preinstalled. On GitHub: **Code → Codespaces → ⋯ → New with
options**, branch `step-0-hello-world`. It uses your personal GitHub quota: stop it when you're done.

## 2. Steps: one branch each

| Branch | Adds |
|---|---|
| `step-0-hello-world` | The app + a TODO skeleton for `api/Dockerfile` (**start here**) |
| `step-1-containerized` | Working Dockerfiles for `api` and `web` (**rescue for Hands-on 1**) |
| `step-2-compose-db` | Notes feature + Postgres, `compose.yaml` to run everything locally |
| `step-3-k8s` | `/health` endpoint + Kubernetes YAML (Deployment with probes, Service, Secret) |
| `step-4-terraform` | Terraform for the frontend's storage account (static website) |
| `step-5-github-actions` | CI/CD workflows for `web` and `api` |
| `main` | Final state (same as step 5) |
| `extra-observability` | Side branch from step 2: OpenTelemetry + Grafana LGTM, local demo |

See what a step changed: `git diff step-1-containerized step-2-compose-db`, or on GitHub
[compare step-1...step-2](https://github.com/davidcediel12/cloud-deployments-seminar/compare/step-1-containerized...step-2-compose-db).

Jump to a step: `git stash` (puts your edits aside; `git stash pop` brings them back), then `git switch <branch>`.

## 3. The app

**API** (`api/`, Go + Postgres driver `pgx`):

| Endpoint | Does |
|---|---|
| `GET /api/message` | `{"message":"<APP_MESSAGE>"}` |
| `GET /api/notes` | List of notes, newest first |
| `POST /api/notes` | Adds a note: `{"text":"..."}` |

| Env variable | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | Port to listen on |
| `APP_MESSAGE` | `Hello from the API` | Text returned by `/api/message` |
| `CORS_ORIGIN` | `http://localhost:5173` | Frontend origin allowed to call the API (in the cloud: the frontend's URL) |
| `DB_URL` | (required) | Postgres connection string: `postgres://user:password@host:5432/db` |

**Frontend** (`web/`, React + Vite): shows the API's message and the notes. The API URL is read at runtime from
`public/config.js` (`dist/config.js` after building), so the same build works against any API.

**Run everything** (from `step-2-compose-db` on):

```bash
docker compose up --build    # API on :8081, web on http://localhost:3000, Postgres inside
docker compose down          # stop; notes are kept in the "dbdata" volume
docker compose down -v       # stop and delete the volume: notes are gone
```

## 4. Hands-on 1: containerize the API (15 min)

1. **Write the Dockerfile (7 min):** open `api/Dockerfile` and replace each `TODO` with one instruction.
2. **Build and run (4 min):**
   ```bash
   docker build -t api ./api
   docker run --rm -p 8081:8080 api   # --rm: delete the container when it stops · -p host port:container port
   ```
   ✅ The terminal shows a JSON log line `"msg":"starting API"`.
   Error `port is already allocated` or `bind: Only one usage of each socket address`? Another program uses 8081 on
   your machine: change only the host side, e.g. `-p 8082:8080`, and use 8082 below.
3. **Try it (4 min):** open http://localhost:8081/api/message (Codespaces: **Ports** tab → 8081). Stop it (Ctrl+C) and
   inject config from outside:
   ```bash
   docker run --rm -p 8081:8080 -e APP_MESSAGE="Hi from my container" api   # -e: set an environment variable
   ```
   ✅ The message changed: same image, different config.

⭐ **Stretch:** `docker images` and compare the size of `api` with `golang`. Why so different?

🛟 **Rescue:** `git stash`, `git switch step-1-containerized`, build again.

## 5. Hands-on 2: deploy the API to Azure Container Apps (20 min)

Replace `<you>` with your Docker Hub username.

1. **Push the image to Docker Hub (5 min):**
   ```bash
   docker login
   docker build -t <you>/seminar-api:1.0 ./api
   docker push <you>/seminar-api:1.0
   ```
   ✅ The image is on hub.docker.com (public by default, so Azure can pull it).
   Apple-silicon Mac (M1 and later)? Add `--platform linux/amd64` to `docker build`: Azure runs Intel/AMD machines,
   and an ARM image fails there.
2. **Create the Container App in the Azure Portal (10 min):** search **Container Apps** → **Create**.
   The portal may offer a newer **express** mode with fewer settings: either works, pick what you prefer. The steps
   below follow the classic form.
   - **Basics:** your subscription, new resource group `rg-seminar`, name `seminar-api`, a region, new Container Apps
     environment. Policy error about the region? Pick another: student subscriptions allow only some.
   - **Container:** uncheck *Use quickstart image*; source **Docker Hub or other registries**, type **Public**,
     registry `docker.io`, image `<you>/seminar-api:1.0`.
   - **Ingress:** enabled, traffic from **anywhere**, target port **8080**: the container port (Azure maps public HTTPS to it, like `-p` does locally).
   - **Review + create.**
3. **Open it (5 min):** resource → **Overview** → **Application Url** → add `/api/message`.
   ✅ Your API answers on a public HTTPS URL. (The first request can take a few seconds: scale-to-zero cold start.)

⭐ **Stretch:** **Containers** → edit → add env var `APP_MESSAGE` → save → a new **revision** → reload.
(Express mode keeps a single revision: the change still shows up after reload.)

🛟 **Rescue:** use the image `davidcediel12/seminar-api:1.0`.

🧹 **When done:** delete the resource group (**Resource groups → rg-seminar → Delete**) to save credit.

## 6. Observability: logs, metrics, traces (from `extra-observability`, local demo)

The API sends its telemetry with **OpenTelemetry** (`api/telemetry.go`) to Grafana's all-in-one
[`grafana/otel-lgtm`](https://github.com/grafana/docker-otel-lgtm) container: an OpenTelemetry collector plus
**L**oki (logs), **G**rafana, **T**empo (traces) and **M**etrics (Prometheus). Demo setup, not for production.

```bash
git switch extra-observability
docker pull grafana/otel-lgtm:0.38.0   # 2.75 GB: do it before the demo
docker compose up --build              # API :8081, web :3000, Grafana http://localhost:3001 (no login)
```

Use the app at http://localhost:3000 (add a few notes), then in Grafana open **Dashboards → Seminar API**:

| Panel | Shows | Comes from |
|---|---|---|
| Requests / responses / latency | Rate per route, per status code, p50/p95 | `otelhttp` middleware: automatic for every request |
| Notes created | Custom counter | `notesCreated.Add(...)` in `api/notes.go` |
| Logs | One line per request | `slog`, sent to Loki with the request's `trace_id` |
| Recent traces | Click a trace ID: the request and its SQL query (`query INSERT`) below it | `otelhttp` + `otelpgx` |

- **Logs → traces:** expand a log line; its `trace_id` belongs to one request, which you can open in Tempo.
- **Your own queries:** **Explore → Prometheus**, e.g.
  `sum by (http_route) (rate(http_server_request_duration_seconds_count[1m]))`.
- **Config** is the standard `OTEL_*` variables in `compose.yaml` (`OTEL_EXPORTER_OTLP_ENDPOINT`,
  `OTEL_SERVICE_NAME`). Without `OTEL_EXPORTER_OTLP_ENDPOINT` the API runs as before, without telemetry.
- **Dashboard as code:** `observability/grafana/seminar-api.json`. Changed it in the UI and want to keep it?
  **Export → JSON** over that file, then `docker compose restart lgtm` (Grafana reads it at startup).

Gotchas:

- New traces take up to ~30 s to appear in search; metrics and logs within ~5 s.
- Telemetry data lives in the container: `docker compose down` deletes it (the dashboard comes back on the next start).
- Dashboard empty although the app works, or `port is already allocated` on 8081? Another API answers on 8081
  (e.g. the Kubernetes one from step 3, without telemetry): stop it first.
