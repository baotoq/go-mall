load('ext://restart_process', 'docker_build_with_restart')
load('ext://helm_resource', 'helm_resource', 'helm_repo')
allow_k8s_contexts(['docker-desktop', 'orbstack'])
docker_prune_settings(num_builds=1, keep_recent=1)

# Usage:
#   tilt up                    Delve waits for debugger to attach
#   tilt up -- --continue      Delve starts immediately (no wait)
config.define_bool('continue', args=False, usage='Start Delve with --continue')
dlv_continue = config.parse().get('continue', False)

dlv_flags = '--headless --listen=:7000 --accept-multiclient --only-same-user=false --log'
if dlv_continue:
    dlv_flags += ' --continue'

# (name, http_host_port, grpc_host_port, dlv_host_port, extra_k8s_deps)
SERVICES = [
    ('catalog', 8001, 9001, 7001, ['postgres']),
    ('cart',    8002, 9002, 7002, ['postgres']),
    ('payment', 8003, 9003, 7003, ['postgres']),
    ('order',   8004, 9004, 7004, ['postgres']),
]

def compile_cmd(name):
    return (
        'mkdir -p dist && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 ' +
        'go build -gcflags="all=-N -l" -ldflags "-X main.Version=dev" ' +
        '-o ./dist/{name} ./app/{name}/cmd/server'
    ).format(name=name)

def dlv_entrypoint(name):
    return ['sh', '-c', 'exec dlv exec /app/{name} {flags} -- -conf /data/conf'.format(
        name=name, flags=dlv_flags,
    )]

for svc in SERVICES:
    name, http_port, grpc_port, dlv_port, extra_k8s_deps = svc

    local_resource(
        'compile-' + name,
        cmd=compile_cmd(name),
        deps=['./app/' + name, './api/' + name, 'go.mod', 'go.sum'],
        labels=['build'],
    )

    docker_build_with_restart(
        name,
        '.',
        entrypoint=dlv_entrypoint(name),
        dockerfile='app/' + name + '/Dockerfile.dev.debug',
        only=['./dist/' + name],
        live_update=[sync('./dist/' + name, '/app/' + name)],
    )

    k8s_resource(
        name,
        port_forwards=[
            str(http_port) + ':8000',
            str(grpc_port) + ':9000',
            str(dlv_port) + ':7000',
        ],
        resource_deps=extra_k8s_deps + ['compile-' + name, 'dapr', 'dapr-components'],
        labels=['app'],
    )

# Fetch/update all Helm subchart tarballs (postgres, redis).
# helm() is evaluated at Tiltfile load time, so this must run synchronously.
local('helm dependency update deploy/helm', quiet=True)

helm_repo('dapr-repo', 'https://dapr.github.io/helm-charts/', labels=['infra'])

# Delete any pre-existing Dapr CRDs (e.g. from `dapr init`) so Helm can claim
# field ownership cleanly. No-ops when Helm already manages the release.
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

k8s_yaml(helm(
    'deploy/helm',
    name='deps',
    namespace='go-mall',
    values=['deploy/helm/values.yaml'],
))
k8s_yaml(kustomize('deploy/k8s/overlays/debug', flags=['--load-restrictor=LoadRestrictionsNone']))

k8s_resource('postgres', port_forwards=['5432:5432'], labels=['infra'])
k8s_resource('redis',    port_forwards=['6379:6379'], labels=['infra'])
k8s_resource('pgadmin',  port_forwards=['5050:80'],   labels=['infra'], resource_deps=['postgres'])
k8s_resource('keycloak', port_forwards=['8080:8080'], labels=['infra'])

k8s_resource(
    objects=[
        'pubsub:Component',
        'secretstore:Component',
    ],
    new_name='dapr-components',
    resource_deps=['dapr'],
    labels=['infra'],
)
