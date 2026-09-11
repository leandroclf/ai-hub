#!/usr/bin/env python3
"""Gera as dependências do perfil kind independente do Compose.

Os volumes são efêmeros de laboratório e as imagens são as mesmas referências
fixadas no Compose local. O perfil não é uma topologia de produção: serve para
provar DNS, probes, identidade, mensageria, cofre, observabilidade e o caminho
HTTP dentro do cluster sem endpoints apontando para IPs de containers externos.
"""
import os
import yaml

NAMESPACE = os.environ.get("R2_KIND_NAMESPACE", "ai-hub-local-kind")


def metadata(name, labels=None):
    out = {"name": name, "namespace": NAMESPACE}
    if labels:
        out["labels"] = labels
    return out


def service(name, port, target_port=None, extra_ports=None):
    ports = [{"name": "http", "port": port, "targetPort": target_port or port}]
    if extra_ports:
        ports.extend(extra_ports)
    return {
        "apiVersion": "v1",
        "kind": "Service",
        "metadata": metadata(name, {"app.kubernetes.io/name": name, "app.kubernetes.io/part-of": "ai-hub"}),
        "spec": {"selector": {"app.kubernetes.io/name": name}, "ports": ports},
    }


def deployment(name, image, container_port, env=None, command=None, args=None, mounts=None, volumes=None, replicas=1):
    container = {"name": name, "image": image, "imagePullPolicy": "IfNotPresent", "ports": [{"name": "http", "containerPort": container_port}]}
    if env:
        container["env"] = [{"name": key, "value": str(value)} for key, value in env.items()]
    if command:
        container["command"] = command
    if args:
        container["args"] = args
    if mounts:
        container["volumeMounts"] = mounts
    pod = {"automountServiceAccountToken": False, "containers": [container]}
    if volumes:
        pod["volumes"] = volumes
    return {
        "apiVersion": "apps/v1",
        "kind": "Deployment",
        "metadata": metadata(name, {"app.kubernetes.io/name": name, "app.kubernetes.io/part-of": "ai-hub"}),
        "spec": {"replicas": replicas, "selector": {"matchLabels": {"app.kubernetes.io/name": name}}, "template": {"metadata": {"labels": {"app.kubernetes.io/name": name, "app.kubernetes.io/part-of": "ai-hub"}}, "spec": pod}},
    }


def config_mount(config_name, key, path, mount_path):
    return {"name": config_name, "mountPath": mount_path, "subPath": key, "readOnly": True}


docs = []
docs.append({"apiVersion": "v1", "kind": "Secret", "metadata": metadata("postgres-credentials"), "stringData": {"POSTGRES_PASSWORD": "r2-local-fixture"}})
docs.append({"apiVersion": "v1", "kind": "ConfigMap", "metadata": metadata("postgres-init"), "data": {"01-databases.sh": "#!/bin/sh\nset -eu\nfor db in hub_control hub_core hub_finance hub_identity hub_control_kind hub_core_kind hub_finance_kind; do psql -v ON_ERROR_STOP=1 -U \"$POSTGRES_USER\" -d postgres -c \"CREATE DATABASE $db\"; done\n"}})
docs.append(service("postgres", 5432))
docs.append(deployment("postgres", "ai-hub-kind-postgres:16", 5432, {"POSTGRES_USER": "hub", "POSTGRES_PASSWORD": "r2-local-fixture", "POSTGRES_DB": "postgres"}, command=["docker-entrypoint.sh"], args=["postgres", "-c", "max_connections=200"], mounts=[config_mount("postgres-init", "01-databases.sh", "/docker-entrypoint-initdb.d/01-databases.sh", "/docker-entrypoint-initdb.d/01-databases.sh")], volumes=[{"name": "postgres-data", "emptyDir": {}}, {"name": "postgres-init", "configMap": {"name": "postgres-init"}}]))

docs.append(service("localstack", 4566))
docs.append(deployment("localstack", "ai-hub-kind-localstack:3.8", 4566, {"SERVICES": "sqs,sns,s3,secretsmanager", "DEFAULT_REGION": "us-east-1", "PERSISTENCE": "1", "SNAPSHOT_SAVE_STRATEGY": "ON_SHUTDOWN"}, volumes=[{"name": "localstack-data", "emptyDir": {}}], mounts=[{"name": "localstack-data", "mountPath": "/var/lib/localstack"}]))

