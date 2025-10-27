# Code Optimization & Refactoring Plan

> **Last Updated:** 2025-10-27
> **Status:** ✅ Phase 1 & 2 Complete - All Critical and High Priority Issues Resolved

## Overview

This document tracks identified optimization opportunities, code quality improvements, and refactoring tasks for the rmc-go project. Issues are prioritized by severity and effort required.

---

## Summary Statistics

| Severity | Count | Status |
|----------|-------|--------|
| **Critical** | 1 | ✅ Complete |
| **High** | 6 | ✅ Complete |
| **Medium** | 15 | 🔄 In Progress |
| **Low** | 24 | ⏳ Pending |

---

## Implementation Phases

### ✅ Phase 1: Critical Fixes (COMPLETE)

**Goal:** Fix issues that could cause crashes or data loss

| Issue | File | Lines | Status |
|-------|------|-------|--------|
| Fix OOM in `Skip()` method | `internal/parser/limited_reader.go` | 42-55 | ✅ Done |
| Fix HTML escaping performance | `internal/export/svg.go` | 322-342 | ✅ Done |
| Fix error suppression in `drawText()` | `internal/export/svg.go` | 248-252 | ✅ Done |

### ✅ Phase 2: High-Value Quick Wins (COMPLETE)

**Goal:** Maximum impact with minimal effort

| Issue | File | Lines | Status |
|-------|------|-------|--------|
| Extract `NewEmptyGroup()` helper | `internal/parser/scene_stream.go` + `types.go` | Multiple | ✅ Done |
| Add constants for magic numbers | `internal/export/svg.go`, `internal/parser/scene_stream.go` | Multiple | ✅ Done |
| Extract generic `ReadLww()` helper | `internal/parser/block_reader.go` | 204-297 | ✅ Done |

### 🔄 Phase 3: Code Organization (PENDING)

**Goal:** Improve maintainability and readability

| Issue | File | Lines | Effort | Status |
|-------|------|-------|--------|--------|
| Break up `readLine()` function | `internal/parser/scene_stream.go` | 388-485 | Medium | ⏳ Todo |
| Break up `readRootTextBlock()` | `internal/parser/scene_stream.go` | 563-656 | Medium | ⏳ Todo |
| Refactor `BuildTextDocument()` | `internal/parser/text.go` | 22-106 | Medium | ⏳ Todo |
| Fix `buildAnchorPos()` style bug | `internal/export/svg.go` | 93-121 | Medium | ⏳ Todo |
| Add error context in loops | `internal/parser/scene_stream.go` | Multiple | Easy | ⏳ Todo |
| Improve error wrapping consistency | Multiple files | Multiple | Easy | ⏳ Todo |

### 🔧 Phase 4: Polish (PENDING)

**Goal:** Final touches for production quality

| Issue | File | Lines | Effort | Status |
|-------|------|-------|--------|--------|
| Replace `Printf` with logging | `internal/parser/scene_stream.go` | 69, 437 | Medium | ⏳ Todo |
| Standardize error messages | Multiple files | Multiple | Easy | ⏳ Todo |
| Improve PDF export error handling | `internal/export/pdf.go` | 43-50 | Medium | ⏳ Todo |
| Add input validation | `internal/parser/types.go` | Multiple | Easy | ⏳ Todo |
| Fix PDF temp file cleanup order | `internal/export/pdf.go` | 26-38 | Easy | ⏳ Todo |

---

## Detailed Findings

### 1. DRY VIOLATIONS (Repeated Code Patterns)

#### ✅ 1.1: Repeated Group Creation Pattern
**Location:** `internal/parser/scene_stream.go` - lines 135-142, 145-152, 282-289, 368-375

**Description:** The pattern for creating new Group objects is repeated 4 times with identical initialization.

**Severity:** Medium
**Effort:** Easy
**Status:** ✅ Complete

**Solution Implemented:**
```go
// Added to types.go
func NewEmptyGroup(id CrdtID) *Group {
    return &Group{
        NodeID:   id,
        Children: NewCrdtSequence(),
        Label:    LwwValue[string]{Timestamp: CrdtID{}, Value: ""},
        Visible:  LwwValue[bool]{Timestamp: CrdtID{}, Value: true},
    }
}
```

