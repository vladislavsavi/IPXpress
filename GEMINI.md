# IPXpress Project Analysis

This file serves as a foundational context for Gemini CLI when working on the IPXpress project.

## Project Purpose
IPXpress is a high-performance image processing service written in Go. It provides a RESTful API to fetch, transform, and transcode images on-the-fly, leveraging the speed and low memory footprint of `libvips`.

## Technology Stack
- **Language:** Go
- **Core Library:** `libvips` (via `govips` bindings)
- **Deployment:** Docker support available

## Architecture Overview
The project follows a modular design:
- **HTTP Layer (`pkg/ipxpress/server.go`):** Handles requests, manages concurrency (semaphores), and coordinates components.
- **Core Processor (`pkg/ipxpress/ipxpress.go`):** Chainable API for transformations using `libvips`.
- **Caching (`pkg/ipxpress/cache.go`):** In-memory TTL cache for results and errors.
- **Fetcher (`pkg/ipxpress/fetcher.go`):** Optimized HTTP client for remote image retrieval.
- **Parameter Parsing (`pkg/ipxpress/params.go`):** Flexible query parameter handling (short and long forms).

## Main Components
- **`Processor`**: Wraps `libvips` operations.
- **`Handler`**: The main HTTP entry point.
- **`InMemoryCache`**: Thread-safe caching mechanism.
- **`ProcessingParams`**: Transformation request representation.

## Notable Features
- **Formats:** JPEG, PNG, GIF, WebP, AVIF.
- **Transformations:** Resize (Lanczos3), Crop, Blur, Sharpen, HSB modulation, Gamma correction.
- **Performance:** Concurrency limiting (default 256), memory-efficient processing, ETag/Cache-Control support.

## Project Structure
- `cmd/ipxpress/`: CLI entry point.
- `pkg/ipxpress/`: Core logic and library implementation.
- `examples/`: Usage demonstrations.
- `ARCHITECTURE.md`, `API.md`, `CHANGES.md`: Detailed documentation.
