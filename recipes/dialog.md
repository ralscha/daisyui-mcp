### dialog
A native modal dialog for a destructive confirmation, including safe focus and cancellation behavior.

## When to use

Use a modal only when the user must make a decision before continuing. Prefer an inline confirmation for low-risk or easily reversible actions.

## Example

```html
<button class="btn btn-error" type="button"
  onclick="document.getElementById('delete-dialog').showModal()">
  Delete project
</button>

<dialog id="delete-dialog" class="modal" aria-labelledby="delete-title" aria-describedby="delete-description">
  <div class="modal-box">
    <h2 id="delete-title" class="text-lg font-bold">Delete this project?</h2>
    <p id="delete-description" class="py-4">
      This permanently removes the project and its deployment history.
    </p>

    <div class="modal-action">
      <form method="dialog">
        <button class="btn" value="cancel">Cancel</button>
      </form>
      <form action="/projects/123/delete" method="post">
        <button class="btn btn-error" type="submit">Delete project</button>
      </form>
    </div>
  </div>

  <form method="dialog" class="modal-backdrop">
    <button aria-label="Close dialog">close</button>
  </form>
</dialog>
```

## Accessibility notes

- Open native dialogs with `showModal()` so focus is trapped and the background becomes inert.
- Give focus to the least destructive sensible action when the dialog opens.
- Keep Escape and the explicit Cancel button available.
- Return focus to the trigger after the dialog closes.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/modal/, https://daisyui.com/components/button/
- Adapted or copied code: No
