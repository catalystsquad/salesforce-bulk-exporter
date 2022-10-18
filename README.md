# salesforce-bulk-exporter

The salesforce-bulk-exporter is a utility to assist in submitting bulk query
jobs to the Salesforce API and downloading the results. 


## Configuration

In order for the tool to work, you will need to supply your Salesforce URL, a
connected app client ID and client secret, and a username and password to
authenticate.

The configuration can be passed as environment variables, over the CLI, or via
a config file.

The config file default location is `$HOME/.salesforce-bulk-exporter.yaml`
structure looks like this:
```
base-url: "https://yoursalesforce.my.salesforce.com"
client-id: someclientid
client-secret: someclientsecret
username: yourusername@example.com
password: agoodpassword
```

## Commands

### export

The export command submits a job and optionally waits and downloads the job's
results after completion. By default it will query for all fields, excluding
compound fields.

example usage:
```
./salesforce-bulk-exporter export task -f task-export -w
```
example output:
```
Submitted bulk query job with ID: 00000ABC1234567890
job in progress, sleeping for 10s...
job in progress, sleeping for 10s...
job in progress, sleeping for 10s...
job in progress, sleeping for 10s...
job in progress, sleeping for 10s...
Saved export to files: task-export.0.csv
```

### download

The download command downloads the results of completed tasks.

example usage:
```
./salesforce-bulk-exporter download 00000ABC1234567890
```
example output:
```
Saved export to files: export.0.csv
```

### list-jobs

The list command lists previously executed tasks.

example usage:
```
./salesforce-bulk-exporter list-jobs
```
example output:
```
+--------------------+-------------+-------------+------------------------------+
| ID                 | OBJECT      | STATE       | SYSTEMMODSTAMP               |
+--------------------+-------------+-------------+------------------------------+
| 00000ABC1234567890 | Task        | JobComplete | 2022-09-28T20:14:29.000+0000 |
+--------------------+-------------+-------------+------------------------------+
```

