package calc

// SidingCalculator will handle siding estimates.
// TODO: Implement siding measurements and calculations
//
// Key measurements needed:
//   - Wall area per elevation (sq ft)
//   - Window/door openings to deduct
//   - Gable end areas
//   - Inside/outside corner count
//   - J-channel linear feet
//   - Starter strip linear feet
//   - Soffit/fascia measurements
//
// Materials to calculate:
//   - Siding panels (by sq ft coverage)
//   - J-channel
//   - Inside/outside corner posts
//   - Starter strip
//   - Utility trim
//   - Soffit panels
//   - Fascia covers
//   - Nails/fasteners

type SidingCalculator struct{}
