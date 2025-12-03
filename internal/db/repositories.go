package db

type Repositories struct {
	User       *UserRepository
	APIKey     *APIKeyRepository
	Chat       *ChatbotRepository
	Document   *DocumentRepository
	File       *FileRepository
	Message    *ChatMessageRepository
	Revision   *RevisionRepository
	SharedKB   *SharedKnowledgeBaseRepository
	Schedule   *CrawlScheduleRepository
	LLMUsage   *LLMUsageRepository
	Org        *OrganizationRepository
	OrgMembers *OrganizationMemberRepository
	OrgInvites *OrganizationInviteRepository
}

// NewRepositories creates all repository instances
func NewRepositories(db *Database) *Repositories {
	return &Repositories{
		User:       NewUserRepository(db),
		APIKey:     NewAPIKeyRepository(db),
		Chat:       NewChatbotRepository(db),
		Document:   NewDocumentRepository(db),
		File:       NewFileRepository(db),
		Message:    NewChatMessageRepository(db),
		Revision:   NewRevisionRepository(db),
		SharedKB:   NewSharedKnowledgeBaseRepository(db),
		Schedule:   NewCrawlScheduleRepository(db),
		LLMUsage:   NewLLMUsageRepository(db),
		Org:        NewOrganizationRepository(db),
		OrgMembers: NewOrganizationMemberRepository(db),
		OrgInvites: NewOrganizationInviteRepository(db),
	}
}
