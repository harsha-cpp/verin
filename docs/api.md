# API Notes

## Contract source

- OpenAPI spec: [openapi.yaml](/Users/sriharsha/Developer/projects/verin/docs/openapi.yaml)
- Generated client wrapper: [index.ts](/Users/sriharsha/Developer/projects/verin/packages/api-client/src/index.ts)

## Main endpoint groups

- `auth`
  - login, logout, current session, MFA setup, MFA verify
- `documents`
  - list, init upload, complete upload, detail, update, archive, restore, versions, signed download, comments
- `search`
  - keyword search, advanced search, saved searches
- `audit`
  - event listing, export queueing
- `notifications`
  - list, mark read
- `admin`
  - users, roles, quotas, retention, settings, jobs, usage, health

## Error envelope

All error responses use:

```json
{
  "error": {
    "code": "SOME_CODE",
    "message": "Human-readable message",
    "details": {}
  },
  "requestId": "req_123"
}
```

## Security expectations

- Cookie-authenticated routes require CSRF protection for state-changing requests.
- Signed download URLs are issued only after a fresh permission check.
- Sensitive document and admin actions are recorded in `audit_events`.
