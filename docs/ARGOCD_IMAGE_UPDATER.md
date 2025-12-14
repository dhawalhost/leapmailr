# ArgoCD Image Updater (recommended)

This repo’s GitHub Actions pipelines build/push container images only.
CD is handled by ArgoCD + Helm, with **ArgoCD Image Updater** automatically updating image tags in your Helm repo.

## Images produced by CI

- Backend: `ghcr.io/<OWNER>/leapmailr`
- Frontend: `ghcr.io/<OWNER>/leapmailr-ui`

Tagging strategy (recommended for Image Updater):
- release tags (semver): `vMAJOR.MINOR.PATCH` (example: `v1.2.3`)
- optional: `latest` for convenience
- optional: immutable SHA tags: `sha-<12-hex>`

## Install ArgoCD Image Updater

Install ArgoCD Image Updater in the same cluster/namespace as ArgoCD (common: `argocd`).

Key recommendations:
- Use **Git write-back** so Image Updater commits updated image tags into your Helm Git repo.
- Use a dedicated Git deploy key or PAT with least privilege.
- Prefer **semver** if you release with version tags; otherwise use **newest-build** with an allow-tags regexp.

## Configure your ArgoCD Application

In your *Helm repo* (or GitOps repo), add annotations to the ArgoCD `Application` that deploys LeapMailR.

### Option A (recommended): track release tags via semver

This updates `values.yaml` (or whatever file your app uses) by changing the image tag field.

Example `Application` annotations (put under `metadata.annotations`):

```yaml
metadata:
  annotations:
    argocd-image-updater.argoproj.io/image-list: |
      backend=ghcr.io/<OWNER>/leapmailr
      frontend=ghcr.io/<OWNER>/leapmailr-ui

    # Only consider release tags produced by CI
    argocd-image-updater.argoproj.io/backend.allow-tags: regexp:^v[0-9]+\.[0-9]+\.[0-9]+$
    argocd-image-updater.argoproj.io/frontend.allow-tags: regexp:^v[0-9]+\.[0-9]+\.[0-9]+$

    # Pick the highest semver tag
    argocd-image-updater.argoproj.io/backend.update-strategy: semver
    argocd-image-updater.argoproj.io/frontend.update-strategy: semver

    # Write back into Git (Helm repo)
    argocd-image-updater.argoproj.io/write-back-method: git

    # Tell Image Updater where the tag lives in Helm values
    # For this repo's chart at helm/generic-app:
    argocd-image-updater.argoproj.io/backend.helm.image-tag: workloads.backend.image.tag
    argocd-image-updater.argoproj.io/frontend.helm.image-tag: workloads.frontend.image.tag

    # Optional: keep commits tidy
    argocd-image-updater.argoproj.io/git-commit-message: "chore(images): bump {{ .AppName }}"
```

Your Helm values should look like (example only):

```yaml
workloads:
  backend:
    image:
      repository: ghcr.io/<OWNER>/leapmailr
      tag: v0.0.0
  frontend:
    image:
      repository: ghcr.io/<OWNER>/leapmailr-ui
      tag: v0.0.0
```

### Option B: track immutable tags via regexp (no semver)

If you prefer CI to publish immutable tags like `sha-<12-hex>`, use `newest-build` + an allow-tags regexp.

```yaml
metadata:
  annotations:
    argocd-image-updater.argoproj.io/image-list: |
      backend=ghcr.io/<OWNER>/leapmailr
      frontend=ghcr.io/<OWNER>/leapmailr-ui

    argocd-image-updater.argoproj.io/backend.allow-tags: regexp:^sha-[0-9a-f]{12}$
    argocd-image-updater.argoproj.io/frontend.allow-tags: regexp:^sha-[0-9a-f]{12}$

    argocd-image-updater.argoproj.io/backend.update-strategy: newest-build
    argocd-image-updater.argoproj.io/frontend.update-strategy: newest-build

    argocd-image-updater.argoproj.io/write-back-method: git
    argocd-image-updater.argoproj.io/backend.helm.image-tag: workloads.backend.image.tag
    argocd-image-updater.argoproj.io/frontend.helm.image-tag: workloads.frontend.image.tag
```

### Option B: update by digest (works well with mutable tags)

If you prefer keeping `tag: main` and only tracking the digest, use `digest` strategy (requires chart supporting digest or ArgoCD handling it). Many Helm charts model this as `image.digest`.

```yaml
metadata:
  annotations:
    argocd-image-updater.argoproj.io/image-list: |
      backend=ghcr.io/<OWNER>/leapmailr:main
      frontend=ghcr.io/<OWNER>/leapmailr-ui:main

    argocd-image-updater.argoproj.io/backend.update-strategy: digest
    argocd-image-updater.argoproj.io/frontend.update-strategy: digest

    argocd-image-updater.argoproj.io/write-back-method: git

    # Update these to match your chart’s digest fields.
    argocd-image-updater.argoproj.io/backend.helm.image-digest: backend.image.digest
    argocd-image-updater.argoproj.io/frontend.helm.image-digest: frontend.image.digest
```

## Registry credentials (GHCR)

If your GHCR images are private, Image Updater needs credentials.

Typical approaches:
- Create a robot PAT with `read:packages` and store as a Kubernetes secret referenced by Image Updater.
- Or make the packages public.

## What CI does (and does not do)

- CI builds/tests and pushes images to GHCR.
- CI does **not** deploy.
- ArgoCD syncs changes; Image Updater is what drives tag bumps.
