# Fish Reports Go Server

This is a lightweight Go API server that loads raw JSON data from the Fish Survey Scraper and serves it through web endpoints. It provides data in various formats for use by a client-side mobile application.

## Overview

- Data Loading: Reads raw JSON data produced by the scraper
- API Endpoints: Offers endpoints to retrieve survey data, graph data, counties, and species information
- Flexible Queries: Supports filtering, sorting, and pagination for survey data

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/MNwake/FishReportsServer.git
cd FishReportsServer
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Run the Server

```bash
go run main.go
```

The server will start on port 8080 (by default). Adjust settings as needed.

## Endpoints Overview

### Survey Data

- `GET /surveys`: Retrieve survey data with filtering, sorting, pagination, and game fish filtering

### Analytics

- `GET /graph`: Get fish count data based on day-of-week, species, and survey date

### Reference Data

- `GET /counties`: List all counties
- `GET /species`: List all species
- `GET /species/id/:species_id`: Get statistics for a specific species
- `GET /counties/id/:id`: Get details and statistics for a specific county

For more details, please refer to the source code.

Happy coding!