**Impact:** Eliminated ~50 lines of duplication

---

#### ✅ 1.2: Repeated LWW Value Reading Pattern
**Location:** `internal/parser/block_reader.go` - lines 204-297

**Description:** Five `ReadLww*` methods follow nearly identical patterns with only type differences.

**Severity:** High
**Effort:** Medium
**Status:** ✅ Complete

**Solution Implemented:**
```go
// Generic helper function
func readLww[T any](tbr *TaggedBlockReader, index uint8, readFn func(uint8) (T, error)) (LwwValue[T], error)
```

**Impact:** Reduced ~80 lines to ~20 lines, improved consistency

---

#### ⏳ 1.3: Repeated Error Handling Patterns
**Location:** `internal/parser/scene_stream.go` - Multiple locations

**Description:** Repetitive pattern for reading CRDT IDs with error handling.

**Severity:** Low
**Effort:** Medium
**Status:** ⏳ Deferred (idiomatic Go pattern)

---

#### ⏳ 1.4: Repeated String-to-RGB Color Conversion
**Location:** `internal/export/pen.go` - lines 133-160

**Description:** Color intensity calculations repeated with minor variations.

**Severity:** Low
**Effort:** Easy
**Status:** ⏳ Todo

---

### 2. ERROR HANDLING ISSUES

#### ⏳ 2.1: Inconsistent Error Wrapping
**Location:** Multiple files

**Description:** Some functions wrap errors with context, others return bare errors.

**Severity:** Medium
**Effort:** Easy
**Status:** ⏳ Todo

**Suggested Fix:**
```go
// Standardize to always include context
if err != nil {
    return fmt.Errorf("failed to read line: %w", err)
}
```

---

#### ✅ 2.2: Silent Error Suppression in drawText
**Location:** `internal/export/svg.go` - lines 248-252

**Description:** Error from `BuildTextDocument` was silently ignored.

**Severity:** High
**Effort:** Easy
**Status:** ✅ Complete

**Solution:** Changed signature to return error and propagate up the stack.

---

#### ⏳ 2.3: Silent Error Suppression in readLine
**Location:** `internal/parser/scene_stream.go` - lines 436-438, 460-474

**Description:** Extra bytes in subblock are read but error is discarded.

**Severity:** Medium
**Effort:** Easy
**Status:** ⏳ Todo

---

#### ⏳ 2.4: Insufficient Error Context in Batch Operations
**Location:** `internal/parser/scene_stream.go` - lines 563-598, 616-623

**Description:** Loop errors don't indicate which iteration failed.

**Severity:** Medium
**Effort:** Easy
**Status:** ⏳ Todo

**Suggested Fix:**
```go
for i := 0; i < int(numTextItems); i++ {
    item, err := readTextItem(reader)
    if err != nil {
        return fmt.Errorf("failed to read text item %d: %w", i, err)
    }
}
```

---

### 3. MAGIC NUMBERS

#### ✅ 3.1: Hardcoded Point Size Constants
**Location:** `internal/parser/scene_stream.go` - lines 416-419

**Severity:** Medium
**Effort:** Easy
**Status:** ✅ Complete

**Solution Implemented:**
```go
const (
    PointSizeV2 = 0x0E  // 14 bytes per point (version 2)
    PointSizeV1 = 0x18  // 24 bytes per point (version 1)
)
```

---

#### ⏳ 3.2: Hardcoded Threshold Values in readLine
**Location:** `internal/parser/scene_stream.go` - line 458

**Severity:** Low
**Effort:** Easy
**Status:** ⏳ Todo

**Suggested Fix:**
```go
const MinColorDataBytes = 6  // 2 prefix + 4 RGBA bytes
```

---

#### ⏳ 3.3: Hardcoded Font and Layout Values
**Location:** `internal/export/svg.go` - lines 19-27

**Severity:** Low
**Effort:** Easy
**Status:** ⏳ Todo

