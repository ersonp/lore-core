package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
)

// validTypeNameRegex allows alphanumeric and underscores only.
var validTypeNameRegex = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// EntityTypeService manages entity types.
type EntityTypeService struct {
	relationalDB ports.RelationalDB
	cache        map[string]*entities.EntityType
	cacheMu      sync.RWMutex
}

// NewEntityTypeService creates a new EntityTypeService.
func NewEntityTypeService(relationalDB ports.RelationalDB) *EntityTypeService {
	return &EntityTypeService{
		relationalDB: relationalDB,
		cache:        make(map[string]*entities.EntityType),
	}
}

// LoadDefaults seeds the default entity types into the database.
func (s *EntityTypeService) LoadDefaults(ctx context.Context) error {
	for _, et := range entities.DefaultEntityTypes {
		existing, err := s.relationalDB.FindEntityType(ctx, et.Name)
		if err != nil {
			return fmt.Errorf("checking entity type %s: %w", et.Name, err)
		}
		if existing == nil {
			etCopy := et
			if err := s.relationalDB.SaveEntityType(ctx, &etCopy); err != nil {
				return fmt.Errorf("seeding entity type %s: %w", et.Name, err)
			}
		}
	}
	s.invalidateCache()
	return nil
}

// List returns all entity types.
func (s *EntityTypeService) List(ctx context.Context) ([]entities.EntityType, error) {
	return s.relationalDB.ListEntityTypes(ctx)
}

// Add creates a new custom entity type.
func (s *EntityTypeService) Add(ctx context.Context, name, description string) error {
	name = strings.ToLower(strings.TrimSpace(name))

	if !validTypeNameRegex.MatchString(name) {
		return errors.New("invalid type name: must be lowercase alphanumeric with underscores, starting with a letter")
	}

	existing, err := s.relationalDB.FindEntityType(ctx, name)
	if err != nil {
		return fmt.Errorf("checking entity type: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("entity type '%s' already exists", name)
	}

	et := &entities.EntityType{
		Name:        name,
		Description: description,
	}
	if err := s.relationalDB.SaveEntityType(ctx, et); err != nil {
		return fmt.Errorf("saving entity type: %w", err)
	}

	s.invalidateCache()
	return nil
}

// Remove deletes a custom entity type.
func (s *EntityTypeService) Remove(ctx context.Context, name string) error {
	if entities.IsDefaultType(name) {
		return fmt.Errorf("cannot remove default entity type '%s'", name)
	}

	existing, err := s.relationalDB.FindEntityType(ctx, name)
	if err != nil {
		return fmt.Errorf("checking entity type: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("entity type '%s' not found", name)
	}

	if err := s.relationalDB.DeleteEntityType(ctx, name); err != nil {
		return fmt.Errorf("deleting entity type: %w", err)
	}

	s.invalidateCache()
	return nil
}

// IsValid checks if a type name is valid (exists in database).
func (s *EntityTypeService) IsValid(ctx context.Context, name string) bool {
	s.cacheMu.RLock()
	if _, ok := s.cache[name]; ok {
		s.cacheMu.RUnlock()
		return true
	}
	s.cacheMu.RUnlock()

	// Cache miss - load from database
	types, err := s.relationalDB.ListEntityTypes(ctx)
	if err != nil {
		return false
	}

	s.cacheMu.Lock()
	s.cache = make(map[string]*entities.EntityType)
	for i := range types {
		s.cache[types[i].Name] = &types[i]
	}
	s.cacheMu.Unlock()

	s.cacheMu.RLock()
	_, ok := s.cache[name]
	s.cacheMu.RUnlock()
	return ok
}

// GetValidTypes returns all valid type names.
func (s *EntityTypeService) GetValidTypes(ctx context.Context) ([]string, error) {
	types, err := s.relationalDB.ListEntityTypes(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(types))
	for i, t := range types {
		names[i] = t.Name
	}
	return names, nil
}

// BuildPromptTypeList builds a comma-separated list for LLM prompts.
func (s *EntityTypeService) BuildPromptTypeList(ctx context.Context) (string, error) {
	names, err := s.GetValidTypes(ctx)
	if err != nil {
		return "", err
	}
	return strings.Join(names, ", "), nil
}

func (s *EntityTypeService) invalidateCache() {
	s.cacheMu.Lock()
	s.cache = make(map[string]*entities.EntityType)
	s.cacheMu.Unlock()
}
