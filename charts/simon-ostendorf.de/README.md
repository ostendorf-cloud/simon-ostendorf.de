# simon-ostendorf-de

Installs [simon-ostendorf.de](https://github.com/ostendorf-cloud/simon-ostendorf.de) on a
Kubernetes cluster.

## Prerequisites

- Kubernetes 1.27 or newer.
- Only if `httpRoute.enabled` is set: the Gateway API CRDs and a running `Gateway` with a
  listener that allows this release's namespace. See "External routing" below; this chart
  ships neither.

There is no secret-shaped configuration yet, and so no `Secret` prerequisite: see
"Configuration" below.

## Installing

```sh
helm install simon-ostendorf-de charts/simon-ostendorf.de
```

This chart is also published as an OCI artifact, at
`oci://ghcr.io/ostendorf-cloud/charts/simon-ostendorf-de`, versioned in lock-step with every
release:

```sh
helm install simon-ostendorf-de oci://ghcr.io/ostendorf-cloud/charts/simon-ostendorf-de --version <version>
```

or pull it without installing:

```sh
helm pull oci://ghcr.io/ostendorf-cloud/charts/simon-ostendorf-de --version <version>
```

## Configuration

Every setting is plain, non-secret configuration, read from `config.*`; see `values.yaml` for
the full, commented list and the root `README.md`'s own Configuration table for what each
environment variable does. This chart never offers a values-driven way to create a `Secret`:
there is nothing secret-shaped to put in one yet.

## Replicas

`replicaCount` has no upper bound: this application keeps no server-side session state, so
more than one replica, or two pods briefly coexisting during an ordinary rolling update, is
harmless. The Deployment rolls out with the default `RollingUpdate` strategy.

## Health, readiness and metrics

`/healthz`, `/readyz` and `/metrics` are served on a second, administrative listener
(`service.adminPort`/`config.adminAddr`, default `8081`), never on the public `service.port`:
`templates/deployment.yaml`'s `startupProbe`, `livenessProbe` and `readinessProbe` all point at
the `admin` container port, and `templates/ingress.yaml` only ever routes to `service.port`. A
`ServiceMonitor` or `PodMonitor` (not part of this chart) can scrape `service.adminPort`
directly, in-cluster.

## External routing

This chart reaches external traffic to its own `Service` two ways, chosen by
`ingress.enabled` and `httpRoute.enabled`. Both default to `false`: a default install creates
neither, reachable only from inside the cluster, until one is deliberately turned on. They are
mutually exclusive: turning both on fails the install or upgrade with a readable message
naming which one to turn off, enforced once in `templates/deployment.yaml`, the one template
that always renders regardless of which routing flag is set.

### Ingress

Set `ingress.enabled: true`. `ingress.hosts` names the host and path rules, and `ingress.tls`
carries the certificate configuration directly on the `Ingress` object: this chart owns that
configuration.

### Gateway API HTTPRoute

Set `httpRoute.enabled: true`, `httpRoute.parentRefs.name` and `httpRoute.parentRefs.namespace`
to the `Gateway` this release attaches to. The namespace is required, not optional the way
`ingress.className` is: a `Gateway` is almost always platform owned and lives in a different
namespace than this release, and the Gateway API's own default for an unset
`parentRefs.namespace` is the `HTTPRoute`'s own namespace, normally wrong here.

This chart never creates the `Gateway` and never ships the Gateway API CRDs (`Gateway`,
`HTTPRoute` and the rest): both are cluster-level prerequisites, installed and operated once
per cluster by the cluster operator, outside this chart. The `Gateway`'s own `allowedRoutes`
decides whether it actually accepts this `HTTPRoute`; a `httpRoute.enabled: true` release that
the `Gateway` does not allow installs without error but never actually receives traffic.

Unlike `Ingress`, an `HTTPRoute` carries no `tls` field of its own. TLS terminates on the
`Gateway`'s listener instead, which the cluster operator, not this chart, owns and configures;
`httpRoute.hostnames` only has to match a hostname the listener itself already serves for
traffic to reach this `HTTPRoute` at all.

Routing itself is fixed to this chart's own `Service`, on `service.port`, never
`service.adminPort`: no values-driven way to add a second backend or a per-path rule, the same
single-backend assumption `ingress.hosts` already makes. An `HTTPRoute` rule with no `matches`
defaults to a `/` prefix match per the Gateway API spec, the same default an `Ingress` with one
`/` path rule gives, so this chart's `HTTPRoute` sets none explicitly.

## Values

See `values.yaml` for the full, commented list. The ones most deployments touch:

| Key | Description | Default |
| --- | --- | --- |
| `image.repository` | Container image | `ghcr.io/ostendorf-cloud/simon-ostendorf.de` |
| `image.tag` | Image tag | `.Chart.AppVersion` |
| `replicaCount` | Pod replicas | `1` |
| `service.port` | Public Service and container port, the only one the `Ingress` ever routes to | `8080` |
| `service.adminPort` | govue's admin listener: health, readiness and metrics, in-cluster only | `8081` |
| `ingress.enabled` | Create an `Ingress`. Mutually exclusive with `httpRoute.enabled`; see "External routing" above. | `false` |
| `ingress.annotations` | Annotations for the `Ingress` | `{}` |
| `ingress.tls` | TLS configuration for the `Ingress` | `[]` |
| `httpRoute.enabled` | Create a Gateway API `HTTPRoute`. Mutually exclusive with `ingress.enabled`; see "External routing" above. | `false` |
| `httpRoute.hostnames` | Hostnames the `HTTPRoute` matches | `[]` |
| `httpRoute.annotations` | Annotations for the `HTTPRoute` | `{}` |
| `httpRoute.parentRefs.name` | Name of the `Gateway` to attach to. Required when `httpRoute.enabled` is true | `""` |
| `httpRoute.parentRefs.namespace` | Namespace the `Gateway` lives in. Required when `httpRoute.enabled` is true; see "External routing" above for why this has no same-namespace fallback | `""` |
| `httpRoute.parentRefs.sectionName` | Specific listener on the `Gateway` to attach to, when it exposes more than one | `""` |
| `resources` | Pod resource requests/limits | see `values.yaml` |
| `config.addr` | Address the public HTTP server listens on | `:8080` |
| `config.adminAddr` | govue's admin listener address (`ADMIN_ADDR`); must match `service.adminPort` | `:8081` |
| `config.logLevel` | `debug`, `info`, `warn` or `error` | `info` |

## Upgrading

No breaking change has shipped yet; `Chart.yaml`'s `version` tracks that going forward.

## Uninstalling

```sh
helm uninstall simon-ostendorf-de
```

Removes the `Deployment`, `Service`, `ServiceAccount`, and the `Ingress` or `HTTPRoute`,
whichever was enabled, all created by this same chart.

## Verification

This chart is verified with:

- `helm lint`, over the default values and the documented ingress-enabled and httpRoute-enabled
  combinations. `helm lint` never propagates a template's `fail` call; it logs it at INFO and
  still exits zero, so it covers a values mistake schema validation would catch, not this
  chart's own render-time guard.
- `helm template`, for the same combinations as `helm lint` above, and for the one guard in
  `templates/deployment.yaml` and `templates/httproute.yaml` each, asserted to fail with that
  guard's own message rather than merely to render.
- `kubeconform`, against the rendered manifests of the must-succeed combinations.
- [`helm-unittest`](https://github.com/helm-unittest/helm-unittest), run with
  `helm unittest charts/simon-ostendorf.de`; see `mise.toml` for how the plugin is pinned. Its
  `failedTemplate` suites cover the same render-time guard as the must-fail set above,
  independently: both have to agree before a guard counts as checked.
