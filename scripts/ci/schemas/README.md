# CRD manifest validation

Kubeconform's default catalog omits the `CustomResourceDefinition` root schema.
The adjacent JSON file adds it from Kubernetes' apiextensions OpenAPI definitions.
Its provenance records the source commit, Git blob ID and source/generated
SHA-256 hashes. The source is Apache-2.0 licensed.

To regenerate, fetch the pinned `pkg/apiserver/testdata/swagger.json` from
`kubernetes/apiextensions-apiserver`. Verify its blob ID and SHA-256 against the
provenance before processing it. Retain all `definitions`, remove their
`description` keys recursively, then write a JSON document with:

```json
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "$ref": "#/definitions/io.k8s.apiextensions-apiserver.pkg.apis.apiextensions.v1.CustomResourceDefinition",
  "definitions": {}
}
```

Replace the empty `definitions` with the processed source definitions. Serialize
with Python's `json.dumps(document, indent=2)` and append one newline. Verify
the generated hash. Kubernetes release changes need a reviewed provenance update.

This checks the CRD manifest's types and structure. Actual API-server admission
checks its embedded schema, CEL rules and supported Kubernetes versions in the
collector's real-cluster CI.
