package test

const (
	finalizerName = "test.kkohtaka.org/finalizer"

	ownerLabelKey = "test.kkohtaka.org/owner"

	// If a child has updatingUntilAnnotation and the value specifies the time doesn't come yet, make child's phase
	// Updating.
	updatingUntilAnnotationKey = "test.kkohtaka.org/updating-until"
)
