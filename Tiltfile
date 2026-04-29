# go-mall - Tiltfile (Kubernetes mode)
# Targets Docker Desktop or OrbStack k8s.
# Run: tilt up
# Prereqs: kubectl context pointing at docker-desktop or orbstack, Helm 3, Tilt

allow_k8s_contexts(['docker-desktop', 'orbstack'])

update_settings(max_parallel_updates=3)

# Watch ignores live in .tiltignore; only service-local paths needed here.

load('ext://helm_resource', 'helm_resource', 'helm_repo')
load('ext://restart_process', 'docker_build_with_restart')

# --- Dapr control plane (installs CRDs + sidecar injector) ---
helm_repo('dapr-repo', 'https://dapr.github.io/helm-charts/', labels=['infra'])
local_resource(
    'dapr-crds-pre',
    cmd='set -euo pipefail; helm show crds dapr-repo/dapr --version=1.17.6 | kubectl apply --server-side --force-conflicts -f -',
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

# Treat everything non-x86_64 as arm64 (valid for docker-desktop/orbstack on Apple Silicon).
arch = str(local('uname -m', quiet=True)).strip()
goarch = 'amd64' if arch == 'x86_64' else 'arm64'

def go_service(name, debug):
    build_flags = '-gcflags="all=-N -l"' if debug else ''
    compile_cmd = 'cd src/{svc} && mkdir -p bin && CGO_ENABLED=0 GOOS=linux GOARCH={arch} go build {flags} -o bin/app .'.format(
        svc=name, arch=goarch, flags=build_flags)

    local_resource(
        '%s-compile' % name,
        cmd=compile_cmd,
        deps=['src/%s' % name, 'src/shared'],
        # bin/ and etc/ excluded here; test/testdata covered by .tiltignore.
        ignore=['src/%s/bin' % name, 'src/%s/etc' % name],
        allow_parallel=True,
        labels=['compile'],
    )
    docker_build_with_restart(
        'go-mall/%s' % name,
        'src',
        dockerfile='src/Dockerfile.dev',
        build_args={'SERVICE': name, 'DEBUG': '1' if debug else '0'},
        only=['%s/bin/app' % name, 'entrypoint.sh', 'Dockerfile.dev'],
        # docker_build_with_restart injects its own wrapper; restate the real entrypoint.
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

# In debug mode: remove liveness probe (frozen process at breakpoint = desired state,
# not a failure). Readiness stays TCP so Service routing still works.
def debug_probes(name, port=8080, deps=None):
    patch = encode_json([
        {'op': 'remove',  'path': '/spec/template/spec/containers/0/livenessProbe'},
        {'op': 'replace', 'path': '/spec/template/spec/containers/0/readinessProbe',
         'value': {'tcpSocket': {'port': port}, 'periodSeconds': 5, 'failureThreshold': 3}},
    ])
    local_resource(
        '%s-debug-probes' % name,
        cmd="kubectl patch deployment %s -n go-mall --type=json -p '%s'" % (name, patch),
        deps=deps or ['deploy/k8s/services/%s-api/%s.yaml' % (name, name)],
        resource_deps=[name],
        labels=['debug'],
    )

# Single source of truth for Go services: name, app port, debug port, API path.
GO_SERVICES = [
    {'name': 'catalog', 'port': 9001, 'debug_port': 40001, 'path': '/api/v1/products'},
    {'name': 'cart',    'port': 9002, 'debug_port': 40002, 'path': '/api/v1/carts'},
    {'name': 'payment', 'port': 9003, 'debug_port': 40003, 'path': '/api/v1/payments'},
]

_sf_debug = 'storefront' in debug_services
docker_build(
    'go-mall/storefront',
    'src/storefront',
    dockerfile='src/storefront/Dockerfile.dev',
    build_args={'DEBUG': '1' if _sf_debug else '0'},
    live_update=[
        fall_back_on([
            'src/storefront/package.json',
            'src/storefront/package-lock.json',
            'src/storefront/next.config.ts',
            'src/storefront/tsconfig.json',
            'src/storefront/postcss.config.mjs',
            'src/storefront/Dockerfile.dev',
            'src/storefront/entrypoint.sh',
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

k8s_resource('redis',      port_forwards=['6379:6379'], labels=['infra'])
k8s_resource('keycloak',   port_forwards=['8080:8080'], links=[link('http://localhost:8080/admin', 'keycloak admin')], labels=['infra'])
k8s_resource('jaeger',     port_forwards=['16686:16686'], links=[link('http://localhost:16686', 'jaeger UI')], labels=['infra'])
k8s_resource('postgres',   port_forwards=['5432:5432'], labels=['infra'])

for s in GO_SERVICES:
    name = s['name']
    debug = name in debug_services
    go_service(name, debug)
    if debug:
        debug_probes(name)
    pf = ['%d:8080' % s['port']]
    if debug:
        pf.append('%d:40000' % s['debug_port'])
    k8s_resource(
        name,
        port_forwards=pf,
        resource_deps=['%s-compile' % name] + _INFRA_DEPS,
        links=[link('http://localhost:%d%s' % (s['port'], s['path']), name)],
        labels=['service'],
    )

if _sf_debug:
    debug_probes('storefront', port=3000, deps=['deploy/k8s/services/storefront.yaml'])
k8s_resource(
    'storefront',
    port_forwards=['3000:3000'] + (['9229:9229', '9230:9230'] if _sf_debug else []),
    resource_deps=['catalog', 'cart', 'payment', 'keycloak'],
    links=[link('http://localhost:3000', 'storefront')],
    labels=['service'],
)
