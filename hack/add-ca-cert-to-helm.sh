#!/bin/bash
set -euo pipefail

DEPLOYMENT_FILE="chart/templates/deployment.yaml"

if [ ! -f "$DEPLOYMENT_FILE" ]; then
    echo "Error: $DEPLOYMENT_FILE not found"
    exit 1
fi

TMP_FILE=$(mktemp)
trap 'rm -f "$TMP_FILE"' EXIT

awk '
/^        - name: KUBERNETES_CLUSTER_DOMAIN$/ {
    print
    getline
    print
    print "        {{- if .Values.instanceSecret.caCert }}"
    print "        - name: SSL_CERT_FILE"
    print "          value: /etc/dvls-ca-cert/ca.crt"
    print "        {{- end }}"
    next
}
/^        securityContext: {{- toYaml .Values.controllerManager.manager.containerSecurityContext$/ {
    print
    getline
    print
    print "        {{- if .Values.instanceSecret.caCert }}"
    print "        volumeMounts:"
    print "        - name: dvls-ca-cert"
    print "          mountPath: /etc/dvls-ca-cert"
    print "          readOnly: true"
    print "        {{- end }}"
    next
}
/^      terminationGracePeriodSeconds: 10$/ {
    print
    print "      {{- if .Values.instanceSecret.caCert }}"
    print "      volumes:"
    print "      - name: dvls-ca-cert"
    print "        secret:"
    print "          secretName: {{ include \"chart.fullname\" . }}-ca-cert"
    print "          items:"
    print "          - key: ca.crt"
    print "            path: ca.crt"
    print "      {{- end }}"
    next
}
{ print }
' "$DEPLOYMENT_FILE" > "$TMP_FILE"

mv "$TMP_FILE" "$DEPLOYMENT_FILE"
