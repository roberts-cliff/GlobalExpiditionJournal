# Phase 5: Integration, Polish & Production

## Backend (Go)

- API rate limiting and security hardening
- PostgreSQL migration scripts
- Data export functionality
- Performance optimization (caching, query optimization)
- Logging and monitoring setup

### LTI Advantage Services

- NRPS (Names and Role Provisioning) - Sync class rosters for teachers
- AGS (Assignment and Grade Services) - Report progress/grades back to Canvas
- Deep Linking - Allow teachers to embed specific journal activities

## Frontend (React Native + Expo)

- Cross-feature deep linking (Library <-> Passport <-> Scrapbook)
- Offline support with local storage and sync
- Kid-friendly UI polish (animations, colors per mockups)
- Accessibility pass (screen readers, contrast, touch targets)
- E2E testing with Detox (mobile) and Playwright (web)
- App store builds (iOS App Store, Google Play)
- Web deployment (Expo Web build)

## Teacher Dashboard (Canvas Integration)

- View student progress across the class
- See aggregated stats (countries visited, badges earned)
- Link journal activities to Canvas assignments

## Acceptance Criteria

- All features are interconnected with working deep links
- App works offline with cached data and syncs when online
- UI matches kid-friendly mockups with animations
- App passes accessibility audit (touch targets, contrast, VoiceOver/TalkBack)
- E2E test suite covers all user journeys
- Production PostgreSQL deployment is stable
- API handles rate limiting and returns proper error responses
- User can export their data (scrapbook, visits, awards)
- Teachers can view class progress via NRPS roster sync
- Completion data syncs to Canvas gradebook via AGS
- iOS and Android apps submitted to app stores
