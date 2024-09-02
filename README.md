# esaj-collector

The first open-source brazilian court data collector.

# Features

- Collects data from the Brazilian court system
- Uses machine learning to extract relevant information from the data
- Provides a REST API to access the data

# What is possible to do with this data?

## Collect

- Collect all data related to a specific OAB number
- Collect all data related to a specific process
- Download all PDFs documents related to a specific process

## Parse

- Extract unstructured data from PDFs and transform it into structured data.
- Use AI to extract relevant information from the data

# Command Line Examples

## Collect all data related to a specific OAB number

```bash
esaj-collector collect --oab 123456
```

## Collect all data related to a specific process

```bash
esaj-collector collect --process 123456
```

## Download all PDFs documents related to a specific process

```bash
esaj-collector download --process 123456
```

# Getting Started

`make help`

# Environment Variables

- ESAJ_USERNAME
- ESAJ_PASSWORD
- LLAMA_CLOUD_API_KEY
- OPENAI_API_TOKEN
