load('ext://restart_process', 'docker_build_with_restart')
load('ext://helm_resource', 'helm_resource', 'helm_repo')
allow_k8s_contexts(['docker-desktop', 'orbstack'])
docker_prune_settings(num_builds=1, keep_recent=1)

# Usage:
#   tilt up                    Delve waits for debugger to attach
#   tilt up -- --continue      Delve starts immediately (no wait)
config.define_bool('continue', args=False, usage='Start Delve with --continue')
dlv_continue = config.parse().get('continue', False)

dlv_flags = '--headless --listen=:2345 --accept-multiclient --only-same-user=false --log'
if dlv_continue:
    dlv_flags += ' --continue'

entrypoint_greeter = ['sh', '-c', 'exec dlv exec /app/greeter ' + dlv_flags + ' -- -conf /data/conf']
entrypoint_catalog = ['sh', '-c', 'exec dlv exec /app/catalog ' + dlv_flags + ' -- -conf /data/conf']
entrypoint_cart = ['sh', '-c', 'exec dlv exec /app/cart ' + dlv_flags + ' -- -conf /data/conf']
entrypoint_payment = ['sh', '-c', 'exec dlv exec /app/payment ' + dlv_flags + ' -- -conf /data/conf']

compile_greeter = 'mkdir -p dist && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -gcflags="all=-N -l" -ldflags "-X main.Version=dev" -o ./dist/greeter ./app/greeter/cmd/server'
compile_catalog = 'mkdir -p dist && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -gcflags="all=-N -l" -ldflags "-X main.Version=dev" -o ./dist/catalog ./app/catalog/cmd/server'
compile_cart = 'mkdir -p dist && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -gcflags="all=-N -l" -ldflags "-X main.Version=dev" -o ./dist/cart ./app/cart/cmd/server'
compile_payment = 'mkdir -p dist && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -gcflags="all=-N -l" -ldflags "-X main.Version=dev" -o ./dist/payment ./app/payment/cmd/server'

# Compile locally on every Go source change.
# Result is synced into the running container — no full image rebuild needed.
local_resource('compile-greeter',
    cmd=compile_greeter,
    deps=['./app/greeter', './api/greeter', 'go.mod', 'go.sum'],
    labels=['build'],
)

local_resource('compile-catalog',
    cmd=compile_catalog,
    deps=['./app/catalog', './api/catalog', 'go.mod', 'go.sum'],
    labels=['build'],
)

local_resource('compile-cart',
    cmd=compile_cart,
    deps=['./app/cart', './api/cart', 'go.mod', 'go.sum'],
    labels=['build'],
)

local_resource('compile-payment',
    cmd=compile_payment,
    deps=['./app/payment', './api/payment', 'go.mod', 'go.sum'],
    labels=['build'],
)

# Helm chart tarballs for postgres/redis. helm() below is evaluated at
# Tiltfile load time, so this fetch must run synchronously here — a
# local_resource would fire too late to affect the current load.
local('[ -d deploy/helm/charts ] || helm dependency update deploy/helm', quiet=True)

helm_repo('dapr-repo', 'https://dapr.github.io/helm-charts/', labels=['infra'])

# If Dapr is not yet installed via Helm, delete any pre-existing CRDs (e.g. from
# `dapr init`) so Helm can claim field ownership cleanly. No-ops when Helm already
# manages the release.
local_resource('patch-dapr-crds',
    cmd="""
    helm status dapr -n dapr-system 2>/dev/null | grep -q 'STATUS: deployed' || \
        kubectl get crds -o name 2>/dev/null | grep dapr.io | xargs kubectl delete --ignore-not-found
    """,
    resource_deps=['dapr-repo'],
    labels=['infra'],
)

helm_resource(
    'dapr',
    'dapr-repo/dapr',
    namespace='dapr-system',
    flags=[
        '--version=1.17.6',
        '--create-namespace',
        '--set=global.ha.enabled=false',
    ],
    resource_deps=['dapr-repo', 'patch-dapr-crds'],
    labels=['infra'],
)

