# Phase 2: Passport & Scrapbook

## Backend (Go)

- Countries API (list, detail, visited status)
- Visits API (CRUD for country visits with dates)
- Scrapbook entries API (photos, notes, tags)
- File upload handling for media (local storage dev, S3/cloud production)
- Country-to-scrapbook linking
- All data scoped to Canvas user context from LTI launch

## Frontend (React Native + Expo)

- Passport screen with interactive world map (react-native-maps or SVG)
- Passport list view with filtering and search
- Country detail screen
- Scrapbook gallery per country (FlatList/grid layout)
- Add/edit memory modal (image picker, notes, location tags)
- Stats dashboard component
- Tablet-optimized layouts with master-detail views

## Acceptance Criteria

- User can view all countries on an interactive map
- Tapping a country shows its detail page with visit history
- User can mark a country as visited with a date
- User can add scrapbook entries with photos and notes
- Scrapbook entries are linked to specific countries
- User stats (countries visited, entries created) display correctly
- All data is associated with the user's Canvas identity
- Works on iOS, Android, and Web
