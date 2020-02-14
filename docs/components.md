### Custom Components
This page will document some of the custom components of transportd and how they can be used by consumers.

##### Validate Headers

The `validate_headers` component can be used to validate the presence of headers and the existence of specific values in those headers.
It verifies that at least one specified header contains at least one specified value. Each key under `allowed`
is the header name and the list of values for that key represent the the values we want to verify exist.

If a configured header does not exist, or the specified value is not found, it will check for the next `allowed` header, if configured,
and if the configured value exists. If it can't find any configured header or configured value for that header, it will reject the request.
This makes the plugin flexible to check for the presence of any header values combination. It does not currently do strict matching
if you want to validate that *multiple* values exist.

Example: This configuration would verify that a header named `Ldap-Groups` exists and that it has any of the listed values `hr` or `pm`.
```yaml
validateheaders:
  allowed:
    ldap-groups:
      - "hr"
      - "pm"
```

`validate_headers` also supports splitting on a configurable delimiter for each `allowed` header. To accomplish this, we set a `split` configuration using the headers name and the character(s) we want to `split` on.

Example: This configuration would verify that a header named `Ldap-Groups` or `Username` exists with any of the listed values after being split with the configured delimiter. For instance, incoming headers with values set like:
```yaml
Ldap-Groups: "hr&pm&sre"
Username: "jsmith,jakesmith,jakes"
```
Would need the following configuration in order to be split and validated correctly.
```yaml
validateheaders:
  allowed:
    ldap-groups:
      - "hr"
      - "pm"
    username:
      - "cloud-admin"
      - "jakesmith"
    split:
      ldap-groups: "&"
      username: ","
```

