# Architecture

## Stack

- **Backend**: Go (Gin framework)
- **Frontend**: React Native + Expo (iOS, Android, Web)
- **Database**: SQLite3 (local dev), PostgreSQL (production)
- **LMS Integration**: Canvas LMS via LTI 1.3

## Canvas LTI 1.3 Integration

The app integrates with Canvas LMS as an External Tool using LTI 1.3.

### Authentication Flow

1. Student clicks app link in Canvas course navigation
2. Canvas initiates OIDC login request to our backend
3. Backend validates and redirects to Canvas for authorization
4. Canvas sends signed JWT with user/course context
5. Backend validates JWT, creates session, serves Angular app

### LTI Library

Using [peregrine-lti](https://github.com/StevenWeathers/peregrine-lti) for Go LTI 1.3 launch handling.

### LTI Advantage Services (Future)

- **NRPS** (Names and Role Provisioning) - Roster sync
- **AGS** (Assignment and Grade Services) - Grade passback
- **Deep Linking** - Embed specific content in Canvas

### Data Model Considerations

- Store Canvas `user_id` and `course_id` from LTI launch context
- Link app Users to Canvas identities (no separate registration needed)
- Track which Canvas course each Visit/Scrapbook entry is associated with
