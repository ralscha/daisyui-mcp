### newsletter-signup
A concise newsletter form with an explicit label, consent context, and success feedback.

## When to use

Use this composition near related content or in a footer. State what subscribers receive and how often before asking for an address.

## Example

```html
<section class="bg-primary px-4 py-12 text-primary-content" aria-labelledby="newsletter-title">
  <div class="mx-auto flex max-w-5xl flex-col items-start justify-between gap-6 lg:flex-row lg:items-center">
    <div class="max-w-xl">
      <h2 id="newsletter-title" class="text-3xl font-bold">Practical product updates</h2>
      <p class="mt-2 opacity-80">One useful email each month. Unsubscribe at any time.</p>
    </div>

    <form class="w-full max-w-xl" aria-describedby="newsletter-privacy">
      <label class="label text-primary-content" for="newsletter-email">Email address</label>
      <div class="join w-full">
        <input id="newsletter-email" name="email" type="email" autocomplete="email"
          class="input join-item w-full text-base-content" placeholder="you@example.com" required>
        <button class="btn join-item" type="submit">Subscribe</button>
      </div>
      <p id="newsletter-privacy" class="mt-2 text-sm opacity-75">We use your address only for this newsletter.</p>
      <p class="mt-2 hidden text-sm font-semibold" role="status">Check your inbox to confirm your subscription.</p>
    </form>
  </div>
</section>
```

## Accessibility notes

- Keep a visible label even when a descriptive placeholder is present.
- Announce submission results through a status region.
- Explain double opt-in and privacy handling in plain language.
- Stack the input and button when translated button text no longer fits comfortably.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/join/, https://daisyui.com/components/input/, https://daisyui.com/components/button/
- Adapted or copied code: No
