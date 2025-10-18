# search

Request URL: https://us2.app.sysdig.com/api/cspm/v1/compliance/violations/acceptances/search

Request Method: POST

Payload: {filter: "", pageNumber: 2, pageSize: 20, sort: "acceptanceDate", orderBy: "desc"}
- filter : ""
- orderBy : "desc"
- pageNumber : 2
- pageSize : 20
- sort : "acceptanceDate"

```
{
    "data": [
        {
            "tenantId": "1005999",
            "expiresAt": "0",
            "acceptanceDate": "1734758151762",
            "reason": "Risk Owned",
            "username": "xxxx@example.com",
            "userDisplayName": "xxxx@example.com",
            "type": 8,
            "sourceId": "",
            "filter": "name in (\"eu-west-1\")",
            "id": "abcedefg",
            "description": "foo-bar",
            "acceptPeriod": "Never",
            "isExpired": false,
            "isSystem": false,
            "zoneId": "0",
            "controlId": "16022"
        },
        ...
    ],
    "totalCount": 600
}
```

# delete
Request URL: https://us2.app.sysdig.com/api/cspm/v1/compliance/violations/revoke
Request Method: POST
Payload: {id: "6763aab48ebb8c821a3ddf89"}
