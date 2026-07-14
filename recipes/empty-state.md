### empty-state
A focused empty state that explains why content is absent and offers a relevant next action.

## When to use

Use this composition when a collection has no records. Distinguish a first-use empty state from an empty search result or a permissions problem.

## Example

```html
<section class="hero min-h-96 rounded-box border border-dashed border-base-300 bg-base-100" aria-labelledby="empty-title">
  <div class="hero-content max-w-lg text-center">
    <div>
      <div class="mx-auto mb-5 grid size-16 place-items-center rounded-full bg-primary/10 text-3xl text-primary" aria-hidden="true">＋</div>
      <h2 id="empty-title" class="text-2xl font-bold">Create your first project</h2>
      <p class="py-4 text-base-content/70">Projects keep environments, deployments, and team access organized in one place.</p>
      <div class="flex flex-col justify-center gap-2 sm:flex-row">
        <a class="btn btn-primary" href="/projects/new">Create project</a>
        <a class="btn btn-ghost" href="/docs/projects">Read the project guide</a>
      </div>
    </div>
  </div>
</section>
```

## Empty search variation

Replace the creation action with “Clear filters” and repeat the search phrase in the heading, for example: `No projects match “archived API”`.

## Accessibility notes

- Make the heading explain the state, not merely say “Nothing here.”
- Offer one primary recovery action and avoid presenting unrelated choices.
- Decorative illustrations should have empty alternative text or be hidden from assistive technology.
- Use a status region if the empty state appears after a client-side search.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/hero/, https://daisyui.com/components/button/, https://daisyui.com/components/link/
- Adapted or copied code: No