# Tilt-optimised Dockerfile contains no Go toolchain — it just copies ./dist/greeter.
# only=['./dist'] means docker_build watches *only* that dir, so Go source
# changes never trigger an image rebuild; they go through compile → sync instead.
docker_build_with_restart(
    'greeter',
    '.',
    entrypoint=entrypoint_greeter,
    dockerfile='app/greeter/Dockerfile.dev.debug',
    only=['./dist'],
    live_update=[
        sync('./dist/greeter', '/app/greeter'),
    ],
)

docker_build_with_restart(
    'catalog',
    '.',
    entrypoint=entrypoint_catalog,
    dockerfile='app/catalog/Dockerfile.dev.debug',
    only=['./dist'],
    live_update=[
        sync('./dist/catalog', '/app/catalog'),
    ],
)

docker_build_with_restart(
    'cart',
    '.',
    entrypoint=entrypoint_cart,
    dockerfile='app/cart/Dockerfile.dev.debug',
    only=['./dist'],
    live_update=[
        sync('./dist/cart', '/app/cart'),
    ],
)

docker_build_with_restart(
    'payment',
    '.',
    entrypoint=entrypoint_payment,
    dockerfile='app/payment/Dockerfile.dev.debug',
    only=['./dist'],
    live_update=[
        sync('./dist/payment', '/app/payment'),
    ],
)

k8s_yaml(helm(
    'deploy/helm',
    name='deps',
    namespace='greeter',
    values=['deploy/helm/values.yaml'],
))
k8s_yaml(kustomize('deploy/k8s/overlays/debug', flags=['--load-restrictor=LoadRestrictionsNone']))

k8s_resource('postgres', port_forwards=['5432:5432'], labels=['infra'])
k8s_resource('redis',    port_forwards=['6379:6379'], labels=['infra'])
k8s_resource('pgadmin',  port_forwards=['5050:80'],   labels=['infra'], resource_deps=['postgres'])

# Keycloak
local_resource(
    'keycloak-namespace',
    cmd='kubectl apply -f deploy/k8s/base/infra/keycloak/namespace.yaml',
    labels=['infra'],
)

local_resource(
    'keycloak-secret',
    cmd='kubectl apply -f deploy/k8s/base/infra/keycloak/secret.yaml',
    deps=['deploy/k8s/base/infra/keycloak/secret.yaml'],
    resource_deps=['keycloak-namespace'],
    labels=['infra'],
)

helm_resource(
    'keycloak',
    'oci://registry-1.docker.io/bitnamicharts/keycloak',
    namespace='keycloak',
    flags=[
        '--version=24.0.6',
        '--set=auth.adminUser=admin',
        '--set=auth.adminPassword=admin',
        '--set=postgresql.enabled=true',
    ],
    resource_deps=['keycloak-namespace', 'keycloak-secret'],
    labels=['infra'],
    port_forwards=['8080:80'],
)

k8s_resource(
    objects=[
        'pubsub:Component:greeter',
        'secretstore:Component:greeter',
    ],
    new_name='dapr-components',
    resource_deps=['dapr'],
    labels=['infra'],
)

k8s_resource('greeter',
    port_forwards=['8000:8000', '9000:9000', '2345:2345'],
    resource_deps=['postgres', 'redis', 'compile-greeter', 'dapr', 'dapr-components'],
    labels=['app'],
)

k8s_resource('catalog',
    port_forwards=['8001:8000', '9001:9000', '2346:2345'],
    resource_deps=['postgres', 'compile-catalog', 'dapr', 'dapr-components'],
    labels=['app'],
)

k8s_resource('cart',
    port_forwards=['8002:8000', '9002:9000', '2347:2345'],
    resource_deps=['postgres', 'compile-cart', 'dapr', 'dapr-components'],
    labels=['app'],
)

k8s_resource('payment',
    port_forwards=['8003:8000', '9003:9000', '2348:2345'],
    resource_deps=['postgres', 'compile-payment', 'dapr', 'dapr-components'],
    labels=['app'],
)
