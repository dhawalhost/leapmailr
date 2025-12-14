# generic-app Helm chart

A reusable Helm chart that deploys one or more workloads from a `workloads` map.

## Install

```bash
helm upgrade --install leapmailr ./helm/generic-app \
  --namespace leapmailr --create-namespace
```

## Values model

- `workloads.<name>` defines one component.
- Each workload creates a `Deployment` and (optionally) a `Service` and `Ingress`.

Common keys per workload:
- `enabled`
- `replicas`
- `image.repository`, `image.tag`, `image.pullPolicy`
- `containerPort`
- `service.enabled`, `service.type`, `service.port`
- `ingress.enabled`, `ingress.className`, `ingress.annotations`, `ingress.hosts`, `ingress.tls`
- `env` (map) merged with `global.env`
- `envFrom` appended with `global.envFrom`
- `resources`, `livenessProbe`, `readinessProbe`

This chart is intentionally generic so you can copy it to future projects.