**Suggested Fix:** Add comments explaining design system values.

---

#### ⏳ 3.4: Multiple Hardcoded Color Values
**Location:** `internal/parser/types.go` + `internal/export/pen.go`

**Severity:** Medium
**Effort:** Medium
**Status:** ⏳ Todo

**Note:** Consider single source of truth for color definitions.

---

#### ✅ 3.5: Hardcoded Special Anchor IDs
**Location:** `internal/export/svg.go` - lines 70-71

**Severity:** High
**Effort:** Easy
**Status:** ✅ Complete

**Solution Implemented:**
```go
const (
    SpecialAnchorID1 = 281474976710654  // 2^48 - 2
    SpecialAnchorID2 = 281474976710655  // 2^48 - 1
    SpecialAnchorYPos = 100.0
)
```

---

#### ✅ 3.6: Hardcoded Screen Dimensions and DPI
**Location:** `internal/export/svg.go` - lines 12-15

**Severity:** Medium
**Effort:** Easy
**Status:** ✅ Complete

**Note:** Added comments documenting reMarkable device specifications.

---

#### ⏳ 3.7: Hardcoded Pen Parameters
**Location:** `internal/export/pen.go` - lines 68-86, 102-105

**Severity:** Medium
**Effort:** Easy
**Status:** ⏳ Todo

---

### 4. STRING BUILDING INEFFICIENCIES

#### ✅ 4.1: Inefficient HTML Escaping
**Location:** `internal/export/svg.go` - lines 322-342

**Description:** O(n²) string concatenation in loop.

**Severity:** High
**Effort:** Easy
**Status:** ✅ Complete

**Solution Implemented:** Used standard library `html.EscapeString()`

**Impact:** O(n²) → O(n) complexity, more correct escaping

---

### 5. PERFORMANCE ISSUES

#### ✅ 5.1: Potential Unbounded Memory Allocation in Skip()
**Location:** `internal/parser/limited_reader.go` - lines 42-55

**Description:** Allocates entire remaining buffer into memory (could be gigabytes).

**Severity:** Critical
**Effort:** Easy
**Status:** ✅ Complete

**Solution Implemented:** Chunked reading with 8KB buffer.

**Impact:** Prevents OOM crashes on large files.

---

#### ⏳ 5.2: Inefficient Peekable TaggedBlockReader
**Location:** `internal/parser/block_reader.go` - lines 131-154

**Severity:** Medium
**Effort:** Medium
**Status:** ⏳ Todo

---

#### ⏳ 5.3: N+1 Pattern in buildAnchorPos
**Location:** `internal/export/svg.go` - lines 93-121

**Description:** Style lookup issue - doesn't use actual character styles.

**Severity:** Medium
**Effort:** Medium
**Status:** ⏳ Todo (Phase 3)

---

### 6. LIBRARY OPPORTUNITIES

#### ✅ 6.1: HTML Escaping
**Location:** `internal/export/svg.go` - lines 322-342

**Severity:** Medium
**Effort:** Easy
**Status:** ✅ Complete

**Solution:** Used `html.EscapeString()` from standard library.

---

#### ⏳ 6.2: Manual VarUint Parsing
**Location:** `internal/parser/datastream.go` - lines 96-116

**Severity:** Low
**Effort:** Hard
**Status:** ⏳ Deferred (current implementation is correct and clear)

---

### 7. CODE ORGANIZATION

#### ⏳ 7.1: Large readLine Function (97 lines)
**Location:** `internal/parser/scene_stream.go` - lines 388-485

**Severity:** Medium
**Effort:** Medium
**Status:** ⏳ Todo (Phase 3)

**Suggested Refactoring:**
- `readLineMetadata()` - tool, color, thickness
- `readLinePoints()` - point data
- `parseColorOverride()` - optional color data

---

#### ⏳ 7.2: Large readRootTextBlock Function (93 lines)
**Location:** `internal/parser/scene_stream.go` - lines 563-656

**Severity:** Medium
**Effort:** Medium
**Status:** ⏳ Todo (Phase 3)

