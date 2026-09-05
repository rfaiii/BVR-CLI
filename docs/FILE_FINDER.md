# BVR File Finder

BVR includes an in-app project file finder built as a native BVR dialog. It
keeps keyboard focus, mouse events, theming, and the agent session in one
Bubble Tea program instead of handing control to a separate application.

Open it from the `/` command palette by choosing **Open File Finder**, or type
`/finder` to filter directly to it. `ctrl+shift+f` is a quick toggle that opens
the finder and closes it when it is already open. The
finder provides:

- a project directory listing with mouse-clickable rows;
- a live text preview in the adjacent pane;
- selected-file metadata and clipboard status below the panes;
- path copying with `y` (and the clipboard pane reflects the copied path);
- context handoff with `a`, which inserts the selected file into the conversation
  editor and attaches its contents;
- external-editor handoff with `o`, which opens the selected file using the
  configured `$EDITOR`.
- **switch project** with `s`, which restarts BVR in the selected directory.
- **create new file** with `ctrl+n`, which opens the New File dialog.

Keys: `↑`/`↓` or `j`/`k` navigate, `pgup`/`pgdn` scroll by a page, `home`/`g`
go to the first entry, and `end` goes to the last entry. `enter`/`l` opens a
directory, `h`/`left` or backspace goes to the parent directory, `y` copies the
selected path, `.` toggles hidden dot-files, `a` attaches a file, `o` opens the
external editor, `s` switches project, `ctrl+n` creates a new file, and `esc` closes the finder.

The finder uses a responsive panel and reflows with terminal resizes. It keeps a
two-cell terminal inset, uses the available terminal height while open, and is
capped at 120 columns. Using the available height deliberately covers the large
application banner while the Finder is active, giving the list and preview more
vertical room without changing the global header/logo implementation. The block
scrollbar beside the file list shows the current position when more entries are
It can be opened by clicking the **OPEN FILE FINDER** button (folder icon) on
the landing screen or the compact folder icon in the header, or the
**CREATE FILE** button (paper icon) on the landing screen for new files.
Click **[X]** in the upper-right corner, press `esc`, or click anywhere
**outside the panel** to close it. The finder keeps its current directory
and selection when closed and reopened.

The selected file uses a light-pink highlight for visibility. The file list and
preview are separated by a divider, and the footer uses distinct colors for
metadata, clipboard state, and keyboard instructions. Work is tracked
in `docs/TASKLIST.md`.

## Layout and clipping contract

The Finder is intentionally rendered as several fixed regions rather than one
large ANSI/Lip Gloss row:

- The outer panel is calculated first, including its rounded border and padding.
  Content width and height are derived by subtracting that frame, so file
  contents cannot resize or push the panel outside the terminal.
- The title row contains the green `BVR FILE FINDER` label and the right-aligned
  `[X]` control. A grey pane-header rule is drawn immediately below it, making
  the top edge of both information panes visible.
- The body is split into a file-list pane, a one-cell center divider, and a
  preview/information pane. The pane-width calculation reserves a three-cell
  gutter before the outer right frame. This is intentional extra protection for
  terminals whose styled-string renderer behaves differently at a hard edge.
- The footer and lower divider have fixed rows. The list viewport reserves one
  row for the current directory, and the scrollbar reserves one track column.
  The scrollbar is drawn from the same body origin as the list, below the pane
  header rule, so clicks and scrolling remain aligned.
- Every list, preview, metadata, clipboard, and footer row is normalized and
  clipped by `fixedLines` before styling. Tabs become spaces and carriage returns
  are removed. Clipping is width-based and never wraps long unbroken tokens.
- The list and preview are then drawn independently into hard one-row
  rectangles. This prevents RTF/control-heavy text, Markdown, shell scripts,
  URLs, and other long tokens from reflowing across the divider, into adjacent
  rows, or through the outer right border.

These rules are the important regression boundary. Do not replace the separate
pane draws with a single `JoinHorizontal`/composed ANSI row unless the renderer
is also given an equivalent hard clipping implementation.

## Stabilization history

The Finder originally composed both panes into one styled row. RTF control
sequences, carriage returns, and very long first lines could then cause Lip Gloss
or the terminal renderer to reflow text and visually cross the center divider or
the panel edge. Stabilization proceeded as follows:

1. Fixed panel dimensions and calculated the content rectangle after border and
   padding frame sizes.
2. Added `fixedLines` normalization and clipping for arbitrary preview text.
3. Drew the list, divider, and preview independently inside hard rectangles.
4. Added a visible top pane rule and a right-side safety gutter.
5. Expanded the open Finder to the available height so the large banner no
   longer consumes the Finder's vertical workspace.
6. Corrected the scrollbar origin after adding the pane-header row and kept
   mouse hitboxes derived from the final panel geometry.

Regression tests cover exact-width rows, carriage-return normalization, RTF-like
text, unbroken lines, selection visibility, scrollbar movement, close-label
behavior, and narrow/normal pane-width calculations.

Processes and pinned tabs are intentionally omitted so the finder remains a
compact companion to the conversation UI.