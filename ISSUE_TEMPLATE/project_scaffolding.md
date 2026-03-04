## Goal
Prepare the core directory/file structure according to our architecture and Copilot instructions. This sets up clear separation for Go backend, Go frontend, and future docker setups.

---

### Acceptance criteria
- The repo contains the following top-level directories:
  - `/backend` (for all Go backend code)
  - `/frontend` (for all Go frontend/server/static assets)
- Each subproject has a `/cmd/server/main.go` (can just print hello world)
- Add dummy `go.mod` in backend and frontend
- Include clear `README.md` at root pointing to both subprojects
- Add `.dockerignore` and `.gitignore` (Go, Docker, Node)

---

No implementation code required yet, just the skeleton structure above.
Leave stubs/todos where appropriate (e.g., directories for internal packages, static assets, etc.)