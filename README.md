# RALF

**RALF** is a powerful, and customizable calendar filtering and modification tool. It allows you to define complex workflows for processing calendar events from various sources, apply dynamic filters, transform event properties, and even manage context-driven event manipulation.

[ [Installation](#installation) | [Usage](#usage) | [Examples](#examples) | [Configuration](#configuration) | [License](#license) ]

## Installation

### Prerequisites

- **Go 1.22+**: Ensure you have Go installed. If not, download and install it from [golang.org](https://golang.org/).

### Clone the Repository

```console
git clone https://github.com/darmiel/ralf.git
cd ralf
```

### Build the Server

```console
go build -tags server -o ralf-server ./cmd/server
```
## Examples

You can find a few example flows in the [`examples/`](examples/) directory.

## Usage

### Source

Define the source of your calendar data, either HTTP or HTML. HTML sources require selectors to parse event details. See [`examples/`](examples/) for more details.

### Flows

Flows are the heart of RALF, where you define the logic for filtering and transforming calendar events. Use `if`, `then`, `else` for conditional logic, and `do` for actions.

### Context Management

Context variables allow you to dynamically influence the flow logic, storing intermediate results or flags for later use. Prefix the key with `$` if you want to use expressions.

```yaml
- do: ctx/set
  with:
    static: "This is always the same"
    $flag: "Event.Description().Contains('Confidential')"
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
