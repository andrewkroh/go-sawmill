# go-sawmill

> **Note**
> This project is a proof-of-concept and is incomplete.

go-sawmill is an event processing library and command-line application for
processing text and structured data. The processing steps are defined in YAML as
a list of discrete processors that manipulate the data.

## Installation

```
# Go 1.17+
go install github.com/andrewkroh/go-sawmill/cmd/sawmill@latest
```

## Usage

```
cat some-app.log | sawmill -p my-pipeline.yml > output.ndjson
```

## Pipeline Definition

The pipeline is constructed as YAML. An `id` and one or more `processors` are
required. The `description` and `on_failure` handler are optional. `on_failure`
is executed when any processor fails.

```yaml
---
id: my-pipeline-identifier
description: >
  Summarize the purpose of the pipeline.
processors:
  - set:
      field: label.app
      value: some-app
on_failure:
  - set:
      target_field: event.kind
      value: pipeline_error
      ignore_failure: true
```

## Processors

- [append](#append)
- [community_id](#community_id)
- [lowercase](#lowercase)
- [remove](#remove)
- [set](#set)
- [uppercase](#uppercase)
- [webassembly](#webassembly)


### append

Appends one or more values to an existing array if the field already exists and it is an array. Converts a scalar to an array and appends one or more values to it if the field exists and it is a scalar. Creates an array containing the provided values if the field doesn’t exist. Accepts a single value or an array of values.

| Option | Required | Optional | Type | Default | Description |
|--------|----------|----------|------|---------|-------------|
| allow_duplicates |  | x | bool |  | If false, the processor does not append values already present in the field. |
| field | x |  | string |  | Source field to process. |
| ignore_missing |  | x | bool |  | If true and field does not exist or is null, the processor quietly returns without modifying the document. |
| value |  |  | string |  | The value to be appended. |


### community_id

Computes the Community ID for network flow data as defined in the
[Community ID Specification](https://github.com/corelight/community-id-spec).
You can use a community ID to correlate network events related to a single
flow.

The community ID processor reads network flow data from related
Elastic Common Schema (ECS) fields by default. If you use the ECS, no
configuration is required.

| Option | Required | Optional | Type | Default | Description |
|--------|----------|----------|------|---------|-------------|
| destination_ip |  | x | string | destination.ip | Field containing the destination IP address. |
| destination_port |  | x | string | destination.port | Field containing the destination port. |
| iana_number |  | x | string | network.iana_number | Field containing the IANA number. |
| icmp_code |  | x | string | icmp.code | Field containing the ICMP code. |
| icmp_type |  | x | string | icmp.type | Field containing the ICMP type. |
| ignore_failure |  | x | bool |  | Ignore failures for the processor. |
| seed |  | x | int16 |  | Seed for the community ID hash. Must be between 0 and 65535 (inclusive). The seed can prevent hash collisions between network domains, such as a staging and production network that use the same addressing scheme. |
| source_ip |  | x | string | source.ip | Field containing the source IP address. |
| source_port |  | x | string | source.port | Field containing the source port. |
| target_field |  | x | string | network.community_id | The field to assign the output value to, by default field is updated in-place. |
| transport |  | x | string | network.transport | Field containing the transport protocol. Used only when the iana_number field is not present. |


### lowercase

Lowercase converts a string to its lowercase equivalent. If the field is an array of strings, all members of the array will be converted.

| Option | Required | Optional | Type | Default | Description |
|--------|----------|----------|------|---------|-------------|
| field | x |  | string |  | Source field to process. |
| ignore_missing |  | x | bool |  | If true and field does not exist or is null, the processor quietly returns without modifying the document. |
| target_field |  | x | string |  | The field to assign the output value to, by default field is updated in-place. |


### remove

Removes existing fields. If one field doesn’t exist the processor
will fail.

| Option | Required | Optional | Type | Default | Description |
|--------|----------|----------|------|---------|-------------|
| fields | x |  | []string |  | Source fields to remove. |
| ignore_missing |  | x | bool |  | If true and field does not exist or is null, the processor quietly returns without modifying the document. |


### set

Sets one field and associates it with the specified value. If the field
already exists, its value will be replaced with the provided one.

| Option | Required | Optional | Type | Default | Description |
|--------|----------|----------|------|---------|-------------|
| copy_from |  | x | string |  | The origin field which will be copied to target_field. |
| ignore_failure |  | x | bool |  | Ignore failures for the processor. |
| ignore_missing |  | x | bool |  | If true and field does not exist or is null, the processor quietly returns without modifying the document. |
| target_field |  | x | string |  | The field to assign the output value to, by default field is updated in-place. |
| value |  | x | any |  | The value to be set for the field. |


### uppercase

Uppercase converts a string to its uppercase equivalent. If the field is an array of strings, all members of the array will be converted.

| Option | Required | Optional | Type | Default | Description |
|--------|----------|----------|------|---------|-------------|
| field | x |  | string |  | Source field to process. |
| ignore_missing |  | x | bool |  | If true and field does not exist or is null, the processor quietly returns without modifying the document. |
| target_field |  | x | string |  | The field to assign the output value to, by default field is updated in-place. |


### webassembly

Executes a WebAssembly module to process the event.

| Option | Required | Optional | Type | Default | Description |
|--------|----------|----------|------|---------|-------------|
| file | x |  | string |  | Path to the WebAssembly module to load. Binary (`.wasm`) and text (`.wat`) formats are supported.  |
| ignore_failure |  | x | bool |  | Ignore failures for the processor. |
| params |  | x | map |  | A dictionary of parameters that are passed to module's `register` function.  |



