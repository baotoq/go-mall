# go-mall - Tiltfile (Kubernetes mode)
# Targets Docker Desktop or OrbStack k8s.
# Run: tilt up
# Prereqs: kubectl context pointing at docker-desktop or orbstack, Helm 3, Tilt

allow_k8s_contexts(['docker-desktop', 'orbstack'])

update_settings(max_parallel_updates=6)

watch_settings(ignore=[
    '**/*_test.go',
    '**/testdata/**',
    'src/storefront/node_modules/**',
    'src/storefront/.next/**',
])

load('ext://helm_resource', 'helm_resource', 'helm_repo')
load('ext://restart_process', 'docker_build_with_restart')

# --- Dapr control plane (installs CRDs + sidecar injector) ---
helm_repo('dapr-repo', 'https://dapr.github.io/helm-charts/', labels=['infra'])
local_resource(
    'dapr-crds-pre',
    cmd='set -euo pipefail; helm show crds dapr-repo/dapr | kubectl apply --server-side --force-conflicts -f -',
    resource_deps=['dapr-repo'],
    labels=['infra'],
)
helm_resource(
    'dapr',
    'dapr-repo/dapr',
    namespace='dapr-system',
    flags=[
        '--create-namespace',
        '--set=global.ha.enabled=false',
        '--skip-crds',
    ],
    resource_deps=['dapr-repo', 'dapr-crds-pre'],
    labels=['infra'],
)

# --- Go service fast iteration ---
# Pass --debug=catalog,cart etc. to run a service under dlv.
# Example: tilt up -- --debug=catalog
config.define_string_list('debug', usage='Services to run under dlv (comma-separated)')
cfg = config.parse()
debug_services = cfg.get('debug', [])

arch = str(local('uname -m', quiet=True)).strip()
goarch = 'amd64' if arch == 'x86_64' else 'arm64'

def go_service(name):
    debug = name in debug_services
    # No -ldflags stripping in dev: irrelevant for sync, can hurt incremental cache
    build_flags = '-gcflags="all=-N -l"' if debug else ''
    compile_cmd = 'cd src/{svc} && mkdir -p bin && CGO_ENABLED=0 GOOS=linux GOARCH={arch} go build {flags} -o bin/app .'.format(
        svc=name, arch=goarch, flags=build_flags)

    # Compile once on first run so binary exists before docker_build fires.
    # Skip if binary already exists (avoids recompiling on every Tiltfile reload).
    local('[ -f src/%s/bin/app ] || (%s)' % (name, compile_cmd), quiet=True)

    local_resource(
        '%s-compile' % name,
        cmd=compile_cmd,
        deps=['src/%s' % name, 'src/shared'],
        ignore=['src/%s/bin' % name, '**/*_test.go', '**/testdata'],
        labels=['compile'],
    )
    docker_build_with_restart(
        'go-mall/%s' % name,
        'src',
        dockerfile='src/Dockerfile.dev',
        build_args={'SERVICE': name, 'DEBUG': '1' if debug else '0'},
        only=['%s/bin/app' % name, 'entrypoint.sh'],
        entrypoint=['/app/entrypoint.sh'],
        live_update=[
            fall_back_on([
                'src/%s/go.mod' % name,
                'src/%s/go.sum' % name,
                'src/go.work',
                'src/go.work.sum',
            ]),
            sync('src/%s/bin/app' % name, '/app/app'),
        ],
    )

go_service('catalog')
go_service('cart')
go_service('payment')

# In debug mode: remove liveness probe (frozen goroutine at breakpoint = desired state,
# not a failure). Readiness stays TCP so Service routing still works.
_PATCH_REMOVE_LIVENESS   = '{"op":"remove","path":"/spec/template/spec/containers/0/livenessProbe"}'
_PATCH_READINESS_TO_TCP  = '{"op":"replace","path":"/spec/template/spec/containers/0/readinessProbe","value":{"tcpSocket":{"port":8080},"periodSeconds":5,"failureThreshold":3}}'

for _svc in debug_services:
    local_resource(
        '%s-debug-probes' % _svc,
        cmd="kubectl patch deployment %s -n go-mall --type=json -p '[%s,%s]'" % (_svc, _PATCH_REMOVE_LIVENESS, _PATCH_READINESS_TO_TCP),
        deps=['deploy/k8s/services/%s-api/%s.yaml' % (_svc, _svc)],
        resource_deps=[_svc],
        labels=['debug'],
    )

docker_build(
    'go-mall/storefront',
    'src/storefront',
    dockerfile='src/storefront/Dockerfile.dev',
    live_update=[
        fall_back_on([
            'src/storefront/package.json',
            'src/storefront/package-lock.json',
            'src/storefront/next.config.ts',
            'src/storefront/tsconfig.json',
            'src/storefront/postcss.config.mjs',
            'src/storefront/Dockerfile.dev',
        ]),
        sync('src/storefront/src', '/app/src'),
        sync('src/storefront/public', '/app/public'),
        sync('src/storefront/components.json', '/app/components.json'),
    ],
)

# --- Kubernetes manifests via Kustomize ---
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
_INFRA_DEPS = ['dapr-crds', 'redis', 'keycloak', 'jaeger', 'postgres']

k8s_resource('redis',      labels=['infra'])
k8s_resource('keycloak',   port_forwards=['8080:8080'], labels=['infra'])
k8s_resource('jaeger',     port_forwards=['16686:16686'], labels=['infra'])
k8s_resource('postgres',   port_forwards=['5432:5432'], labels=['infra'])

k8s_resource(
    'catalog',
    port_forwards=['9001:8080'] + (['40001:40000'] if 'catalog' in debug_services else []),
    resource_deps=['catalog-compile'] + _INFRA_DEPS,
    links=[link('http://localhost:9001/api/v1/products', 'products')],
    labels=['service'],
)
k8s_resource(
    'cart',
    port_forwards=['9002:8080'] + (['40002:40000'] if 'cart' in debug_services else []),
    resource_deps=['cart-compile'] + _INFRA_DEPS,
    links=[link('http://localhost:9002/api/v1/carts', 'carts')],
    labels=['service'],
)
k8s_resource(
    'payment',
    port_forwards=['9003:8080'] + (['40003:40000'] if 'payment' in debug_services else []),
    resource_deps=['payment-compile'] + _INFRA_DEPS,
    links=[link('http://localhost:9003/api/v1/payments', 'payments')],
    labels=['service'],
)
k8s_resource(
    'storefront',
    port_forwards=['3000:3000'],
    resource_deps=['catalog', 'cart', 'payment'],
    links=[link('http://localhost:3000', 'storefront')],
    labels=['service'],
)
