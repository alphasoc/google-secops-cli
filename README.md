This is the very beginning of CLI tool interacting with Google SecOps. The intetion is to provide the bits required for integration with AlphaSOC.

At the moment it only allows for managing DataTaps (https://cloud.google.com/chronicle/docs/reference/datatapconfig-api)
and allows to list, create, and delete DataTaps.

## Running

You can build the binary (`go build .`) and use `google-secops-datatap ...` binary or run it directly via `go run main.go ...`.

## Examples

**Show help:**

```
./google-secops-datatap -h
```

**Create DataTap** named "datatap-1", writing data to a specified Pub/Sub topic, in JSON format, using credentials from a file:

```
./google-secops-datatap -credentials="gcp-creds.json" create  -displayName="datatap-1" -topic="projects/my-lovely-project/topics/secops-datatap-1"
```
(please note that -credentials need to be provided before command name as the argument belongs to the root and not a sub-command)

**List DataTaps:**

```
./google-secops-datatap -credentials="gcp-creds.json" list
```

**Delete DataTap:**

```
./google-secops-datatap -credentials="gcp-creds.json" delete 2c91636d-acf2-105b-a6d0-281e1a5c368b
```
