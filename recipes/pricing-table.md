### pricing-table
A responsive three-tier pricing comparison with one clearly identified recommended plan.

## When to use

Use this composition when plans differ by a manageable set of features. For complex comparisons, pair the cards with a semantic comparison table below them.

## Example

```html
<section class="bg-base-200 px-4 py-16" aria-labelledby="pricing-title">
  <div class="mx-auto max-w-6xl">
    <div class="mx-auto mb-10 max-w-2xl text-center">
      <h2 id="pricing-title" class="text-3xl font-bold">Plans for every stage</h2>
      <p class="mt-2 text-base-content/70">All plans include a 14-day trial. No credit card required.</p>
    </div>

    <div class="grid gap-6 lg:grid-cols-3">
      <article class="card bg-base-100 shadow">
        <div class="card-body">
          <h3 class="card-title">Starter</h3>
          <p><span class="text-4xl font-bold">$0</span> / month</p>
          <ul class="my-4 space-y-2" aria-label="Starter features">
            <li>✓ One project</li><li>✓ Community support</li><li>✓ Basic analytics</li>
          </ul>
          <div class="card-actions mt-auto"><a class="btn w-full" href="/signup?plan=starter">Start free</a></div>
        </div>
      </article>

      <article class="card border-2 border-primary bg-base-100 shadow-xl" aria-labelledby="pro-plan">
        <div class="card-body">
          <div class="flex items-center justify-between gap-2">
            <h3 id="pro-plan" class="card-title">Professional</h3>
            <span class="badge badge-primary">Recommended</span>
          </div>
          <p><span class="text-4xl font-bold">$29</span> / month</p>
          <ul class="my-4 space-y-2" aria-label="Professional features">
            <li>✓ Unlimited projects</li><li>✓ Priority support</li><li>✓ Advanced analytics</li>
          </ul>
          <div class="card-actions mt-auto"><a class="btn btn-primary w-full" href="/signup?plan=pro">Start trial</a></div>
        </div>
      </article>

      <article class="card bg-base-100 shadow">
        <div class="card-body">
          <h3 class="card-title">Enterprise</h3>
          <p><span class="text-4xl font-bold">Custom</span></p>
          <ul class="my-4 space-y-2" aria-label="Enterprise features">
            <li>✓ Dedicated support</li><li>✓ Security reviews</li><li>✓ Custom agreements</li>
          </ul>
          <div class="card-actions mt-auto"><a class="btn w-full" href="/contact-sales">Contact sales</a></div>
        </div>
      </article>
    </div>
  </div>
</section>
```

## Accessibility notes

- Do not communicate the recommended plan through border color alone; include visible text.
- State billing periods and renewal behavior next to prices.
- Keep feature names consistent between cards so they are easy to compare.
- Avoid preselecting costly upgrades or hiding important limitations.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/card/, https://daisyui.com/components/badge/, https://daisyui.com/components/button/
- Adapted or copied code: No
