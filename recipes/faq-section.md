### faq-section
A native disclosure-based FAQ section that supports keyboard interaction without custom JavaScript.

## When to use

Use this composition for a short set of genuinely frequent questions. Keep critical policies and instructions in permanent documentation rather than hiding them only inside an FAQ.

## Example

```html
<section class="bg-base-100 px-4 py-16" aria-labelledby="faq-title">
  <div class="mx-auto max-w-3xl">
    <div class="mb-8 text-center">
      <h2 id="faq-title" class="text-3xl font-bold">Frequently asked questions</h2>
      <p class="mt-2 text-base-content/70">Can't find an answer? <a class="link link-primary" href="/support">Contact support</a>.</p>
    </div>

    <div class="space-y-3">
      <details class="collapse collapse-arrow border border-base-300 bg-base-100">
        <summary class="collapse-title font-semibold">Can I change plans later?</summary>
        <div class="collapse-content"><p>Yes. Upgrades apply immediately and downgrades begin at the next billing period.</p></div>
      </details>
      <details class="collapse collapse-arrow border border-base-300 bg-base-100">
        <summary class="collapse-title font-semibold">How does the trial work?</summary>
        <div class="collapse-content"><p>The trial lasts 14 days and does not require payment information.</p></div>
      </details>
      <details class="collapse collapse-arrow border border-base-300 bg-base-100">
        <summary class="collapse-title font-semibold">Can I export my data?</summary>
        <div class="collapse-content"><p>Workspace owners can export data from Settings at any time.</p></div>
      </details>
    </div>
  </div>
</section>
```

## Accessibility notes

- Prefer native `details` and `summary`, which provide keyboard and expanded-state behavior.
- Write descriptive question text so each disclosure makes sense out of context.
- Do not nest interactive controls inside `summary`.
- Allow multiple answers to remain open unless there is a strong reason not to.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/collapse/, https://daisyui.com/components/link/
- Adapted or copied code: No
