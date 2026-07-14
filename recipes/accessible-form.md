### accessible-form
An accessible profile form with grouped fields, help text, validation hooks, and a clear submission action.

## When to use

Use this pattern for settings and data-entry forms. Break long forms into meaningful sections and ask only for information needed at the current step.

## Example

```html
<form class="card mx-auto max-w-2xl bg-base-100 shadow" novalidate>
  <div class="card-body gap-6">
    <div>
      <h1 class="card-title text-2xl">Profile settings</h1>
      <p id="form-help" class="text-base-content/70">Fields marked required must be completed.</p>
    </div>

    <div id="form-errors" class="alert alert-error hidden" role="alert" tabindex="-1">
      <span>Correct the highlighted fields and submit again.</span>
    </div>

    <fieldset class="fieldset">
      <legend class="fieldset-legend">Contact details</legend>

      <label class="label" for="full-name">Full name</label>
      <input id="full-name" name="full_name" class="input w-full" autocomplete="name" required>

      <label class="label" for="contact-email">Email address</label>
      <input id="contact-email" name="email" type="email" class="input w-full"
        autocomplete="email" required aria-describedby="email-description">
      <p id="email-description" class="label">We'll send account notifications to this address.</p>
    </fieldset>

    <fieldset class="fieldset">
      <legend class="fieldset-legend">Notification preference</legend>
      <label class="label cursor-pointer justify-start gap-3">
        <input class="radio" type="radio" name="notifications" value="important" checked>
        Important updates only
      </label>
      <label class="label cursor-pointer justify-start gap-3">
        <input class="radio" type="radio" name="notifications" value="all">
        All product updates
      </label>
    </fieldset>

    <div class="card-actions justify-end">
      <button class="btn btn-primary" type="submit">Save changes</button>
    </div>
  </div>
</form>
```

## Accessibility notes

- Associate every control with a visible label and group related controls with `fieldset` and `legend`.
- On failure, reveal the error summary, focus it, and link each error to its field.
- Set `aria-invalid` only after validation fails; remove it when corrected.
- Keep the submit button enabled so users can discover validation errors.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/fieldset/, https://daisyui.com/components/input/, https://daisyui.com/components/radio/
- Adapted or copied code: No
