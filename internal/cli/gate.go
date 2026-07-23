package cli

import (
	"github.com/thedavidweng/zenodo-cli/internal/model"
	"github.com/thedavidweng/zenodo-cli/internal/output"
)

// RiskTier classifies a command's mutation risk, matching the three tiers
// documented in docs/ARCHITECTURE.md and CONTEXT.md's 安全门控 concept.
type RiskTier int

const (
	// RiskRead is a no-mutation command: always allowed, never blocked by
	// --read-only or --confirm. May still support --dry-run for commands
	// with local side effects (e.g. files download).
	RiskRead RiskTier = iota
	// RiskMediumWrite is a mutation blocked by --read-only and supports
	// --dry-run. Does not require --confirm.
	RiskMediumWrite
	// RiskHighWrite is an irreversible mutation: blocked by --read-only,
	// requires --confirm, and supports --dry-run.
	RiskHighWrite
)

// Plan describes what a dry-run would do. The gate emits both the human
// message and the JSON plan envelope when --dry-run is set.
type Plan struct {
	Action    string         // optional, included in JSON "action" field when non-empty
	HumanMsg  string         // format string for human dry-run output (e.g. "Would delete %s\n")
	HumanArgs []any          // arguments for HumanMsg
	Data      map[string]any // extra fields merged into the JSON plan envelope
}

// Gate enforces the safety policy (安全门控) for a command. It owns the
// three-tier decision: read-only blocking, confirm requirement, and dry-run
// plan emission. Commands declare their tier and plan; the gate decides.
type Gate struct {
	app  *AppContext
	r    *output.Renderer
	meta output.RuntimeMetaInput
}

func newGate(app *AppContext, r *output.Renderer, meta output.RuntimeMetaInput) *Gate {
	return &Gate{app: app, r: r, meta: meta}
}

// Allow checks the safety policy for the given tier.
//
// Returns (true, nil) when the command may proceed with the real operation.
// Returns (false, err) when blocked by --read-only or missing --confirm;
// the error is already rendered to the output.
// Returns (false, nil) when --dry-run is set; the plan has been rendered
// and the caller should return immediately.
func (g *Gate) Allow(tier RiskTier, plan Plan) (bool, error) {
	if tier >= RiskMediumWrite {
		if g.app.ReadOnly {
			return false, g.r.Failure(g.meta, output.Errorf(model.ErrReadOnlyViolation, "--read-only blocks this mutation"))
		}
		if tier == RiskHighWrite && !g.app.Confirm {
			return false, g.r.Failure(g.meta, output.Errorf(model.ErrConfirmationRequired, "use --confirm to proceed"))
		}
	}

	if g.app.DryRun {
		g.r.Human(plan.HumanMsg, plan.HumanArgs...)
		data := map[string]any{"planned": true}
		if plan.Action != "" {
			data["action"] = plan.Action
		}
		for k, v := range plan.Data {
			data[k] = v
		}
		return false, g.r.Success(g.meta, data, nil)
	}

	return true, nil
}
