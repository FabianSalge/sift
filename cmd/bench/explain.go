package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/FabianSalge/sift/allocator"
)

// formatExplain renders one filter -> score -> bind trace as a text block: the
// workload's requirements, every device's verdict in fleet order (rejected with
// reasons, or feasible with its cost-weighted score and rank), and the bound
// winner with its true hourly price. It only formats; allocator.Explain decides.
func formatExplain(scenario string, w allocator.Workload, tr allocator.Trace, fleet []allocator.Device) string {
	price := make(map[string]float64, len(fleet))
	for _, d := range fleet {
		price[d.ID] = d.CostPerHr
	}
	bound := make(map[string]bool, len(tr.Bound))
	for _, id := range tr.Bound {
		bound[id] = true
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s  ·  %s · needs ≥%gGB, %s · %s · cost-weight %g\n",
		w.Name, w.Kind, w.MinMemoryGB, joinPrecisions(w.RequiredPrecisions), countDevices(w), w.CostWeight)
	fmt.Fprintf(&b, "%s · %d devices · filter → score → bind\n", scenario, len(tr.Verdicts))
	b.WriteString("score = $/hr × cost-weight, then memory waste, then ID (lower wins)\n\n")

	fmt.Fprintf(&b, "  %-4s  %-14s  %s\n", "rank", "device", "verdict")
	fmt.Fprintf(&b, "  %-4s  %-14s  %s\n", "----", "------", "-------")
	var feasible, rejected int
	for _, v := range tr.Verdicts {
		if v.Feasible {
			feasible++
		} else {
			rejected++
		}
		row := fmt.Sprintf("  %-4s  %-14s  %s", rankCell(v), v.DeviceID, verdictCell(v))
		if bound[v.DeviceID] {
			row += "   BIND"
		}
		b.WriteString(row + "\n")
	}

	b.WriteString("\n")
	if tr.Err != "" {
		fmt.Fprintf(&b, "  bound   (none)   %s\n", tr.Err)
		return b.String()
	}
	fmt.Fprintf(&b, "  bound   %s   $%.2f/hr   (lowest score of %d feasible; %d rejected)\n",
		strings.Join(tr.Bound, ", "), boundPrice(tr.Bound, price), feasible, rejected)
	return b.String()
}

// verdictCell renders one device's outcome: the failed hard constraints, or the
// soft-preference score the survivor was ranked by.
func verdictCell(v allocator.DeviceVerdict) string {
	if !v.Feasible {
		return "reject: " + rejectText(v.Reasons)
	}
	if v.Allocated {
		return "taken (bound by an earlier workload)"
	}
	return fmt.Sprintf("ok    cost %.2f   waste %gGB", v.Score.CostComponent, v.Score.MemoryWaste)
}

// rejectText abbreviates the hard constraints a device failed, reusing the
// trace's reasons (it formats them; it does not recompute feasibility).
func rejectText(rs []allocator.Reason) string {
	parts := make([]string, len(rs))
	for i, r := range rs {
		switch r.Code {
		case allocator.ReasonNotTrainable:
			parts[i] = "not trainable"
		case allocator.ReasonInsufficientMemory:
			parts[i] = strings.Replace(r.Detail, " < ", "<", 1) // "32GB < 80GB" -> "32GB<80GB"
		case allocator.ReasonMissingPrecision:
			parts[i] = strings.Replace(r.Detail, "missing ", "no ", 1) // "missing bf16" -> "no bf16"
		default:
			parts[i] = r.Detail
		}
	}
	return strings.Join(parts, ", ")
}

func rankCell(v allocator.DeviceVerdict) string {
	if v.Rank == 0 {
		return "—"
	}
	return strconv.Itoa(v.Rank)
}

func boundPrice(ids []string, price map[string]float64) float64 {
	var sum float64
	for _, id := range ids {
		sum += price[id]
	}
	return sum
}

func joinPrecisions(ps []allocator.Precision) string {
	s := make([]string, len(ps))
	for i, p := range ps {
		s[i] = string(p)
	}
	return strings.Join(s, "+")
}

func countDevices(w allocator.Workload) string {
	if w.DeviceCount <= 1 {
		return "1 device"
	}
	if w.SameIsland {
		return fmt.Sprintf("%d devices, same island", w.DeviceCount)
	}
	return fmt.Sprintf("%d devices", w.DeviceCount)
}
