package calc

// GutterCalculator will handle gutter estimates.
// TODO: Implement gutter measurements and calculations
//
// Key measurements needed:
//   - Linear feet of gutter per run
//   - Number of downspouts
//   - Inside/outside corners
//   - End caps (left/right)
//   - Number of hangers (every 2-3 feet)
//   - Downspout elbows
//   - Extensions
//
// Materials to calculate:
//   - Gutter sections (by LF or 10'/20' sections)
//   - Downspout sections
//   - Inside/outside miters
//   - End caps
//   - Hangers/brackets
//   - Elbows (A-style, B-style)
//   - Extensions/splash blocks
//   - Sealant
//   - Rivets/screws

type GutterCalculator struct{}
