// Package filesystem owns filesystem safety: path validation, change
// planning, conflict detection, dry-run, backup, commit, and best-effort
// rollback.
//
// Operations must be injectable so I/O failures can be simulated in tests.
package filesystem

// Plan describes the intended filesystem changes for a generation run.
type Plan struct {
	// Create holds paths that will be written for the first time.
	Create []string
	// Replace holds existing paths that will be overwritten.
	Replace []string
	// Conflicts holds paths that would be overwritten without consent.
	Conflicts []string
}

// FS is the filesystem operation interface used by the engine.
type FS interface {
	// PlanBuild computes the change plan for the given artifacts.
	PlanBuild(artifacts []string) (Plan, error)
	// Commit applies the plan. On failure it performs best-effort rollback.
	Commit(Plan) error
	// DryRun reports conflicts without writing anything.
	DryRun(Plan) (Plan, error)
}