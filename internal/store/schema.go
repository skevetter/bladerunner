package store

import (
	"github.com/pocketbase/pocketbase/core"
)

func EnsureCollections(app core.App) error {
	// Repositories collection
	repositories := core.NewBaseCollection("repositories")
	repositories.Fields.Add(
		&core.TextField{
			Name:     "owner",
			Required: true,
		},
		&core.TextField{
			Name:     "name",
			Required: true,
		},
		&core.TextField{
			Name:     "token",
			Required: false,
		},
	)
	repositories.AddIndex("idx_owner_name", true, "owner, name", "")

	// Runners collection
	runners := core.NewBaseCollection("runners")
	runners.Fields.Add(
		&core.TextField{
			Name:     "name",
			Required: true,
		},
		&core.SelectField{
			Name:     "status",
			Required: true,
			Values:   []string{"idle", "busy", "offline"},
		},
		&core.TextField{
			Name:     "container_id",
			Required: false,
		},
	)

	// Jobs collection
	jobs := core.NewBaseCollection("jobs")
	jobs.Fields.Add(
		&core.RelationField{
			Name:          "repo_id",
			Required:      true,
			CollectionId:  "repositories",
			CascadeDelete: true,
			MaxSelect:     1,
		},
		&core.SelectField{
			Name:      "status",
			Required:  true,
			Values:    []string{"queued", "in_progress", "completed", "failed"},
			MaxSelect: 1,
		},
		&core.RelationField{
			Name:          "runner_id",
			Required:      false,
			CollectionId:  "runners",
			CascadeDelete: false,
			MaxSelect:     1,
		},
		&core.TextField{
			Name:     "logs",
			Required: false,
		},
	)

	collections := []*core.Collection{repositories, runners, jobs}

	for _, c := range collections {
		existing, _ := app.FindCollectionByNameOrId(c.Name)
		if existing != nil {
			// Collection exists, skip
			continue
		}

		if err := app.Save(c); err != nil {
			return err
		}
	}

	return nil
}
