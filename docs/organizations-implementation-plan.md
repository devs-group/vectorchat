# Organizations Implementation Plan

## Executive Summary

This plan introduces a multi-tenant organization system that enables team collaboration while maintaining the existing personal workspace experience. The implementation uses a **stateless architecture** with explicit context headers, secure invitation systems, and role-based access control.

## Key Design Principles

1. **Stateless Context**: Organization context is passed via `X-Organization-ID` header rather than stored in user session
2. **Billing Entity**: Organizations are the primary billing and subscription unit
3. **Seamless Migration**: Existing users retain personal workspaces with minimal disruption
4. **Security-First**: Token-based invitations with RBAC and proper isolation
5. **Developer Experience**: Clean separation between personal and organizational resources

## Database Schema

### Core Tables

```sql
-- Organizations: Central billing and collaboration entity
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE, -- URL routing: app.com/acme-corp/chats
    description TEXT,
    billing_email VARCHAR(255),
    stripe_customer_id VARCHAR(255),
    plan_tier VARCHAR(50) DEFAULT 'free' CHECK (plan_tier IN ('free', 'pro', 'enterprise')),
    settings JSONB DEFAULT '{}', -- Flexible org-wide settings
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Organization Members: User-to-org relationships with roles
CREATE TABLE organization_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'billing')),
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMP,
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(organization_id, user_id)
);

-- Organization Invites: Secure, time-limited invitations
CREATE TABLE organization_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'member' CHECK (role IN ('admin', 'member', 'billing')),
    token_hash VARCHAR(255) NOT NULL, -- SHA256 hash for security
    invited_by UUID REFERENCES users(id),
    message TEXT, -- Personal invitation message
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    accepted_at TIMESTAMP,
    UNIQUE(organization_id, email)
);
```

### Resource Table Updates

```sql
-- Chatbots: Can belong to user (personal) or organization
ALTER TABLE chatbots
    ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    ADD COLUMN created_by UUID NOT NULL REFERENCES users(id),
    ADD COLUMN updated_by UUID REFERENCES users(id),
    ADD CONSTRAINT chatbots_owner_check CHECK (
        (organization_id IS NULL AND created_by IS NOT NULL) OR
        (organization_id IS NOT NULL)
    );

-- Knowledge Bases: Can belong to user (personal) or organization
ALTER TABLE shared_knowledge_bases
    ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    ADD COLUMN created_by UUID NOT NULL REFERENCES users(id),
    ADD COLUMN updated_by UUID REFERENCES users(id),
    ADD CONSTRAINT shared_knowledge_bases_owner_check CHECK (
        (organization_id IS NULL AND created_by IS NOT NULL) OR
        (organization_id IS NOT NULL)
    );

-- Add indexes for performance
CREATE INDEX idx_chatbots_organization_id ON chatbots(organization_id);
CREATE INDEX idx_chatbots_created_by ON chatbots(created_by);
CREATE INDEX idx_shared_knowledge_bases_organization_id ON shared_knowledge_bases(organization_id);
CREATE INDEX idx_shared_knowledge_bases_created_by ON shared_knowledge_bases(created_by);
CREATE INDEX idx_organization_members_user_id ON organization_members(user_id);
CREATE INDEX idx_organization_members_org_id ON organization_members(organization_id);
```

## Backend Architecture

### A. Context & Middleware (Stateless Design)

#### Organization Context Middleware
```go
type OrganizationContext struct {
    ID          uuid.UUID
    Name        string
    Slug        string
    UserRole    string // 'owner', 'admin', 'member', 'billing'
    Permissions []string
}

func OrganizationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        userID := getUserIDFromContext(ctx)

        // Extract organization ID from header
        orgIDHeader := r.Header.Get("X-Organization-ID")

        var orgCtx *OrganizationContext

        if orgIDHeader == "" {
            // Personal workspace context
            orgCtx = &OrganizationContext{
                ID:          uuid.Nil,
                Name:        "Personal Workspace",
                UserRole:    "owner",
                Permissions: []string{"read", "write", "delete"},
            }
        } else {
            // Organization context
            orgID, err := uuid.Parse(orgIDHeader)
            if err != nil {
                http.Error(w, "Invalid organization ID", http.StatusBadRequest)
                return
            }

            // Verify user membership and get role
            membership, err := orgRepo.GetMembership(orgID, userID)
            if err != nil {
                http.Error(w, "Not a member of organization", http.StatusForbidden)
                return
            }

            org, err := orgRepo.GetByID(orgID)
            if err != nil {
                http.Error(w, "Organization not found", http.StatusNotFound)
                return
            }

            permissions := getRolePermissions(membership.Role)

            orgCtx = &OrganizationContext{
                ID:          orgID,
                Name:        org.Name,
                Slug:        org.Slug,
                UserRole:    membership.Role,
                Permissions: permissions,
            }
        }

        // Inject context
        ctx = context.WithValue(ctx, "organization", orgCtx)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

#### Permission System
```go
func getRolePermissions(role string) []string {
    switch role {
    case "owner":
        return []string{"read", "write", "delete", "manage_members", "manage_billing", "delete_org"}
    case "admin":
        return []string{"read", "write", "delete", "manage_members", "manage_settings"}
    case "member":
        return []string{"read", "write"}
    case "billing":
        return []string{"read_billing", "manage_billing"}
    default:
        return []string{}
    }
}

