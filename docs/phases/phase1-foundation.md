# Phase 1: Foundation & Core Infrastructure

## Backend (Go)

- Project scaffolding with Go modules
- REST API framework setup (Gin)
- Database layer with GORM (SQLite3 dev / PostgreSQL prod)
- Core domain models: User, Country, Visit
- API versioning structure (`/api/v1/`)

### LTI 1.3 Integration

- Integrate peregrine-lti library for Canvas LTI 1.3 launch
- OIDC login endpoint (`/lti/login`)
- LTI launch endpoint (`/lti/launch`)
- JWKS endpoint for key exchange (`/.well-known/jwks.json`)
- Platform registration storage (Canvas instance config)
- Session management post-launch (JWT or session cookie)
- Extract and store Canvas user_id and course_id from launch context

## Frontend (React Native + Expo)

- Expo project initialization with TypeScript
- React Navigation setup with stack and tab navigators
- Shared components and hooks
- Auth context using LTI session (no separate login flow)
- Main tab navigation shell (6 sections)
- Responsive layout for phone, tablet, and web

### Platform Targets

- iOS (tablet-first for classroom use)
- Android (tablet and phone)
- Web (via React Native Web / Expo)

## Acceptance Criteria

- Go server starts and responds to health check at `/api/v1/health`
- LTI 1.3 launch from Canvas successfully authenticates user
- User identity persisted from Canvas LTI context (no registration form)
- Database migrations run successfully for User and Country tables
- Expo app runs on iOS simulator, Android emulator, and web browser
- Navigation to all 6 core sections works (stubbed screens)
- Both SQLite3 and PostgreSQL connections work via config switch
