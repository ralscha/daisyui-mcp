### landing-page-hero
A responsive product hero with a focused value proposition, primary action, secondary action, and product preview.

## When to use

Use this composition at the top of a product or campaign landing page. Keep the headline specific and make the primary action match the next step promised by the copy.

## Example

```html
<section class="hero min-h-[70vh] bg-base-200" aria-labelledby="hero-title">
  <div class="hero-content flex-col gap-10 py-16 lg:flex-row-reverse">
    <figure class="w-full max-w-xl">
      <div class="mockup-browser border border-base-300 bg-base-100 shadow-2xl">
        <div class="mockup-browser-toolbar"><div class="input">https://app.example.com</div></div>
        <div class="grid min-h-64 place-items-center bg-base-200 p-8">
          <div class="stats stats-vertical bg-base-100 shadow sm:stats-horizontal">
            <div class="stat"><div class="stat-title">Projects</div><div class="stat-value">128</div></div>
            <div class="stat"><div class="stat-title">Uptime</div><div class="stat-value">99.9%</div></div>
          </div>
        </div>
      </div>
      <figcaption class="sr-only">Product dashboard showing project count and service uptime.</figcaption>
    </figure>

    <div class="max-w-xl text-center lg:text-left">
      <span class="badge badge-primary badge-outline mb-4">Now available</span>
      <h1 id="hero-title" class="text-5xl font-bold tracking-tight">Ship reliable products with less operational work</h1>
      <p class="py-6 text-lg text-base-content/70">
        Monitor deployments, coordinate incidents, and understand product health from one workspace.
      </p>
      <div class="flex flex-col justify-center gap-3 sm:flex-row lg:justify-start">
        <a class="btn btn-primary" href="/signup">Start free</a>
        <a class="btn btn-ghost" href="/demo">Watch the demo</a>
      </div>
      <p class="mt-3 text-sm text-base-content/60">14-day trial · No credit card required</p>
    </div>
  </div>
</section>
```

## Accessibility notes

- Keep one `h1` that states the page purpose rather than repeating marketing slogans.
- Give product images meaningful alternative text or a nearby text description.
- Make both actions explicit; avoid ambiguous labels such as “Learn more” when a clearer destination exists.
- Respect reduced-motion preferences if the preview is animated.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/hero/, https://daisyui.com/components/mockup-browser/, https://daisyui.com/components/stat/
- Adapted or copied code: No
