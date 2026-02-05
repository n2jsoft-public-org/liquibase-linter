# CI/CD Integration Guide

This guide explains how to integrate Liquibase Linter into your CI/CD pipelines.

## Exit Codes

The linter uses standard exit codes:

- `0`: No violations found
- `1`: Violations found (based on severity threshold)
- `2`: Error during execution (file not found, parse error, etc.)

## GitHub Actions

Create `.github/workflows/liquibase-lint.yml`:

```yaml
name: Liquibase Linting

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  lint:
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Download Liquibase Linter
      run: |
        curl -L -o liquibase-linter https://github.com/n2jsoft/liquibase-linter/releases/latest/download/liquibase-linter-linux-amd64
        chmod +x liquibase-linter

    - name: Run Liquibase Linter
      run: |
        ./liquibase-linter check --format=json db/changelog/ > results.json
        exit_code=$?
        if [ $exit_code -ne 0 ]; then
          cat results.json
          exit $exit_code
        fi

    - name: Upload results
      if: always()
      uses: actions/upload-artifact@v4
      with:
        name: linter-results
        path: results.json
```

## GitLab CI

Add to `.gitlab-ci.yml`:

```yaml
liquibase-lint:
  stage: test
  image: alpine:latest
  before_script:
    - apk add --no-cache curl
    - curl -L -o liquibase-linter https://github.com/n2jsoft/liquibase-linter/releases/latest/download/liquibase-linter-linux-amd64
    - chmod +x liquibase-linter
  script:
    - ./liquibase-linter check --format=json db/changelog/
  artifacts:
    when: always
    reports:
      junit: results.xml
```

## Jenkins

Create a `Jenkinsfile`:

```groovy
pipeline {
    agent any
    
    stages {
        stage('Liquibase Lint') {
            steps {
                sh '''
                    curl -L -o liquibase-linter https://github.com/n2jsoft/liquibase-linter/releases/latest/download/liquibase-linter-linux-amd64
                    chmod +x liquibase-linter
                    ./liquibase-linter check --format=junit db/changelog/ > results.xml
                '''
            }
        }
    }
    
    post {
        always {
            junit 'results.xml'
        }
    }
}
```

## CircleCI

Add to `.circleci/config.yml`:

```yaml
version: 2.1

jobs:
  lint:
    docker:
      - image: cimg/base:stable
    steps:
      - checkout
      - run:
          name: Download Liquibase Linter
          command: |
            curl -L -o liquibase-linter https://github.com/n2jsoft/liquibase-linter/releases/latest/download/liquibase-linter-linux-amd64
            chmod +x liquibase-linter
      - run:
          name: Run Linter
          command: ./liquibase-linter check --format=json db/changelog/
      - store_artifacts:
          path: results.json

workflows:
  version: 2
  build_and_test:
    jobs:
      - lint
```

## Docker

Create a `Dockerfile`:

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY . .
RUN go build -o liquibase-linter ./cmd/liquibase-linter

FROM alpine:latest

RUN apk --no-cache add ca-certificates
COPY --from=builder /build/liquibase-linter /usr/local/bin/

ENTRYPOINT ["liquibase-linter"]
CMD ["check", "/workspace"]
```

Build and use:

```bash
docker build -t liquibase-linter .
docker run -v $(pwd)/db:/workspace liquibase-linter
```

## Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

echo "Running Liquibase Linter..."

if ! command -v liquibase-linter &> /dev/null; then
    echo "liquibase-linter not found. Please install it first."
    exit 1
fi

# Run linter on staged files
git diff --cached --name-only --diff-filter=ACM | grep -E '\.(xml|yaml|yml|sql)$' | while read file; do
    liquibase-linter check "$file"
    if [ $? -ne 0 ]; then
        echo "Linting failed for $file"
        exit 1
    fi
done
```

Make it executable:

```bash
chmod +x .git/hooks/pre-commit
```

## Best Practices

1. **Fail Fast**: Set appropriate severity thresholds for your pipeline
2. **Cache Binary**: Cache the linter binary to speed up CI runs
3. **Parallel Execution**: Split large changesets across multiple jobs
4. **Reporting**: Use JSON/SARIF formats for better integration
5. **Ignore Patterns**: Exclude test fixtures and generated files

## Example Configurations

### Strict Mode (Block on Any Issue)

```bash
liquibase-linter check --severity=info db/changelog/
```

### Production Pipeline (Critical Issues Only)

```bash
liquibase-linter check --severity=critical db/changelog/
```

### Development Pipeline (All Warnings)

```bash
liquibase-linter check --severity=warning db/changelog/
```

## Troubleshooting

### Linter Not Found

Ensure the binary is in your PATH or use the full path to the executable.

### Permission Denied

Make sure the binary has execute permissions:

```bash
chmod +x liquibase-linter
```

### Exit Code 2

Check for syntax errors in your configuration file or changelog files.

## Next Steps

- Learn about [configuration](configuration.md)
- Review [available rules](rules.md)
