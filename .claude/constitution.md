# Repository Constitution

## Article I: File Organization

**All files must be placed in their correct directory by type.**

### Section 1: Scripts
- Location: `scripts/`
- Content: Executable files only (`.sh`, `.py`)
- Prohibition: No documentation files (`.md`)

### Section 2: Documentation
- User docs: `docs/`
- Developer docs: `docs/building/`
- Prohibition: No mixing of user and developer documentation

### Section 3: Tests
- Location: `test/`
- Content: Test code, test scripts, fixtures
- Prohibition: No test files in `scripts/` or repository root

### Section 4: Examples
- Location: `examples/{example-name}/`
- Content: Self-contained working examples
- Prohibition: No scattered example files

## Article II: Repository Cleanliness

**The repository must remain organized at all times.**

### Section 1: Zero Tolerance
- No temporary files left in repository
- No files in wrong directories
- No duplicate files
- No orphaned scripts

### Section 2: Immediate Cleanup
When files are found in wrong locations:
1. Stop current work
2. Move files to correct locations
3. Update references
4. Verify structure
5. Resume work

## Article III: Development Discipline

**Think before you create.**

Before creating any file:
1. Determine correct location by type
2. Verify directory exists
3. Create file in correct location
4. Never create and move later

## Article IV: Enforcement

**This constitution is binding.**

- Every file creation must follow these rules
- Every commit must maintain organization
- No exceptions for "temporary" files
- Clean repository is a requirement, not a suggestion

## Rationale

A clean, organized repository:
- Improves discoverability
- Reduces confusion
- Enables automation
- Respects developer time
- Demonstrates professionalism

**Chaos is not acceptable.**
