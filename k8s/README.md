# Kubernetes manifests (optional)

These manifests are optional helpers if you are managing configuration alongside this repo.

## Backend

- `leapmailr-backend-configmap.yaml`: non-secret configuration
- `leapmailr-backend-secret.example.yaml`: **example only** (do not apply as-is)
- `leapmailr-backend-sealedsecret.example.yaml`: **example only** (commit the real sealed secret, not a plaintext secret)

### Create the real Secret

Copy the example and fill in real values:

- Copy `leapmailr-backend-secret.example.yaml` → `leapmailr-backend-secret.yaml`
- Apply it:

```bash
kubectl apply -n leapmailr -f k8s/leapmailr-backend-secret.yaml
kubectl apply -n leapmailr -f k8s/leapmailr-backend-configmap.yaml
```

Recommended for GitOps: use Sealed Secrets / External Secrets instead of committing a plaintext Secret.

## Using Sealed Secrets (recommended)

This workflow lets you commit an encrypted `SealedSecret` to git, and have it materialize into a normal `Secret` at runtime.

### 1. Install the Sealed Secrets controller

Install it once per cluster (namespace commonly `sealed-secrets`). The exact install method depends on your platform; the most common is the Bitnami chart.

### 2. Install `kubeseal` (client)

You run `kubeseal` locally/CI to encrypt the Secret for your cluster.

### 3. Generate and commit a SealedSecret for the backend

Create a plaintext Secret manifest locally (do **not** commit it):

```bash
kubectl create secret generic leapmailr-backend-secrets \
	-n leapmailr \
	--from-literal=DB_PASSWORD='REPLACE_ME' \
	--from-literal=JWT_SECRET='REPLACE_ME' \
	--from-literal=ENCRYPTION_KEY='REPLACE_ME_32_CHARS' \
	--from-literal=NR_LICENSE_KEY='' \
	--dry-run=client -o yaml > /tmp/leapmailr-backend-secret.yaml
```

Seal it for your cluster and write the file you *will* commit:

```bash
# Option A: Use the provided script (recommended)
./scripts/seal-secret.sh \
	/tmp/leapmailr-backend-secret.yaml \
	k8s/leapmailr-backend-sealedsecret.yaml

# Option B: Use kubeseal directly
kubeseal \
	--format yaml \
	< /tmp/leapmailr-backend-secret.yaml \
	> k8s/leapmailr-backend-sealedsecret.yaml
```

Apply (or let GitOps apply) the SealedSecret + ConfigMap:

```bash
kubectl apply -n sealed-secrets -f <sealed-secrets-controller-install.yaml>
kubectl apply -n leapmailr -f k8s/leapmailr-backend-sealedsecret.yaml
kubectl apply -n leapmailr -f k8s/leapmailr-backend-configmap.yaml
```

Notes:
- The created `Secret` name must match what the Helm chart references (`leapmailr-backend-secrets`).
- Sealed secrets are cluster-specific: re-seal after migrating clusters.
