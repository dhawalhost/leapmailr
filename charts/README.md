# leapmailr Helm chart

A reusable Helm chart that deploys one or more workloads from a `workloads` map.

## Install

```bash
helm upgrade --install leapmailr ./charts \
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

This chart deploys the leapmailr backend workload.

## FQDN / Ingress hosts

Set hostnames per workload using `workloads.<name>.ingress.hosts`.

Example:

```yaml
workloads:
  backend:
    ingress:
      enabled: true
      hosts:
        - host: api.leapmailr.example.com
          paths:
            - path: /
              pathType: Prefix
              servicePort: 8080
```

In this repo, deploy-time configuration lives in `charts/values.yaml`.
