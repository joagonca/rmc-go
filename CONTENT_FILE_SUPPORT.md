# Content File Support for Page Ordering

## Overview

The `--content` flag allows `rmc-go` to use reMarkable `.content` files for correct page ordering when processing folders of `.rm` files.

## Background

Individual `.rm` files do not contain page numbers or ordering information. Previously, `rmc-go` used file modification times to order pages, which becomes unreliable when pages are edited after creation. The correct page ordering is stored in reMarkable's `.content` JSON files.

## Implementation

### New Flag

```bash
--content string    Path to .content file for page ordering (only used with folders)
```

### Usage

```bash
# Use .content file for reliable page ordering
./rmc folder/ -o output.pdf --content folder.content

# Without .content file (falls back to modification time with warning)
./rmc folder/ -o output.pdf
```

### Behavior

1. **With folders + --content flag**:
   - Reads the `.content` JSON file
   - Extracts page IDs from `cPages.pages` array (in order)
   - Matches page IDs to `.rm` filenames (without extension)
   - Orders files according to the array order
   - Prints: `"Using page ordering from content file: <path>"`

2. **With folders, no --content flag**:
   - Falls back to modification time ordering
   - Prints warning: `"Warning: Using modification time for page ordering. For reliable ordering, use --content flag."`

3. **With folders + invalid/missing --content file**:
   - Falls back to modification time ordering
   - Prints warning: `"Warning: Could not use content file <path>, falling back to modification time ordering"`

4. **With single file + --content flag**:
   - Flag is silently ignored (no effect on single file processing)

### Content File Format

The `.content` file is a JSON file with this structure:

```json
{
  "cPages": {
    "pages": [
      {
        "id": "page-uuid-1",
        "idx": {
          "timestamp": "1:2",
          "value": "ba"
        },
        "modifed": "1730127536000"
      },
      {
        "id": "page-uuid-2",
        "idx": {
          "timestamp": "1:2",
          "value": "bb"
        },
        "modifed": "1730127526000"
      }
    ]
  },
  "pageCount": 2,
  "fileType": "notebook"
}
```

The `cPages.pages` array order is the authoritative page order. Each page's `id` field should match a `.rm` filename (without the `.rm` extension).

## File Structure

### New Files

- `internal/parser/content.go`: Content file parser
  - `ReadContentFile()`: Reads and parses `.content` JSON
  - `OrderFilesByContent()`: Orders `.rm` files based on content file

### Modified Files

- `cmd/rmc-go/main.go`:
  - Added `contentFile` flag variable
  - Updated `init()` to register `--content` flag
  - Modified `handleDirectory()` to use content file ordering
  - Added user-facing warnings for ordering methods

## Testing

Tested with:
- ✅ Folder with `--content` flag (uses content file)
- ✅ Folder without `--content` flag (uses modification time + warning)
- ✅ Folder with invalid `--content` path (falls back with warning)
- ✅ Single file with `--content` flag (flag ignored)
- ✅ Partial matches (matched pages ordered, unmatched pages appended)

## Fallback Behavior

When content file page IDs don't match all `.rm` files:
- Matched pages are placed first, in content file order
- Unmatched pages are appended at the end, sorted by modification time
- Processing continues successfully (no error)

## Future Enhancements

Potential improvements:
- Auto-detect `.content` files in the same directory
- Support for nested directories with UUID-based structure
- Extract page IDs from `.rm` file metadata (if available in future formats)
- Command-line flag to specify explicit page order (e.g., `--pages 2,1,3`)
