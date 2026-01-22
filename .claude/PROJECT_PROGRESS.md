# Project Progress - sage

## Plan Files
Roadmap: None (v4 archived)
Current Phase: None
Latest Weekly Report: [weekly-2026-W03.md](../docs/reports/weekly-2026-W03.md)

Last Updated: 2026-01-22

## Current Focus
v0.5.0 released - Improved version detection and update workflow

## Active Tasks
None - awaiting direction

## Open Questions/Blockers
None

## Completed This Week
- Version detection and update improvements (v0.5.0)
  - Added runtime/debug.ReadBuildInfo() fallback for go install version detection
  - Implemented 'sage update' command for easy CLI updates
  - Tagged and released v0.5.0
- Debugged --json and --json --stream output
  - Identified flag ordering requirement (flags before positional args)
  - Confirmed both output modes work correctly
  - Documentation already accurate

## Next Session
Integrate sage with hub-core, or start next feature.