for name, image, port, extra in [
    ("provider-sim", "ai-hub-r2-provider-sim:r2", 8090, {"HTTP_ADDR": ":8090", "ENVIRONMENT": "local", "CELL_ID": "r2-cell-a", "CALLBACK_INGRESS_KEY": "r4-callback-ingress-fixture", "PROVIDER_STATE_FILE": "/var/lib/provider-sim/state.json"}),
    ("webhook-sink", "ai-hub-r2-webhook-sink:r2", 8091, {"HTTP_ADDR": ":8091", "ENVIRONMENT": "local", "CELL_ID": "r2-cell-a"}),
]:
    docs.append(service(name, port))
    mounts = [{"name": "data", "mountPath": "/var/lib/provider-sim"}] if name == "provider-sim" else []
    volumes = [{"name": "data", "emptyDir": {}}] if name == "provider-sim" else []
    docs.append(deployment(name, image, port, extra, mounts=mounts, volumes=volumes))

docs.append(service("identity", 8080))
docs.append(deployment("identity", "ai-hub-kind-keycloak:26.7.3", 8080, {
    "KC_DB": "postgres", "KC_DB_URL": "jdbc:postgresql://postgres:5432/hub_identity", "KC_DB_USERNAME": "hub", "KC_DB_PASSWORD": "r2-local-fixture",
    "KC_HOSTNAME": "http://localhost:18085", "KC_HEALTH_ENABLED": "true", "KC_BOOTSTRAP_ADMIN_USERNAME": "r2-bootstrap", "KC_BOOTSTRAP_ADMIN_PASSWORD": "r2-bootstrap-fixture",
}, command=["/opt/keycloak/bin/kc.sh"], args=["start-dev", "--import-realm"], mounts=[{"name": "realm", "mountPath": "/opt/keycloak/data/import", "readOnly": True}], volumes=[{"name": "realm", "configMap": {"name": "identity-realm"}}]))

docs.append(service("loki", 3100))
docs.append(deployment("loki", "ai-hub-kind-loki:3.3.2", 3100, command=["/usr/bin/loki"], args=["-config.file=/etc/loki/config.yaml"], mounts=[config_mount("loki-config", "config.yaml", "/etc/loki/config.yaml", "/etc/loki/config.yaml")], volumes=[{"name": "loki-config", "configMap": {"name": "loki-config"}}, {"name": "loki-data", "emptyDir": {}}]))
docs.append(service("tempo", 3200, extra_ports=[{"name": "otlp-grpc", "port": 4317, "targetPort": 4317}, {"name": "otlp-http", "port": 4318, "targetPort": 4318}]))
docs.append(deployment("tempo", "ai-hub-kind-tempo:2.6.1", 3200, command=["/tempo"], args=["-config.file=/etc/tempo.yaml"], mounts=[config_mount("tempo-config", "tempo.yaml", "/etc/tempo.yaml", "/etc/tempo.yaml")], volumes=[{"name": "tempo-config", "configMap": {"name": "tempo-config"}}, {"name": "tempo-data", "emptyDir": {}}]))
docs.append(service("alloy", 4318, extra_ports=[{"name": "otlp-grpc", "port": 4317, "targetPort": 4317}, {"name": "admin", "port": 12345, "targetPort": 12345}]))
docs.append(deployment("alloy", "ai-hub-kind-alloy:v1.5.1", 4318, command=["/bin/alloy"], args=["run", "--server.http.listen-addr=0.0.0.0:12345", "/etc/alloy/config.alloy"], mounts=[config_mount("alloy-config", "config.alloy", "/etc/alloy/config.alloy", "/etc/alloy/config.alloy")], volumes=[{"name": "alloy-config", "configMap": {"name": "alloy-config"}}]))
docs.append(service("prometheus", 9090))
docs.append(deployment("prometheus", "ai-hub-kind-prometheus:v3.1.0", 9090, command=["/bin/prometheus"], args=["--config.file=/etc/prometheus/prometheus.yml", "--storage.tsdb.retention.time=3d"], mounts=[config_mount("prometheus-config", "prometheus.yml", "/etc/prometheus/prometheus.yml", "/etc/prometheus/prometheus.yml"), config_mount("prometheus-alerts", "alerts.yml", "/etc/prometheus/alerts.yml", "/etc/prometheus/alerts.yml")], volumes=[{"name": "prometheus-config", "configMap": {"name": "prometheus-config"}}, {"name": "prometheus-alerts", "configMap": {"name": "prometheus-alerts"}}, {"name": "prometheus-data", "emptyDir": {}}]))
docs.append(service("grafana", 3000))
docs.append(deployment("grafana", "ai-hub-kind-grafana:11.4.0", 3000, {"GF_SECURITY_ADMIN_USER": "r2-observer", "GF_SECURITY_ADMIN_PASSWORD": "r2-observer-fixture", "GF_AUTH_ANONYMOUS_ENABLED": "false", "GF_USERS_ALLOW_SIGN_UP": "false"}, volumes=[{"name": "grafana-data", "emptyDir": {}}], mounts=[{"name": "grafana-data", "mountPath": "/var/lib/grafana"}]))

