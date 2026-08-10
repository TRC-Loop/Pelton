# Command palette

The command palette is one keyboard-driven surface for everything Pelton can do: every action, every mailbox, every saved view and every settings pane, all one fuzzy search away.

Press **Cmd+K** (macOS) or **Ctrl+K** (Windows and Linux) to open it. Press it again, or Escape, to close.

It opens as a bare input near the top of the window. Nothing is listed until you type, because a list of two hundred entries is not faster to read than the sidebar you already have.

## Typing

Type anything and the palette matches it against every command at once. Matching is fuzzy and works on initials, so you rarely need whole words:

| You type | You get |
| --- | --- |
| `arch` | Archive, plus every mailbox named Archive |
| `mu` | Mark unread |
| `ea` | Empty trash |
| `rn` | Rename mailbox |

Use the arrow keys to move, Enter to run, Escape to close. Matching letters are highlighted so it is obvious why something is in the list.

## Jumping to a result

Each of the first ten rows advertises its own key on the right: the top result runs on Enter, and the nine below it on Cmd+1 through Cmd+9 (Ctrl on Windows and Linux). You never have to arrow down to the fourth result, you just press Cmd+3.

If you would rather keep those combos for something else, turn off **Quick-select keys in the command palette** in **Settings, Shortcuts**. That hides the hints and frees the keys.

An action that already has its own keyboard shortcut shows it next to the quick-select key, so you can learn the permanent binding while using the temporary one.

## Narrowing with a prefix

Prefixes are optional. They are there when you already know what kind of thing you are after and do not want anything else in the way:

| Prefix | Limits results to |
| --- | --- |
| `>` | Actions |
| `@` | Mailboxes and views |
| `#` | Settings |
| `/` | Your mail |

So `@arch` finds only the Archive mailboxes, never the Archive action.

## Searching your mail

`/` searches the message index, the same one behind the search bar, so it understands the same typo-tolerant ranked matching.

Mail is the one thing a prefix is **required** for. An unprefixed query never touches it, so ordinary typing stays instant and local instead of running a query against your whole mailbox on every keystroke.

Results are whatever the index ranks highest, not fuzzy-matched again. Picking one opens the message and puts the query into the search bar behind the palette, so the list and the reading pane are showing the same thing.

## Picking a target

Some commands need to know what to act on. **Rename mailbox** has to know which mailbox; **Flag color** has to know which color. These do not guess. Choosing one keeps the palette open and switches it to a picker, with the command you chose shown as a chip in front of the input.

From a picker, Escape or Backspace on an empty input takes you back out.

Commands that work this way: new subfolder, rename mailbox, delete mailbox, pin or unpin mailbox, empty trash, flag color, apply theme and edit view.

## Acting on mail

Message actions only appear when there is a message to act on. With nothing open, Reply and Archive are simply not in the list rather than sitting there greyed out.

When you have several messages selected in the list, both forms appear and the count is in the label, so **Delete message** and **Delete 3 messages** can never be confused for each other. Bulk entries cover marking read and unread, flagging and unflagging, flag colors, archiving, downloading for offline and deleting.

## Learning what you use

The palette remembers which commands you run and how recently, and ranks them higher next time. If you archive constantly, `a` will start putting Archive first.

This is counted and stored locally, in the same settings database as the rest of your preferences. It never leaves your machine and it is not sent anywhere.

## Keyboard shortcuts for everything in it

Every action in the palette is also an action in **Settings, Shortcuts**, so anything you reach for often can get its own key instead. Actions that need a target, like Rename mailbox, open the palette on their own picker when you press their key.

The palette's own binding lives there too, so if Cmd+K is spoken for on your system you can move it.

## Changing what is in the menu bar

The palette and the in-app menu bar read the same list of actions, so an action you see in one is assignable in the other. **Settings, Menu bar** is where you add them to a menu.
