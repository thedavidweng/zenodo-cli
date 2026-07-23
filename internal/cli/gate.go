package cli

import (
	"github.com/thedavidweng/zenodo-cli/internal/model"
	"github.com/thedavidweng/zenodo-cli/internal/output"
)

type RiskTier int

const (
	RiskRead        RiskTier = iota // no mutation, always allowed
	RiskMediumWrite                 // blocked by --read-only, supports --dry-run
	RiskHighWrite                   // irreversible: blocked by --read-only, requires --confirm
)

type Plan struct {
	Action    string
	HumanMsg  string
	HumanArgs []any
	Data      map[string]any
}

type Gate struct {
	app  *AppContext
	r    *output.Renderer
	meta output.RuntimeMetaInput
}

func newGate(app *AppContext, r *output.Renderer, meta output.RuntimeMetaInput) *Gate {
	return &Gate{app: app, r: r, meta: meta}
}

// Allow returns (true, nil) when the command may proceed.
// Returns (false, err) when blocked by --read-only or missing --confirm.
// Returns (false, nil) when --dry-run is set; the plan has been rendered.
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