docs.append({
    "apiVersion": "batch/v1",
    "kind": "Job",
    "metadata": metadata("hub-migrate", {"app.kubernetes.io/part-of": "ai-hub"}),
    "spec": {
        "backoffLimit": 3,
        "template": {
            "metadata": {"labels": {"app.kubernetes.io/name": "hub-migrate"}},
            "spec": {
                "restartPolicy": "OnFailure",
                "containers": [{
                    "name": "migrate",
                    "image": "ai-hub-kind-postgres:16",
                    "env": [{"name": "PGHOST", "value": "postgres"}, {"name": "PGUSER", "value": "hub"}, {"name": "PGPASSWORD", "value": "r2-local-fixture"}, {"name": "MIGRATION_DB_SUFFIX", "value": "_kind"}],
                    "command": ["/bin/sh", "/scripts/migrate.sh"],
                    "volumeMounts": [
                        {"name": "migrate-script", "mountPath": "/scripts/migrate.sh", "subPath": "migrate.sh", "readOnly": True},
                        {"name": "control-migrations", "mountPath": "/migrations/control", "readOnly": True},
                        {"name": "core-migrations", "mountPath": "/migrations/core", "readOnly": True},
                        {"name": "finance-migrations", "mountPath": "/migrations/finance", "readOnly": True},
                    ],
                }],
                "volumes": [
                    {"name": "migrate-script", "configMap": {"name": "migrate-script"}},
                    {"name": "control-migrations", "configMap": {"name": "control-migrations"}},
                    {"name": "core-migrations", "configMap": {"name": "core-migrations"}},
                    {"name": "finance-migrations", "configMap": {"name": "finance-migrations"}},
                ],
            },
        },
    },
})

docs.append({
    "apiVersion": "batch/v1",
    "kind": "Job",
    "metadata": metadata("identity-reconcile", {"app.kubernetes.io/part-of": "ai-hub"}),
    "spec": {
        "backoffLimit": 5,
        "template": {
            "metadata": {"labels": {"app.kubernetes.io/name": "identity-reconcile"}},
            "spec": {
                "restartPolicy": "OnFailure",
                "containers": [{
                    "name": "reconcile",
                    "image": "ai-hub-kind-python:3.12-alpine",
                    "command": ["python", "/scripts/reconcile_identity.py"],
                    "env": [
                        {"name": "KEYCLOAK_URL", "value": "http://identity:8080"},
                        {"name": "KEYCLOAK_REALM", "value": "ai-hub-r2"},
                        {"name": "KEYCLOAK_ADMIN_USER", "value": "r2-bootstrap"},
                        {"name": "KEYCLOAK_ADMIN_PASSWORD", "value": "r2-bootstrap-fixture"},
                        {"name": "R2_FIXTURE_PASSWORD", "value": "R2-fixture-password!"},
                        {"name": "R2_FIXTURE_OTP", "value": "IFES2SCVIIWVEMRNJVDECLKLIVMS2MBR"},
                    ],
                    "volumeMounts": [{"name": "reconcile-script", "mountPath": "/scripts/reconcile_identity.py", "subPath": "reconcile_identity.py", "readOnly": True}],
                }],
                "volumes": [{"name": "reconcile-script", "configMap": {"name": "identity-reconcile"}}],
            },
        },
    },
})

docs.append(service("kong", 8000))
docs.append(deployment("kong", "ai-hub-kind-kong:3.8", 8000, {"KONG_DATABASE": "off", "KONG_DECLARATIVE_CONFIG": "/kong/kong.yml", "KONG_ADMIN_LISTEN": "off", "KONG_NGINX_WORKER_PROCESSES": "1"}, mounts=[config_mount("kong-config", "kong.yml", "/kong/kong.yml", "/kong/kong.yml")], volumes=[{"name": "kong-config", "configMap": {"name": "kong-config"}}]))
docs.append(service("admin-ui", 8080))
docs.append(deployment("admin-ui", "ai-hub-r2-admin-ui:r2", 8080))

print(yaml.safe_dump_all(docs, sort_keys=False))
