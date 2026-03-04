* Frontend and backend are written in Go.
* Use slog for structured logging, always with context (e.g. request ID, user ID).
* Use OpenAPI for API documentation and client generation.
* Use GraphQL for frontend data queries.
* Use WebRTC for telephony features, with a signaling server implemented in the backend.
* Use a SIP gateway to connect WebRTC clients to a PBX system (Clarity).
* Use https://github.com/urfave/cli for parsing command line arguments and environment variables
* Backend lives in /backend, frontend in /frontend.
* Use Dockerfiles with Multi-stage builds for both frontend and backend, with a docker-compose file for local development.
