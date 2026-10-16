# Test-only MaaS backend setup

This repository's MaaS usage-log coverage needs a reachable PostgreSQL database
for MaaS API state. The repository does not install or manage that database in
production. For short-lived validation on an OpenShift AI test cluster, the
optional [`docs/examples/maas-ephemeral-postgresql.yaml`](examples/maas-ephemeral-postgresql.yaml)
provides a single-replica PostgreSQL deployment, PVC, Service, and the
`maas-db-config` Secret expected by MaaS.

## Apply the test backend

The AI Gateway operator must have created `redhat-ai-gateway-infra` first. If
MaaS is already enabled, apply the manifest and restart the API after the
Secret exists so it reloads the connection URL:

```bash
oc apply -f docs/examples/maas-ephemeral-postgresql.yaml
oc -n redhat-ai-gateway-infra wait --for=condition=Available \
  deployment/maas-postgres --timeout=2m
oc -n redhat-ai-gateway-infra rollout restart deployment/maas-api
oc -n redhat-ai-gateway-infra rollout status deployment/maas-api --timeout=3m
```

The manifest uses `gp3-csi` and a 5 GiB PVC. Change `storageClassName` and the
requested size when validating on a cluster with different storage classes.

## Why `app=postgres` is intentional

The MaaS controller generates `maas-api-egress-restrict`. Its default port
5432 rule selects PostgreSQL pods using the `app=postgres` label (and, on some
versions, the corresponding label in the legacy application namespace). The
temporary Deployment therefore sets `app: postgres` on both its Deployment and
pod template. Without that label, PostgreSQL can be healthy and the connection
URL can be correct while `maas-api` still cannot reach the database and
remains unready.

This fixture also uses a stable in-cluster DNS name and `sslmode=disable` for
test-only use. Do not copy its credentials, image tag, or storage design into
a production environment. Use a supported PostgreSQL service with TLS,
durable backups, secret management, and an appropriate network-policy rule in
production.

## Cleanup

Remove the fixture after validation. This deletes the PVC and the test data:

```bash
oc delete -f docs/examples/maas-ephemeral-postgresql.yaml
```

If MaaS remains enabled, remove or replace `maas-db-config` only after the
MaaS API has been disabled or pointed at a production database.

## Validate the usage dashboards

The test backend only provides MaaS API state. Usage logs still require the
Monitoring resource to configure `spec.usageLogs.storage` with a separate S3
secret and bucket. The observability operator creates the dedicated
`LokiStack/usage`; MaaS then creates its collector, tenancy proxy, Perses
datasources, and dashboards.

Check the complete dependency path in order:

```bash
oc -n redhat-ods-monitoring get lokistack usage
oc -n redhat-ods-monitoring get svc usage-gateway-http
oc -n redhat-ods-monitoring get cm usage-gateway-ca-bundle
oc -n redhat-ods-monitoring get persesdatasource usage-logs-all usage-logs-multi-tenancy
oc -n redhat-ods-monitoring get persesdashboard dashboard-4-maas-usage-logs-admin dashboard-5-maas-usage-logs
```

Both datasources should report `Available=True`. The admin datasource should
target `usage-gateway-http` and reference `usage-gateway-ca-bundle` with
`service-ca.crt`; the tenant datasource should target the MaaS tenancy proxy.
If the gateway or CA resources are missing, fix the `LokiStack/usage` setup
first. Do not disable TLS verification as a workaround.
