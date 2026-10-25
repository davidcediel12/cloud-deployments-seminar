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
| `step-3-k8s` | `/livez` + `/readyz` and graceful shutdown + Kubernetes YAML (Deployment with probes, Service, Secret) |
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
| `GET /api/message` | `{"message":"<APP_MESSAGE>","version":"<APP_VERSION>"}` |
| `GET /api/notes` | List of notes, newest first |
| `POST /api/notes` | Adds a note: `{"text":"..."}` |
| `GET /livez` | Liveness probe: 200 while the process runs |
| `GET /readyz` | Readiness probe: 200, or 503 while shutting down |

| Env variable | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | Port to listen on |
| `APP_MESSAGE` | `Hello from the API` | Text returned by `/api/message` |
| `APP_VERSION` | `dev` | Version shown by `/api/message`; CI sets it in the image to the commit SHA |
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

## 6. Kubernetes (from `step-3-k8s`, bonus reference)

The API on a Kubernetes cluster: `k8s/api.yaml` (Deployment with probes, LoadBalancer Service), `k8s/hpa.yaml`
(autoscaling between 2 and 4 pods) and a Secret with the database URL.

**You need:**
- **kubectl**, the Kubernetes command line: included in Docker Desktop and Codespaces; otherwise
  [install it](https://kubernetes.io/docs/tasks/tools/).
- **A cluster:** locally Docker Desktop (**Settings → Kubernetes → Enable Kubernetes**), or AKS in the portal
  (**Free** tier, 1 node, small VM size), then `az aks get-credentials -g <resource-group> -n <cluster>`.
  Check where kubectl points before applying: `kubectl config current-context`.
- **A Postgres the cluster can reach**, e.g. Azure Database for PostgreSQL (allow access from Azure services in
  its networking settings). Create the database first: the API creates its table, not the database.
- **Your image on Docker Hub:** `docker build -t <you>/seminar-api:3.1 ./api`, `docker push <you>/seminar-api:3.1`,
  and put it in `k8s/api.yaml` (`image:`).

**Deploy:**
```bash
cp k8s/secret.yaml.example k8s/secret.yaml   # fill in the URL; secret.yaml is never committed
kubectl apply -f k8s/         # secret.yaml, api.yaml, hpa.yaml (the .example file is skipped)
kubectl get pods            # wait for 2 pods READY 1/1
kubectl get svc api         # EXTERNAL-IP: the public IP (Docker Desktop: use localhost)
```
✅ `http://<EXTERNAL-IP>:8081/api/message` answers. For the frontend, set `apiUrl` in `config.js` to that address and
`CORS_ORIGIN` in `k8s/api.yaml` to the frontend's URL.

**Try:**
- Self-healing: `kubectl delete pod <name>`, then `kubectl get pods -w`: a new pod replaces it.
- Zero-downtime update: run the load generator (below), then `kubectl rollout restart deployment/api`: 0 errors.
- Autoscaling: `kubectl get hpa api -w` in one terminal, load in another:
  ```bash
  go run ./tools/load/main.go                                   # 20 workers, 60 s, localhost:8081
  go run ./tools/load/main.go -url http://<EXTERNAL-IP>:8081/api/message -c 40 -d 2m
  ```
  CPU goes above the 20% target and pods are added (up to 4); a minute after the load stops it goes back to 2.
  The load generator is optional and needs Go (Codespaces has it). Without it the autoscaler still runs: it keeps
  2 pods and `kubectl get hpa api` shows the current CPU against the target.
  The autoscaler reads CPU from **metrics-server**: AKS has it. **Docker Desktop only** (`kubectl top pods` says
  "Metrics API not available"): install it once. Without it nothing breaks, the API just stays at 2 pods.
  ```bash
  kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
  kubectl patch deployment metrics-server -n kube-system --type=json -p '[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]'
  ```
  (The flag skips checking Docker Desktop's self-signed certificate: local only. Windows PowerShell: write each `"`
  inside the JSON as `\"`, or use Git Bash.)
- Logs: `kubectl logs -l app=api`.

> ⚠️ The LoadBalancer serves plain **HTTP**. A frontend served over HTTPS (Blob Storage static website) can't call
> an HTTP API: the browser blocks it ("mixed content"). Show the AKS API directly (browser or curl); Container Apps
> doesn't have this problem because it provides HTTPS.

🧹 **When done:** delete the AKS cluster (its node and public IP consume credit).

## 7. Terraform: the frontend's hosting as code (from `step-4-terraform`)

`infra/terraform/` creates a resource group and a storage account with the static website turned on, the same thing
you would click in the portal. It creates the empty site; uploading the frontend is a separate step (and automated in
step 5).

**You need:**
- **Azure CLI**, logged in: [install](https://learn.microsoft.com/cli/azure/install-azure-cli), then `az login`.
  Terraform uses this login, no extra password.
- **Terraform**: [install](https://developer.hashicorp.com/terraform/install). Codespaces has both preinstalled (still run `az login`).

**Create it:**
```bash
cd infra/terraform
cp terraform.tfvars.example terraform.tfvars   # put your subscription ID in it: az account show --query id -o tsv
terraform init      # download the Azure plugin (once)
terraform plan      # preview: 4 resources to add, nothing changes yet
terraform apply     # create them (type "yes"); prints website_url and storage_account_name
```
✅ `website_url` opens a "WebContentNotFound" page: the site exists but is still empty.
Error that the region "is not accepting new customers" or a policy error? Set another `location` in `terraform.tfvars`.

**Upload the frontend** (from the repo root):
```bash
npm --prefix web ci
npm --prefix web run build
# edit web/dist/config.js: apiUrl = your API's public URL (e.g. the Container App URL)
az storage blob upload-batch --account-name <storage_account_name> -s web/dist -d '$web' --overwrite
```
✅ Reload `website_url`: the app loads. If it can't reach the API, set the API's `CORS_ORIGIN` to `website_url`
without the trailing `/`.

🧹 **When done:** `terraform destroy` (type "yes") deletes the resource group and everything in it.

## 8. CI/CD with GitHub Actions (from `step-5-github-actions`)

Two workflows in `.github/workflows/` run on every push to `main`:

| Workflow | When | Does |
|---|---|---|
| `api` | `api/` changed | Builds the image, pushes `<you>/seminar-api:<commit>` to Docker Hub, updates the Container App |
| `web` | `web/` changed | Builds the frontend, writes `config.js` with the API URL, uploads it to the static website |

Pull requests only build (nothing is deployed). Run a workflow by hand: **Actions → api / web → Run workflow**.

Workflows run in the repo you push to, so do this in **your own repo**.

**You need:** the Container App from Hands-on 2 (with `DB_URL` and `CORS_ORIGIN` set) and the static website from
section 7.

**1. Docker Hub token:** hub.docker.com → **Account settings → Personal access tokens → Generate**, access
*Read & Write*. Copy it (it's shown once).

**2. An Azure identity for the pipeline**, with no password: Azure trusts GitHub directly (OIDC). Replace the `<...>`:
```bash
az identity create -g <api-resource-group> -n id-github-deploy          # note the "clientId" and "principalId"

az identity federated-credential create -g <api-resource-group> --identity-name id-github-deploy -n github-main \
  --issuer https://token.actions.githubusercontent.com --audiences api://AzureADTokenExchange \
  --subject "repo:<user>@<user-id>/<repo>@<repo-id>:ref:refs/heads/main"

# Allow it to change the two resource groups (API and frontend), nothing else:
az role assignment create --role Contributor --assignee-object-id <principalId> --assignee-principal-type ServicePrincipal \
  --scope /subscriptions/<subscription-id>/resourceGroups/<api-resource-group>
az role assignment create --role Contributor --assignee-object-id <principalId> --assignee-principal-type ServicePrincipal \
  --scope /subscriptions/<subscription-id>/resourceGroups/<web-resource-group>
```
The `--subject` must match exactly. `<user-id>` and `<repo-id>` are numbers: open
`https://api.github.com/repos/<user>/<repo>` and copy `owner.id` and `id`. (Many tutorials show the older form
without the IDs; it doesn't match anymore.)

**3. Settings in GitHub:** your repo → **Settings → Secrets and variables → Actions**.

| Tab | Name | Value |
|---|---|---|
| Secrets | `DOCKERHUB_TOKEN` | the token from step 1 |
| Variables | `DOCKERHUB_USERNAME` | your Docker Hub user |
| Variables | `AZURE_CLIENT_ID` | `clientId` from step 2 |
| Variables | `AZURE_TENANT_ID` | `az account show --query tenantId -o tsv` |
| Variables | `AZURE_SUBSCRIPTION_ID` | `az account show --query id -o tsv` |
| Variables | `AZURE_RESOURCE_GROUP` | the API's resource group |
| Variables | `CONTAINER_APP_NAME` | e.g. `seminar-api` |
| Variables | `STORAGE_ACCOUNT_NAME` | `storage_account_name` from Terraform |
| Variables | `API_URL` | the Container App URL, without the trailing `/` |

**4. Try it:** change something in `api/` (or use **Run workflow**) and push to `main`.
✅ In **Actions**, `build` then `deploy` turn green, and `<API URL>/api/message` shows `"version"`: the commit you
just pushed.

**If it fails:**
- `Variable ... is not set`: add it on the **Variables** tab (exact name, not a secret).
- `AADSTS700213: No matching federated identity record`: the subject in step 2 doesn't match. The error shows the
  subject GitHub sent: create the credential with exactly that text.
- `AuthorizationFailed`: the role assignment is missing or still propagating (wait a minute, then **Re-run jobs**).

⭐ **AKS:** CI deploys only to Container Apps. To run the same image on AKS, put the tag in `k8s/api.yaml`
(`image: <you>/seminar-api:<commit>`) and `kubectl apply -f k8s/`: build once, deploy anywhere.
