package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
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
	sortedNames  []string // cached sorted names, populated with cache
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
// Optimized to list once then insert missing, reducing from O(n) Find calls to O(1) List.
func (s *EntityTypeService) LoadDefaults(ctx context.Context) error {
	existing, err := s.relationalDB.ListEntityTypes(ctx)
	if err != nil {
		return fmt.Errorf("listing entity types: %w", err)
	}

	existingSet := make(map[string]bool, len(existing))
	for _, et := range existing {
		existingSet[et.Name] = true
	}

	for _, et := range entities.DefaultEntityTypes {
		if !existingSet[et.Name] {
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

// Get returns a specific entity type by name, or nil if not found.
func (s *EntityTypeService) Get(ctx context.Context, name string) (*entities.EntityType, error) {
	return s.relationalDB.FindEntityType(ctx, name)
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
	// Fast path: check cache with read lock
	s.cacheMu.RLock()
	if len(s.cache) > 0 {
		_, ok := s.cache[name]
		s.cacheMu.RUnlock()
		return ok
	}
	s.cacheMu.RUnlock()

	// Slow path: need to populate cache
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	// Double-check: another goroutine may have populated the cache
	if len(s.cache) > 0 {
		_, ok := s.cache[name]
		return ok
	}

	// Cache miss - load from database
	types, err := s.relationalDB.ListEntityTypes(ctx)
	if err != nil {
		return false
	}

	s.populateCacheFromTypes(types)
	_, ok := s.cache[name]
	return ok
}

// GetValidTypes returns all valid type names.
// Uses the same cache as IsValid for efficiency during batch operations.
// The returned slice is shared and must not be modified by callers.
func (s *EntityTypeService) GetValidTypes(ctx context.Context) ([]string, error) {
	// Fast path: check cache with read lock
	s.cacheMu.RLock()
	if len(s.cache) > 0 {
		names := s.sortedNames
		s.cacheMu.RUnlock()
		return names, nil
	}
	s.cacheMu.RUnlock()

	// Slow path: need to populate cache
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	// Double-check: another goroutine may have populated the cache
	if len(s.cache) > 0 {
		return s.sortedNames, nil
	}

	// Cache miss - load from database
	types, err := s.relationalDB.ListEntityTypes(ctx)
	if err != nil {
		return nil, err
	}

	s.populateCacheFromTypes(types)
	return s.sortedNames, nil
}

// populateCacheFromTypes fills the cache and sortedNames from a types slice.
// Caller must hold cacheMu write lock.
func (s *EntityTypeService) populateCacheFromTypes(types []entities.EntityType) {
	s.cache = make(map[string]*entities.EntityType, len(types))
	s.sortedNames = make([]string, len(types))
	for i := range types {
		s.cache[types[i].Name] = &types[i]
		s.sortedNames[i] = types[i].Name
	}
	sort.Strings(s.sortedNames)
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
	s.sortedNames = nil
	s.cacheMu.Unlock()
}
