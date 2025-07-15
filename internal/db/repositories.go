package db

type Repositories struct {
	User     *UserRepository
	Chat     *ChatbotRepository
	Document *DocumentRepository
	File     *FileRepository
	APIKey   *APIKeyRepository
}

func NewRepositories(db *Database) *Repositories {
	return &Repositories{
		User:     NewUserRepository(db),
		Chat:     NewChatbotRepository(db),
		Document: NewDocumentRepository(db),
		File:     NewFileRepository(db),
		APIKey:   NewAPIKeyRepository(db),
	}
}
