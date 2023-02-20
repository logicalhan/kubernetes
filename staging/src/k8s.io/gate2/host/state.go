package host

import (
	"crypto/sha256"
)

// remoteHostState collects the things we believe about our sibling hosts.
type remoteHostState struct {
	leader   bool     // true when we are the leader (for our version)
	versions []string // versions of hosts, including ours
}

// clusterState collects the things we believe about the cluster.
type clusterState struct {
	latchedVersion string // current cluster version (from published features)
}

type combinedState struct {
	changingToVersion string // non-empty when we are changing versions.

	sensible bool // true if no unexpected conditions encountered

	remote  remoteHostState
	cluster clusterState
}

func combine(hs remoteHostState, cs clusterState) combinedState {
	s := combinedState{
		remote:  newState,
		cluster: h.state.cluster,
	}

	if s.cluster.latchedVersion == "" {
		return s
	}

	count := 1
	other := ""
	for _, v := range s.remote.versions {
		if v == s.cluster.latchedVersion {
			continue
		}
		other = v
		count++
	}
	if count == 2 {
		s.changingToVersion = other
		s.sensible = true
	} else if count == 1 {
		s.sensible = true
	}
	return s
}

func (h *Host) getState() combinedState {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.state
}

func (h *Host) changeHostState(newState remoteHostState) {
	h.lock.Lock()
	defer h.lock.Unlock()

	s := combine(newState, h.state.cluster)

	h.stateChange_locked(s)
}

func (h *Host) changeClusterState(newState clusterState) {
	h.lock.Lock()
	defer h.lock.Unlock()

	s := combine(h.state.remote, newState)

	h.stateChange_locked(s)
}

func (h *Host) stateChange_locked(ns combinedState) {
	os := h.state
	h.state = ns

	if !ns.sensible {
		return
	}

	if ns.changingToVersion != "" && (ns.changingToVersion != os.changingToVersion || !os.sensible) {
		// A version change has been initiated.
		// TODO: trigger stuff.
	}
}

func (h *Host) summarizeRemotes(remotes []RemoteHost) remoteHostState {
	var out remoteHostState
	var foundUs bool
	myVersion := h.config.Self.Version
	versionToNames := map[string][]string{}

	for _, r := range remotes {
		if r.Class != h.config.Self.Class {
			continue
		}
		if r.Name == h.config.Self.Name {
			foundUs = true
		}
		versionToNames[r.Version] = append(versionToNames[r.Version], r.Name)
	}

	if !foundUs {
		versionToNames[myVersion] = append(versionToNames[myVersion], h.config.Self.Name)
	}

	for v := range versionToNames {
		out.versions = append(out.versions, v)
	}

	out.leader = deterministicLeader(versionToNames[myVersion]) == myVersion
	return out
}

// Given a set of unique names, this function computes a deterministic leader.
func deterministicLeader(names []string) string {
	if len(names) == 0 {
		return ""
	}
	if len(names) == 1 {
		return names[0]
	}
	least := -1
	hashes := []uint64{}
	h := sha256.New()
	for ii, n := range names {
		h.Reset()
		fmt.Fprint(h, n)
		s := h.Sum()
		var u uint64
		for i := 0; i < 8; i++ {
			u <<= 8
			u |= uint64(s[i])
		}
		hashes = append(hashes, u)
		if least == -1 || u < hashes[least] {
			least = ii
		}
	}
	return names[least]
}
