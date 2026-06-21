package dra

import (
	"sort"
	"strconv"
	"strings"

	"github.com/FabianSalge/sift/allocator"
)

// Request is the neutral DRA shape of a Workload's HARD constraints: a CEL device
// selector plus the device count and an optional cross-device match-attribute
// (same-island gangs). It carries no k8s/CEL types, so the core stays lean; the
// driver renders it into a ResourceClaim. Soft cost-weighted scoring is
// deliberately absent — CEL filters, it does not optimize (ADR-0018); that stays
// in the allocator/benchmark.
type Request struct {
	Expression     string // AND-ed CEL device selector; "true" when unconstrained
	Count          int    // devices to bind; max(1, Workload.DeviceCount)
	MatchAttribute string // qualified attribute all bound devices must share, "" if none
}

// WorkloadSelector derives the DRA request for a workload, mirroring
// allocator.Feasible clause-for-clause. domain is the attribute domain (the
// driver name) under which Describe's bare attribute names are published.
func WorkloadSelector(w allocator.Workload, domain string) Request {
	n := w.DeviceCount
	if n < 1 {
		n = 1
	}
	r := Request{Expression: selectorExpr(w, domain), Count: n}
	if w.SameIsland && n > 1 {
		r.MatchAttribute = domain + "/island"
	}
	return r
}

// selectorExpr builds the AND-ed CEL predicate mirroring Feasible: trainability
// (training jobs), memory floor, and required-precision membership. Clauses are
// emitted in a fixed order (precisions sorted) so the result is deterministic.
func selectorExpr(w allocator.Workload, domain string) string {
	attrs := "device.attributes['" + domain + "']"
	var clauses []string
	if w.Kind == allocator.KindTrain {
		clauses = append(clauses, attrs+".trainable")
	}
	if w.MinMemoryGB > 0 {
		gi := strconv.FormatFloat(w.MinMemoryGB, 'f', -1, 64) + "Gi"
		clauses = append(clauses, "device.capacity['"+domain+"'].memory.compareTo(quantity('"+gi+"')) >= 0")
	}
	precisions := make([]string, len(w.RequiredPrecisions))
	for i, p := range w.RequiredPrecisions {
		precisions[i] = string(p)
	}
	sort.Strings(precisions)
	for _, p := range precisions {
		clauses = append(clauses, "'precision_"+p+"' in "+attrs)
	}
	if len(clauses) == 0 {
		return "true"
	}
	return strings.Join(clauses, " && ")
}
