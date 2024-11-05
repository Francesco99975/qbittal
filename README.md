# Qbittal

**Qbittal** is a Golang-powered server designed for automating torrent downloads using Docker Compose deployment. With Qbittal, users can create recurring download tasks based on predefined search patterns. The server scrapes supported websites to find torrent files that match user-defined patterns and then initiates downloads through the qBittorrent API.

## Features

- **Recurring Task Scheduling**: Define tasks with custom recurrence intervals for automated, periodic searches and downloads.
- **Pattern-Based Torrent Searching**: Specify search patterns to target torrents matching certain criteria (e.g., by keyword, category).
- **Website Scraping for Content Discovery**: Qbittal scrapes supported torrent sites to discover and filter content.
- **qBittorrent Integration**: Uses the qBittorrent Web API to manage and download torrents.
- **Docker Compose Deployment**: Easily deployable via Docker Compose with preconfigured environment variables and settings.

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Installation](#installation)
3. [Docker Compose Files](#docker-compose-files)
4. [Usage](#usage)
5. [Supported Websites](#supported-websites)
6. [License](#license)

---

## Getting Started

These instructions will help you set up and run Qbittal on your local machine for development and testing. For production deployment, use Docker Compose as described in the sections below.

## Installation

### Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/)
- [Git](https://git-scm.com/)

### Cloning the Repository

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/qbittal.git
   cd qbittal
   ```

## Docker Compose Files

Qbittal is deployed using Docker Compose. Three Docker Compose configurations are available for different deployment scenarios:

1. [Basic Deployment](docker-compose.yml)
2. [Standalone Deployment](docker-compose_standalone.yml)
3. [Traefik Deployment](docker-compose_traefik.yml)

Ensure that you adapt these files to your environment and configuration needs before deploying.

## Usage

Once the server is running, you can interact with Qbittal via API requests. To simplify interactions, use the [Qbittal Frontend](https://github.com/Francesco99975/qbitter).

### Example Workflow

1. **Define a Search Pattern**: Send a request to create a recurring task for a specific keyword or category.
2. **Schedule the Task**: Qbittal will check for new torrents matching the search pattern at the configured interval.
3. **Download and Manage**: Qbittal communicates with qBittorrent to handle torrent downloads, updates, and removal.

## Supported Websites

Currently, Qbittal supports scraping content from the following websites:

- [Nyaa](https://nyaa.si)
- [ThePirateBay](https://thepiratebay.org)

Support for additional sites can be added by extending the scraping module.

## License

This project is licensed under the BSD 3-Clause License. See the [LICENSE](LICENSE) file for more details.

Happy torrenting with **Qbittal**!