func RequirePermission(permission string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := r.Context()
            orgCtx := getOrganizationContext(ctx)

            if !contains(orgCtx.Permissions, permission) {
                http.Error(w, "Insufficient permissions", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### B. Core Services

#### OrganizationService
```go
type OrganizationService struct {
    repo    OrganizationRepository
    memberRepo OrganizationMemberRepository
    inviteRepo OrganizationInviteRepository
    emailService EmailService
    billingService BillingService
}

func (s *OrganizationService) Create(userID uuid.UUID, name, slug, description string) (*Organization, error) {
    org := &Organization{
        Name:        name,
        Slug:        slug,
        Description: description,
        CreatedBy:   userID,
    }

    createdOrg, err := s.repo.Create(org)
    if err != nil {
        return nil, err
    }

    // Creator becomes owner
    member := &OrganizationMember{
        OrganizationID: createdOrg.ID,
        UserID:         userID,
        Role:           "owner",
        JoinedAt:       time.Now(),
    }

    err = s.memberRepo.Create(member)
    if err != nil {
        return nil, err
    }

    return createdOrg, nil
}

func (s *OrganizationService) TransferResource(resourceID, resourceType, targetOrgID uuid.UUID, requestingUserID uuid.UUID) error {
    // Verify user has admin rights in target org
    membership, err := s.memberRepo.GetMembership(targetOrgID, requestingUserID)
    if err != nil || !hasAdminPermission(membership.Role) {
        return errors.New("insufficient permissions")
    }

    // Verify user owns the resource (personal workspace)
    if resourceType == "chatbot" {
        return s.transferChatbot(resourceID, targetOrgID)
    }
    // ... other resource types

    return nil
}
```

#### InvitationService
```go
type InvitationService struct {
    repo          OrganizationInviteRepository
    memberRepo    OrganizationMemberRepository
    orgRepo       OrganizationRepository
    emailService  EmailService
    cryptoService CryptoService
}

func (s *InvitationService) CreateInvite(orgID uuid.UUID, email, role, message string, invitedBy uuid.UUID) (*OrganizationInvite, error) {
    // Generate secure token
    token := generateRandomToken(32)
    tokenHash := s.cryptoService.Hash(token)

    // Check if user already exists in org
    existingUser, err := s.getUserByEmail(email)
    if err == nil {
        existingMember, err := s.memberRepo.GetMembership(orgID, existingUser.ID)
        if err == nil && existingMember != nil {
            return nil, errors.New("user already a member")
        }
    }

    invite := &OrganizationInvite{
        OrganizationID: orgID,
        Email:          email,
        Role:           role,
        TokenHash:      tokenHash,
        InvitedBy:      invitedBy,
        Message:        message,
        ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days
    }

    createdInvite, err := s.repo.Create(invite)
    if err != nil {
        return nil, err
    }

    // Send invitation email
    org, _ := s.orgRepo.GetByID(orgID)
    inviter, _ := s.getUserByID(invitedBy)

    err = s.emailService.SendInvitation(email, token, org, inviter, message)
    if err != nil {
        // Log but don't fail the creation
        log.Printf("Failed to send invitation email: %v", err)
    }

    return createdInvite, nil
}

func (s *InvitationService) AcceptInvite(token, userID uuid.UUID) error {
    // Find invite by token hash
    tokenHash := s.cryptoService.Hash(token)
    invite, err := s.repo.GetByTokenHash(tokenHash)
    if err != nil {
        return errors.New("invalid invitation token")
    }

    if time.Now().After(invite.ExpiresAt) {
        return errors.New("invitation expired")
    }

    // Get user details
    user, err := s.getUserByID(userID)
    if err != nil {
        return errors.New("user not found")
    }

    if user.Email != invite.Email {
        return errors.New("invitation email mismatch")
    }

    // Add user to organization
    member := &OrganizationMember{
        OrganizationID: invite.OrganizationID,
        UserID:         userID,
        Role:           invite.Role,
        JoinedAt:       time.Now(),
    }

    err = s.memberRepo.Create(member)
    if err != nil {
        return err
    }

    // Mark invite as accepted
    invite.AcceptedAt = time.Now()
    s.repo.Update(invite)

    return nil
}
```

### C. API Endpoints

```go
// Organization Management
GET    /api/v1/organizations                    # List user's organizations
POST   /api/v1/organizations                    # Create organization
GET    /api/v1/organizations/:id                # Get organization details
PUT    /api/v1/organizations/:id                # Update organization (admin+)
DELETE /api/v1/organizations/:id                # Delete organization (owner)

// Members Management
GET    /api/v1/organizations/:id/members        # List members (member+)
POST   /api/v1/organizations/:id/invites        # Send invite (admin+)
GET    /api/v1/organizations/:id/invites        # List pending invites (admin+)
PUT    /api/v1/organizations/:id/members/:uid   # Update member role (admin+)
DELETE /api/v1/organizations/:id/members/:uid   # Remove member (admin+)
POST   /api/v1/invites/accept                   # Accept invitation

// Contextual Resources (Header-based)
GET    /api/v1/chatbots                         # Returns personal OR org bots based on X-Organization-ID
POST   /api/v1/chatbots                         # Creates in personal OR org context
GET    /api/v1/knowledge-bases                  # Returns personal OR org knowledge bases
POST   /api/v1/knowledge-bases                  # Creates in personal OR org context

// Organization Settings
GET    /api/v1/organizations/:id/settings        # Get org settings (admin+)
PUT    /api/v1/organizations/:id/settings        # Update org settings (admin+)
GET    /api/v1/organizations/:id/billing        # Get billing info (billing role+)
PUT    /api/v1/organizations/:id/billing        # Update billing (billing role+)
```

## Frontend Implementation

### A. State Management

```typescript
// stores/organization.ts
import { defineStore } from 'pinia'

interface Organization {
  id: string
  name: string
  slug: string
  role: string
  permissions: string[]
}

export const useOrganizationStore = defineStore('organization', {
  state: () => ({
    organizations: [] as Organization[],
    currentOrganizationId: null as string | null,
    isLoading: false,
  }),

  getters: {
    currentOrganization: (state) => {
      if (!state.currentOrganizationId) {
        return {
          id: null,
          name: 'Personal Workspace',
          slug: 'personal',
          role: 'owner',
          permissions: ['read', 'write', 'delete', 'manage_billing']
        }
      }
      return state.organizations.find(org => org.id === state.currentOrganizationId)
    },

    isInOrganization: (state) => state.currentOrganizationId !== null,
    hasPermission: (state) => (permission: string) => {
      const current = state.organizations.find(org => org.id === state.currentOrganizationId)
      return current?.permissions.includes(permission) ?? false
    }
  },

  actions: {
    async switchOrganization(orgId: string | null) {
      this.currentOrganizationId = orgId
      // Update API client headers
      this.$nuxt.$api.setHeader('X-Organization-ID', orgId || '')
      // Persist to localStorage
      localStorage.setItem('current-org-id', orgId || '')
    },

    async loadOrganizations() {
      this.isLoading = true
      try {
        const response = await this.$nuxt.$api.get('/organizations')
        this.organizations = response.data
      } catch (error) {
        console.error('Failed to load organizations:', error)
      } finally {
        this.isLoading = false
      }
    }
  }
})
```

### B. API Client Configuration

```typescript
// plugins/api.ts
export default function({ $config, app }) {
  const api = $axios.create({
    baseURL: $config.apiBaseURL
  })

  // Request interceptor to add organization header
  api.onRequest(config => {
    const orgStore = useOrganizationStore(app.$pinia)
    if (orgStore.currentOrganizationId) {
      config.headers['X-Organization-ID'] = orgStore.currentOrganizationId
    }
    return config
  })

  // Response interceptor to handle organization context errors
  api.onError(error => {
    if (error.response?.status === 403 &&
        error.response?.data?.code === 'INSUFFICIENT_PERMISSIONS') {
      // Redirect or show permission error
      app.$router.push('/unauthorized')
    }
    return Promise.reject(error)
  })

  app.$api = api
}
```

### C. Vue Components

#### Organization Switcher
```vue
<!-- components/OrganizationSwitcher.vue -->
<template>
  <div class="relative">
    <button
      @click="isOpen = !isOpen"
      class="flex items-center space-x-2 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-md"
    >
      <div class="w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center">
        <span class="text-white text-xs font-bold">
          {{ currentOrg?.name?.charAt(0) || 'P' }}
        </span>
      </div>
      <span>{{ currentOrg?.name || 'Personal' }}</span>
      <ChevronDownIcon class="w-4 h-4" />
    </button>

    <div v-if="isOpen" class="absolute right-0 mt-2 w-64 bg-white rounded-lg shadow-lg border border-gray-200 z-50">
      <div class="py-1">
        <!-- Personal Workspace -->
        <button
          @click="switchToPersonal"
          class="w-full text-left px-4 py-2 hover:bg-gray-100 flex items-center space-x-3"
        >
          <div class="w-6 h-6 bg-gray-400 rounded-full flex items-center justify-center">
            <span class="text-white text-xs font-bold">P</span>
          </div>
          <div>
            <div class="font-medium">Personal Workspace</div>
            <div class="text-xs text-gray-500">Your private space</div>
          </div>
        </button>

        <div class="border-t border-gray-200 my-1"></div>

        <!-- Organizations -->
        <button
          v-for="org in organizations"
          :key="org.id"
          @click="switchToOrganization(org.id)"
          class="w-full text-left px-4 py-2 hover:bg-gray-100 flex items-center space-x-3"
        >
          <div class="w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center">
            <span class="text-white text-xs font-bold">{{ org.name.charAt(0) }}</span>
          </div>
          <div>
            <div class="font-medium">{{ org.name }}</div>
            <div class="text-xs text-gray-500 capitalize">{{ org.role }}</div>
          </div>
        </button>

        <div class="border-t border-gray-200 my-1"></div>

        <!-- Create Organization -->
        <button
          @click="createOrganization"
          class="w-full text-left px-4 py-2 hover:bg-gray-100 flex items-center space-x-3 text-blue-600"
        >
          <PlusIcon class="w-5 h-5" />
          <span>Create Organization</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ChevronDownIcon, PlusIcon } from '@heroicons/vue/24/solid'

const orgStore = useOrganizationStore()
const isOpen = ref(false)

const organizations = computed(() => orgStore.organizations)
const currentOrg = computed(() => orgStore.currentOrganization)

const switchToPersonal = async () => {
  await orgStore.switchOrganization(null)
  isOpen.value = false
}

const switchToOrganization = async (orgId: string) => {
  await orgStore.switchOrganization(orgId)
  isOpen.value = false
}

const createOrganization = () => {
  navigateTo('/organizations/new')
  isOpen.value = false
}

// Close dropdown when clicking outside
onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

const handleClickOutside = (event: Event) => {
  if (!event.target.closest('.relative')) {
    isOpen.value = false
  }
}
</script>
```

#### Member Management Component
```vue
<!-- components/organizations/MemberList.vue -->
<template>
  <div class="bg-white rounded-lg shadow">
    <div class="px-6 py-4 border-b border-gray-200">
      <div class="flex justify-between items-center">
        <h3 class="text-lg font-medium text-gray-900">Team Members</h3>
        <button
          v-if="canManageMembers"
          @click="showInviteModal = true"
          class="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 text-sm"
        >
          Invite Member
        </button>
      </div>
    </div>

    <div class="divide-y divide-gray-200">
      <div
        v-for="member in members"
        :key="member.id"
        class="px-6 py-4 flex items-center justify-between"
      >
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-gray-300 rounded-full flex items-center justify-center">
            <span class="text-gray-600 font-medium">
              {{ member.name.charAt(0).toUpperCase() }}
            </span>
          </div>
          <div>
            <div class="font-medium text-gray-900">{{ member.name }}</div>
            <div class="text-sm text-gray-500">{{ member.email }}</div>
          </div>
        </div>

        <div class="flex items-center space-x-3">
          <select
            v-if="canManageMembers && member.id !== currentUserId"
            :value="member.role"
            @change="updateRole(member.id, $event.target.value)"
            class="text-sm border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="member">Member</option>
            <option value="admin">Admin</option>
            <option value="billing">Billing</option>
          </select>
          <span v-else class="text-sm text-gray-600 capitalize">{{ member.role }}</span>

          <button
            v-if="canManageMembers && member.id !== currentUserId"
            @click="removeMember(member.id)"
            class="text-red-600 hover:text-red-800 text-sm"
          >
            Remove
          </button>
        </div>
      </div>
    </div>

    <!-- Invite Modal -->
    <InviteMembersModal
      v-if="showInviteModal"
      :organization-id="organizationId"
      @close="showInviteModal = false"
      @invited="loadMembers"
    />
  </div>
</template>

<script setup>
import { useOrganizationStore } from '~/stores/organization'

const props = defineProps<{
  organizationId: string
}>()

const orgStore = useOrganizationStore()
const members = ref([])
const showInviteModal = ref(false)
const currentUserId = ref(null)

const canManageMembers = computed(() =>
  orgStore.hasPermission('manage_members')
)

const loadMembers = async () => {
  try {
    const response = await $api.get(`/organizations/${props.organizationId}/members`)
    members.value = response.data
  } catch (error) {
    console.error('Failed to load members:', error)
  }
}

const updateRole = async (memberId: string, newRole: string) => {
  try {
    await $api.put(`/organizations/${props.organizationId}/members/${memberId}`, {
      role: newRole
    })
    await loadMembers()
  } catch (error) {
    console.error('Failed to update role:', error)
  }
}

const removeMember = async (memberId: string) => {
  if (!confirm('Are you sure you want to remove this member?')) {
    return
  }

  try {
    await $api.delete(`/organizations/${props.organizationId}/members/${memberId}`)
    await loadMembers()
  } catch (error) {
    console.error('Failed to remove member:', error)
  }
}

onMounted(() => {
  loadMembers()
  // Load current user info
})
</script>
```

### D. Page Structure

```
pages/
├── organizations/
│   ├── index.vue                    # List and create organizations
│   ├── new.vue                      # Create new organization form
│   ├── [slug]/
│   │   ├── index.vue               # Organization dashboard
│   │   ├── members.vue             # Member management
│   │   ├── settings.vue            # Organization settings
│   │   └── billing.vue             # Billing management
│   └── invites/
│       └── accept/
│           └── [token].vue         # Accept invitation
├── chatbots/
│   ├── index.vue                   # Updated to show org context
│   └── [id].vue                    # Updated to show org context
└── knowledge-bases/
    ├── index.vue                   # Updated to show org context
    └── [id].vue                    # Updated to show org context
```

### E. Updated Navigation

```vue
<!-- layouts/default.vue -->
<template>
  <div class="min-h-screen bg-gray-50">
    <nav class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex items-center space-x-8">
            <NuxtLink to="/" class="text-xl font-bold text-gray-900">
              VectorChat
            </NuxtLink>

            <div class="hidden md:flex space-x-4">
              <NuxtLink
                to="/chatbots"
                class="text-gray-600 hover:text-gray-900 px-3 py-2 text-sm font-medium"
              >
                Chats
              </NuxtLink>
              <NuxtLink
                to="/knowledge-bases"
                class="text-gray-600 hover:text-gray-900 px-3 py-2 text-sm font-medium"
              >
                Knowledge Bases
              </NuxtLink>
              <NuxtLink
                v-if="orgStore.isInOrganization"
                :to="`/organizations/${currentOrgSlug}/members`"
                class="text-gray-600 hover:text-gray-900 px-3 py-2 text-sm font-medium"
              >
                Team
              </NuxtLink>
            </div>
          </div>

          <div class="flex items-center space-x-4">
            <OrganizationSwitcher />
            <UserMenu />
          </div>
        </div>
      </div>
    </nav>

    <main>
      <slot />
    </main>
  </div>
</template>

<script setup>
const orgStore = useOrganizationStore()
const currentOrgSlug = computed(() => orgStore.currentOrganization?.slug)
</script>
```

## Migration Strategy

### Phase 1: Database Foundation (Week 1-2)

1. **Database Migrations**
   ```sql
   -- Migration 018: Add organizations
   CREATE TABLE organizations (...);

   -- Migration 019: Add organization members
   CREATE TABLE organization_members (...);

   -- Migration 020: Add organization invites
   CREATE TABLE organization_invites (...);

   -- Migration 021: Update existing tables
   ALTER TABLE chatbots ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL;
   ALTER TABLE chatbots ADD COLUMN created_by UUID NOT NULL REFERENCES users(id);
   ALTER TABLE chatbots ADD COLUMN updated_by UUID REFERENCES users(id);
   ```

2. **Data Migration Script**
   ```sql
   -- Set existing chatbots to personal (organization_id = NULL)
   UPDATE chatbots SET created_by = user_id WHERE created_by IS NULL;

   -- Set existing knowledge bases to personal
   UPDATE shared_knowledge_bases SET created_by = user_id WHERE created_by IS NULL;
   ```

### Phase 2: Backend Implementation (Week 3-4)

1. **Core Services**
   - OrganizationService with CRUD operations
   - InvitationService with secure token handling
   - Repository layer updates for organization scoping

2. **Middleware & Context**
   - Organization context middleware
   - Permission-based authorization
   - Updated repository queries

3. **API Endpoints**
   - Organization management endpoints
   - Member management endpoints
   - Context-aware resource endpoints

### Phase 3: Frontend Implementation (Week 5-6)

1. **State Management**
   - Organization store setup
   - API client configuration
   - Organization context handling

2. **Core Components**
   - Organization switcher
   - Member management interface
   - Invitation system

3. **Page Updates**
   - Organization management pages
   - Updated resource pages with context
   - Navigation updates

### Phase 4: Invitation System & Testing (Week 7-8)

1. **Email Integration**
   - Invitation email templates
   - Email service integration
   - Token-based acceptance flow

2. **Testing & QA**
   - Unit tests for all services
   - Integration tests for API endpoints
   - E2E tests for key user flows
   - Performance testing

3. **Documentation & Deployment**
   - API documentation updates
   - User documentation
   - Deployment checklist

## Security Considerations

### 1. Authentication & Authorization
- All organization-scoped requests require valid authentication
- Role-based access control (RBAC) for all operations
- Permission checks at middleware level

### 2. Invitation Security
- Cryptographically secure token generation
- Token hashing for database storage
- Expiration-based access control
- Email verification for invite acceptance

### 3. Data Isolation
- Database-level organization isolation
- Query filtering based on organization context
- Audit logging for all organization operations

### 4. XSS & CSRF Protection
- Proper input sanitization
- CSRF tokens for state-changing operations
- Content Security Policy headers

## Performance Optimizations

### 1. Database Indexing
```sql
-- Critical indexes for performance
CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_organization_members_user_org ON organization_members(user_id, organization_id);
CREATE INDEX idx_organization_invites_token_hash ON organization_invites(token_hash);
CREATE INDEX idx_chatbots_org_created ON chatbots(organization_id, created_at DESC);
```

### 2. Caching Strategy
- Cache organization memberships for frequently accessed users
- Cache organization permissions for active sessions
- Implement query result caching for organization lists

### 3. Query Optimization
- Use JOIN queries efficiently for member data
- Implement pagination for large organization member lists
- Optimize resource queries with proper filtering

## Monitoring & Analytics

### 1. Key Metrics
- Organization creation rate
- Member invitation acceptance rate
- Resource creation per organization
- User activity patterns (personal vs organization)

### 2. Error Tracking
- Failed invitation attempts
- Permission denied errors
- Organization context validation failures
- Resource access violations

### 3. Performance Monitoring
- Database query performance for organization operations
- API response times for organization endpoints
- Frontend performance for organization switching

## Success Criteria

1. **Functional Requirements**
   ✅ Users can create and manage organizations
   ✅ Secure invitation system with token-based access
   ✅ Role-based permissions for all operations
   ✅ Seamless switching between personal and organization contexts

2. **Performance Requirements**
   ✅ Organization switching completes in < 500ms
   ✅ Organization member list loads in < 1s
   ✅ Resource queries scale with organization size

3. **Security Requirements**
   ✅ All organization operations properly authorized
   ✅ Invitation tokens are cryptographically secure
   ✅ Data isolation between organizations enforced

4. **User Experience Requirements**
   ✅ Intuitive organization management interface
   ✅ Clear distinction between personal and organization resources
   ✅ Smooth onboarding for team collaboration

This implementation plan provides a robust, secure, and scalable organization system that enhances VectorChat's collaborative capabilities while maintaining the excellent user experience of the personal workspace model.