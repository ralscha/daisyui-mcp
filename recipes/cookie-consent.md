### cookie-consent
A compact cookie-consent notice with equally available accept, reject, and preference actions.

## When to use

Use this composition only when consent is legally or operationally required. Do not block essential site functionality or load optional trackers before a valid choice.

## Example

```html
<aside class="fixed inset-x-4 bottom-4 z-50 mx-auto max-w-4xl" aria-labelledby="cookie-title">
  <div class="alert items-start bg-base-100 shadow-2xl sm:items-center">
    <div class="min-w-0 flex-1">
      <h2 id="cookie-title" class="font-bold">Your privacy choices</h2>
      <p class="text-sm text-base-content/70">
        We use necessary cookies to operate the site. With permission, we also use analytics to improve it.
        <a class="link" href="/privacy">Read our privacy notice</a>.
      </p>
    </div>
    <div class="flex w-full flex-wrap gap-2 sm:w-auto sm:justify-end">
      <button class="btn btn-sm" type="button">Reject optional</button>
      <button class="btn btn-sm btn-ghost" type="button">Preferences</button>
      <button class="btn btn-sm btn-primary" type="button">Accept optional</button>
    </div>
  </div>
</aside>
```

## Accessibility notes

- Give accept and reject actions comparable visual prominence and effort.
- Do not use a modal unless local requirements truly demand one.
- Save the choice and provide a persistent way to reopen preferences.
- Avoid auto-focusing the banner unless it blocks interaction with the page.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/alert/, https://daisyui.com/components/button/, https://daisyui.com/components/link/
- Adapted or copied code: No
