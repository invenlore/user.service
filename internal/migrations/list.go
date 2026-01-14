package migrations

import (
	"github.com/invenlore/core/pkg/migrator"
)

func List() []migrator.Migration {
	return []migrator.Migration{
		Migration_20260114_UsersCollection_1,
		Migration_20260114_UsersUniqueEmailIndex_1,
	}
}
