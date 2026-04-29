# go-mall - Tiltfile (Kubernetes mode)
# Targets Docker Desktop or OrbStack k8s.
# Run: tilt up
# Prereqs: kubectl context pointing at docker-desktop or orbstack, Helm 3, Tilt

allow_k8s_contexts(['docker-desktop', 'orbstack'])

load('ext://helm_resource', 'helm_resource', 'helm_repo')

# --- Dapr control plane (installs CRDs + sidecar injector) ---
helm_repo('dapr-repo', 'https://dapr.github.io/helm-charts/', labels=['infra'])
helm_resource(
    'dapr',
    'dapr-repo/dapr',
    namespace='dapr-system',
    flags=[
        '--create-namespace',
        '--set=global.ha.enabled=false',
    ],
    resource_deps=['dapr-repo'],
    labels=['infra'],
)

# --- Docker image builds ---
docker_build(
    'go-mall/catalog',
    'src',
    dockerfile='src/Dockerfile',
    build_args={'SERVICE': 'catalog'},
)

docker_build(
    'go-mall/cart',
    'src',
    dockerfile='src/Dockerfile',
    build_args={'SERVICE': 'cart'},
)

docker_build(
    'go-mall/payment',
    'src',
    dockerfile='src/Dockerfile',
    build_args={'SERVICE': 'payment'},
)

docker_build(
    'go-mall/storefront',
    'src/storefront',
    dockerfile='src/storefront/Dockerfile',
)

# --- Kubernetes manifests via Kustomize ---
# --load-restrictor=LoadRestrictionsNone allows keycloak realm-export.json
# to be referenced from outside deploy/k8s/base/
k8s_yaml(kustomize('deploy/k8s'))

# --- Dapr CRDs: applied after Dapr helm install registers the CRD types ---
local_resource(
    'dapr-crds',
    cmd='kubectl apply -n go-mall -f deploy/k8s/dapr/dapr-config.yaml -f deploy/k8s/dapr/dapr-components.yaml',
    deps=['deploy/k8s/dapr/dapr-config.yaml', 'deploy/k8s/dapr/dapr-components.yaml'],
    resource_deps=['dapr'],
    labels=['infra'],
)

# --- Resource configuration: port-forwards, deps, labels ---
k8s_resource('redis',      labels=['infra'])
k8s_resource('keycloak',   port_forwards=['8080:8080'], labels=['infra'])
k8s_resource('jaeger',     port_forwards=['16686:16686', '4317:4317', '4318:4318', '9411:9411'], labels=['infra'])

k8s_resource(
    'catalog',
    port_forwards=['9001:8080'],
    resource_deps=['dapr-crds', 'redis', 'keycloak', 'jaeger'],
    labels=['service'],
)
k8s_resource(
    'cart',
    port_forwards=['9002:8080'],
    resource_deps=['dapr-crds', 'redis', 'keycloak', 'jaeger'],
    labels=['service'],
)
k8s_resource(
    'payment',
    port_forwards=['9003:8080'],
    resource_deps=['dapr-crds', 'redis', 'keycloak', 'jaeger'],
    labels=['service'],
)
k8s_resource(
    'storefront',
    port_forwards=['3000:3000'],
    resource_deps=['catalog', 'cart', 'payment'],
    labels=['service'],
)
