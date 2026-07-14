### team-grid
A responsive team directory with names, roles, profile links, and resilient avatar fallbacks.

## When to use

Use this composition for a small team or working group. For a large organization, add search and filters or use a more compact directory table.

## Example

```html
<section class="bg-base-100 px-4 py-16" aria-labelledby="team-title">
  <div class="mx-auto max-w-6xl">
    <div class="mb-10 max-w-2xl">
      <h2 id="team-title" class="text-3xl font-bold">Meet the team</h2>
      <p class="mt-2 text-base-content/70">The people building and supporting the product.</p>
    </div>

    <ul class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3" role="list">
      <li class="card border border-base-300 bg-base-100">
        <div class="card-body items-center text-center">
          <div class="avatar avatar-placeholder">
            <div class="w-24 rounded-full bg-neutral text-neutral-content"><span class="text-2xl">AR</span></div>
          </div>
          <div><h3 class="card-title justify-center">Alex Rivera</h3><p class="text-base-content/70">Engineering</p></div>
          <div class="card-actions"><a class="btn btn-sm btn-ghost" href="/team/alex"><span class="sr-only">View </span>Profile</a></div>
        </div>
      </li>
      <li class="card border border-base-300 bg-base-100">
        <div class="card-body items-center text-center">
          <div class="avatar avatar-placeholder">
            <div class="w-24 rounded-full bg-secondary text-secondary-content"><span class="text-2xl">SK</span></div>
          </div>
          <div><h3 class="card-title justify-center">Sam Kim</h3><p class="text-base-content/70">Design</p></div>
          <div class="card-actions"><a class="btn btn-sm btn-ghost" href="/team/sam"><span class="sr-only">View </span>Profile</a></div>
        </div>
      </li>
      <li class="card border border-base-300 bg-base-100">
        <div class="card-body items-center text-center">
          <div class="avatar avatar-placeholder">
            <div class="w-24 rounded-full bg-accent text-accent-content"><span class="text-2xl">JM</span></div>
          </div>
          <div><h3 class="card-title justify-center">Jordan Morgan</h3><p class="text-base-content/70">Customer success</p></div>
          <div class="card-actions"><a class="btn btn-sm btn-ghost" href="/team/jordan"><span class="sr-only">View </span>Profile</a></div>
        </div>
      </li>
    </ul>
  </div>
</section>
```

## Accessibility notes

- Treat decorative headshots as decorative when the adjacent name identifies the person.
- Ensure initials have sufficient contrast and provide a fallback when images fail.
- Use a list to expose the collection structure.
- Avoid repeating identical accessible link names; include the person's name for assistive technology.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/avatar/, https://daisyui.com/components/card/, https://daisyui.com/components/button/
- Adapted or copied code: No
