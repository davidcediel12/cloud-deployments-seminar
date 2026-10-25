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

**API** (`api/`, Go standard library): `GET /api/message` → `{"message":"<APP_MESSAGE>"}`.

| Env variable | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | Port to listen on |
| `APP_MESSAGE` | `Hello from the API` | Text returned by `/api/message` |
| `CORS_ORIGIN` | `http://localhost:5173` | Frontend origin allowed to call the API (in the cloud: the frontend's URL) |

**Frontend** (`web/`, React + Vite): shows the API's message. The API URL is read at runtime from `public/config.js`
(`dist/config.js` after building), so the same build works against any API.

Without Docker (optional): in `api/` run `PORT=8081 go run .` (PowerShell: `$env:PORT=8081; go run .`), and in `web/` run
`npm ci && npm run dev`.

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