**Suggested Refactoring:**
- `readTextItems()` - CRDT sequence items
- `readTextFormatting()` - style map
- `readTextPosition()` - position and width

---

#### ⏳ 7.3: Large BuildTextDocument Function (85 lines)
**Location:** `internal/parser/text.go` - lines 22-106

**Severity:** Low
**Effort:** Medium
**Status:** ⏳ Todo (Phase 3)

---

### 8. UNUSED CODE & VARIABLES

#### ⏳ 8.1: Unused _isAscii Flag
**Location:** `internal/parser/datastream.go` - line 145

**Severity:** Low
**Effort:** Easy
**Status:** ⏳ Todo

**Action:** Document why this is read but unused.

---

#### ⏳ 8.2: Unused _itemType Variables
**Location:** `internal/parser/scene_stream.go` - lines 265, 352

**Severity:** Low
**Effort:** Easy
**Status:** ⏳ Todo

**Action:** Add validation with meaningful errors.

---

### 9. BUGS & CORRECTNESS ISSUES

#### ⏳ 9.1: buildAnchorPos Style Lookup Bug
**Location:** `internal/export/svg.go` - lines 93-121

**Description:** Code assigns `currentStyle` but always uses `StylePlain` for line height calculation.

**Severity:** Medium
**Effort:** Medium
**Status:** ⏳ Todo (Phase 3)

**Fix Required:**
```go
// Look up actual style for this character
if styleValue, exists := text.Styles[charID]; exists {
    currentStyle = styleValue.Value
}
lineHeight := lineHeights[currentStyle]  // Use actual style
```

---

#### ⏳ 9.2: PDF Temporary File Cleanup Risk
**Location:** `internal/export/pdf.go` - lines 26-38

**Severity:** Medium
**Effort:** Easy
**Status:** ⏳ Todo (Phase 4)

---

### 10. INCONSISTENCIES

#### ⏳ 10.1: Printf Debug Statements
**Location:** `internal/parser/scene_stream.go` - lines 69, 437

**Severity:** Medium
**Effort:** Medium
**Status:** ⏳ Todo (Phase 4)

**Action:** Use proper logging package or debug flag.

---

#### ⏳ 10.2: Inconsistent Error Message Format
**Location:** Multiple files

**Severity:** Low
**Effort:** Easy
**Status:** ⏳ Todo (Phase 4)

**Action:** Standardize to lowercase without trailing period.

---

## Files by Issue Count

| File | Critical | High | Medium | Low | Total |
|------|----------|------|--------|-----|-------|
| `internal/parser/scene_stream.go` | 0 | 1 | 5 | 5 | 11 |
| `internal/export/svg.go` | 0 | 2 | 3 | 2 | 7 |
| `internal/parser/block_reader.go` | 0 | 1 | 2 | 1 | 4 |
| `internal/parser/limited_reader.go` | 1 | 0 | 0 | 0 | 1 |
| `internal/export/pen.go` | 0 | 0 | 1 | 1 | 2 |
| `internal/export/pdf.go` | 0 | 1 | 1 | 0 | 2 |
| `internal/parser/datastream.go` | 0 | 0 | 1 | 1 | 2 |
| `internal/parser/types.go` | 0 | 0 | 1 | 0 | 1 |
| `internal/parser/text.go` | 0 | 0 | 1 | 0 | 1 |

---

## Next Steps

### Immediate (Phase 3)
1. Fix `buildAnchorPos()` style lookup bug
2. Add error context in batch operations
3. Break up large functions for better maintainability
4. Improve error wrapping consistency

### Future (Phase 4)
1. Replace Printf with proper logging
2. Standardize error message format
3. Improve PDF export error handling
4. Add input validation throughout
5. Fix temporary file cleanup order

---

## Notes

- **Test Coverage:** Consider adding unit tests as functions are refactored
- **Performance:** Profile with real-world files before/after optimizations
- **Breaking Changes:** None of these changes should affect the public API
- **Documentation:** Update godoc comments as code is refactored

---

*This document will be updated as optimizations are completed.*
