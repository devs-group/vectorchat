# Organizations Feature Plan

## Overview
Introduce a multi-tenant organization system that allows users to collaborate on chatbots and knowledge bases while maintaining the current user-centric model as a fallback.

## Database Schema Changes

### New Tables
```sql
-- Organizations table
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Organization members table
CREATE TABLE organization_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member', -- 'owner', 'admin', 'member'
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMP,
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, user_id)
);

-- Update existing tables to support organizations
ALTER TABLE chatbots ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE shared_knowledge_bases ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
```

## Backend Implementation

### Models & Types
- Add `Organization` and `OrganizationMember` models
- Update `User` model to include `current_organization_id`
- Add organization roles enum: `owner`, `admin`, `member`

### Services
1. **OrganizationService**
   - Create, update, delete organizations
   - Manage member invitations and roles
   - Handle organization switching

2. **InvitationService**
   - Generate invitation tokens
   - Send email invitations
   - Process invitation acceptance

### API Endpoints
```
GET    /organizations                    # List user's organizations
POST   /organizations                    # Create organization
GET    /organizations/:id                # Get organization details
PUT    /organizations/:id                # Update organization
DELETE /organizations/:id                # Delete organization

GET    /organizations/:id/members        # List members
POST   /organizations/:id/invite         # Invite member
PUT    /organizations/:id/members/:id    # Update member role
DELETE /organizations/:id/members/:id    # Remove member

POST   /organizations/switch             # Switch current organization
GET    /organizations/current            # Get current organization context
```

### Middleware Updates
- Update `AuthMiddleware` to include organization context
- Add `OrganizationMiddleware` for organization-scoped routes
- Update ownership checks to consider organization permissions

## Frontend Implementation

### New Pages & Components
1. **Organization Management**
   - `/organizations` - List and create organizations
   - `/organizations/:id` - Organization dashboard
   - `/organizations/:id/members` - Member management
   - `/organizations/:id/settings` - Organization settings

2. **Invitation System**
   - Invitation acceptance modal
   - Invite members form
   - Pending invitations list

3. **Organization Switcher**
   - Header dropdown for organization switching
   - Personal workspace option

### UX Flow for New Users

#### Initial Account Creation (No Organization)
1. User signs up → lands in **Personal Workspace**
2. Clear messaging: "This is your personal workspace. Create an organization to collaborate with others."
3. Two prominent CTAs:
   - "Create Chatbot" (personal workspace)
   - "Create Organization" (team collaboration)

#### Organization Creation Flow
1. Simple form: Organization name, optional description
2. Auto-generate unique slug (editable)
3. Creator becomes organization owner
4. Onboarding: Invite team members or skip

#### Invitation System
**Inviting Members:**
1. Enter email address
2. Select role (admin/member)
3. Personalized email invitation with secure link
4. Invitation expires in 7 days

**Accepting Invitations:**
1. User clicks invitation link
2. If not logged in → redirect to signup/login
3. Post-auth → show invitation modal
4. Accept/Decline with one click
5. Auto-switch to new organization

### Navigation & Context

#### Updated Sidebar
```
VectorChat [Organization Switcher Dropdown]
├── Chats
├── Knowledge Bases  
├── Organization Members (if org member)
├── Organization Settings (if admin/owner)
├── Subscription
└── API Settings
```

#### Organization Switcher
- Dropdown in header showing:
  - Current organization name & avatar
  - Personal workspace option
  - Other organizations (if member)
  - "Create Organization" option

### Permission System

#### Roles & Permissions
- **Owner**: Full control, can delete organization
- **Admin**: Manage members, resources, settings
- **Member**: View and create resources, invite members (if enabled)

#### Resource Scoping
- Chatbots and knowledge bases belong to either user or organization
- Organization resources visible to all members based on role
- Personal resources remain private

## Migration Strategy

### Phase 1: Backend Foundation
1. Database migrations
2. Core services and APIs
3. Update existing APIs to be organization-aware

### Phase 2: Frontend Integration
1. Organization switcher component
2. Organization management pages
3. Update existing pages to show organization context

### Phase 3: Invitation System
1. Email invitation flow
2. Acceptance modals
3. Member management interface

### Phase 4: Migration of Existing Data
1. Create default personal organization for existing users
2. Keep existing chatbots/knowledge bases as personal resources
3. Allow users to move resources to organizations

## Technical Considerations

### Security
- Invitation tokens with expiration
- Role-based access control (RBAC)
- Organization isolation in queries

### Performance
- Database indexes for organization queries
- Caching organization membership
- Efficient permission checks

### Email Templates
- Invitation email with clear CTA
- Organization welcome email
- Role change notifications

## Testing Strategy

### Backend Tests
- Organization CRUD operations
- Permission enforcement
- Invitation flow edge cases

### Frontend Tests
- Organization switching
- Permission-based UI visibility
- Invitation acceptance flow

### Integration Tests
- End-to-end organization creation
- Multi-user collaboration scenarios
- Resource sharing within organizations

## Implementation Details

### Database Migration Files
- `018_add_organizations.sql` - Create organizations and organization_members tables
- `019_add_organization_to_resources.sql` - Add organization_id to existing tables
- `020_add_user_current_organization.sql` - Add current_organization_id to users table

### File Structure
```
internal/
├── api/
│   ├── organization_handler.go
│   └── invitation_handler.go
├── services/
│   ├── organization_service.go
│   └── invitation_service.go
├── db/
│   ├── organization_repository.go
│   └── organization_member_repository.go
└── middleware/
    └── organization_middleware.go

pkg/models/
├── organization.go
└── organization_member.go

frontend/
├── pages/
│   └── organizations/
│       ├── index.vue
│       ├── [id].vue
│       ├── [id]/members.vue
│       └── [id]/settings.vue
├── components/
│   ├── organization/
│   │   ├── OrganizationSwitcher.vue
│   │   ├── InviteMembersModal.vue
│   │   └── MemberList.vue
│   └── invitations/
│       └── InvitationModal.vue
└── composables/
    └── useOrganizations.ts
```

This plan provides a smooth transition from individual to team usage while maintaining excellent UX for both scenarios. The personal workspace ensures new users can immediately start using the product, while the organization system enables seamless team collaboration when needed.
