package host

import (
	"context"
	featapi "k8s.io/api/features/v1alpha1"
	featapply "k8s.io/client-go/applyconfigurations/features/v1alpha1"
)

func (h *Host) syncFeature(ctx context.Context, name string) {
	local, ok := h.featureToProvider[name]
	if !ok {
		// This feature doesn't exist locally.
		// Figure out what version it is from.
		// If hosts at that version exist, it is their problem.
		// Otherwise, figure out how distant it is from our version.
		// If it is within a version or two, preserve it in case of roll back/forward.
		// Otherwise, delete it.
		// TODO: everything listed above.
		return
	}

	extExists := true
	extFeature, err := h.config.Informer.Lister().Get(h.config.Self.Class, name)
	if err != nil {
		// this doesn't exist in the cluster.
		extExists = false
	} else if name == "ClusterVersion" {
		// "ClusterVersion" is a special feature we use to track the
		// latched version of the cluster.
		h.observeFeatureVersion(extFeature.Status.Version)
	}

	s := h.getState()
	if !s.remote.leader {
		// We aren't the leader, we don't need to update stuff.
		return
	}

	// rules:
	// * when changingToVersion is non-empty, we need to call all the pre-version change functions
	//    .... we much be the prior version
	//    we mark ourselves unsafe to restart until all pre functions have returned without error?
	// * when changingToVersion is empty and the latchedVersion is updated, the post version change functions
	//    .... we must be the latchedVersion
	//

	ac := featapply.FeatureStatus()

	ls := local.Settings()
	ac.
		WithClass(ls.Class).
		WithName(ls.Name).
		WithDefault(interpret(ls.DefaultValue)).
		WithState(interpret(ls.ConfiguredValue)).
		WithVersion(h.config.Self.Version)

	if extExists {
		// By default, don't update the version, if one existed.
		ac.WithVersion(extFeature.Status.Version)
	}

	selfIsLatest := h.config.Self.Version == s.changingToVersion ||
		(s.changingToVersion == "" && h.config.Self.Version == s.cluster.latchedVersion)
	if !selfIsLatest {
		selfIsPrior := extExists && s.changingToVersion != "" && h.config.Self.Version == extFeature.Status.Version
		if selfIsPrior && extFeature.Status.Version == h.config.Self.Version {
			err := local.PreVersionChange(extFeature.Status.Version, s.changingToVersion)
			if err != nil {
			}
		}

	} else { // we are the latest version
		if extFeature.Status.Version != s.cluster.latchedVersion {
			err := local.PostVersionChange(extFeature.Status.Version, s.cluster.latchedVersion)
			if err != nil {
			}
			// on success, update version
			ac.WithVersion(s.cluster.latchedVersion)
		}
	}

	// that was the upgrade flow; we also need to handle the on/off flow
	if extExists && ls.Settability == Dynamic {
		if des := extFeature.Spec.Desired; des != nil {
			if *des != extFeature.Status.State {
				if *des == featapi.FeatureEnablementEnabled {
					if extFeature.Status.State == featapi.FeatureStateOff {
						// begin turn-on flow
						ac.WithState(featapi.FeatureStateTurningOn)
					}
				} else {
					if extFeature.Status.State == featapi.FeatureStateOn {
						// begin turn-off flow
						ac.WithState(featapi.FeatureStateTurningOff)
					}
				}
			}
		}
	}

	// Use SSA to update the status of this feature.
	h.config.Remote.Apply(featApply.Feature().WithStatus(ac), "featurecontroller")
}
