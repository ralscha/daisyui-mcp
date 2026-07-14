### filterable-data-table
A responsive records table with search, status filters, result count, and accessible pagination.

## When to use

Use this composition for data sets that benefit from scanning and comparison. On small screens, preserve horizontal scrolling or offer a purpose-built card view rather than hiding important columns.

## Example

```html
<section class="card bg-base-100 shadow" aria-labelledby="projects-title">
  <div class="card-body gap-5">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div><h2 id="projects-title" class="card-title">Projects</h2><p id="project-count" class="text-sm text-base-content/70">24 results</p></div>
      <a class="btn btn-primary btn-sm" href="/projects/new">New project</a>
    </div>

    <form class="flex flex-col gap-3 lg:flex-row lg:items-end" aria-describedby="project-count">
      <fieldset class="fieldset flex-1">
        <label class="label" for="project-search">Search projects</label>
        <input id="project-search" name="q" type="search" class="input w-full" placeholder="Name or owner">
      </fieldset>
      <fieldset class="fieldset">
        <legend class="fieldset-legend">Status</legend>
        <div class="filter">
          <input class="btn btn-sm btn-square" type="reset" value="×" aria-label="Clear status filter">
          <input class="btn btn-sm" type="radio" name="status" value="active" aria-label="Active">
          <input class="btn btn-sm" type="radio" name="status" value="paused" aria-label="Paused">
          <input class="btn btn-sm" type="radio" name="status" value="archived" aria-label="Archived">
        </div>
      </fieldset>
      <button class="btn btn-sm" type="submit">Apply</button>
    </form>

    <div class="overflow-x-auto">
      <table class="table">
        <caption class="sr-only">Projects matching the current search and status filters</caption>
        <thead><tr><th scope="col">Project</th><th scope="col">Owner</th><th scope="col">Status</th><th scope="col"><span class="sr-only">Actions</span></th></tr></thead>
        <tbody>
          <tr><td><a class="link font-medium" href="/projects/atlas">Atlas</a></td><td>Alex</td><td><span class="badge badge-success">Active</span></td><td><a class="btn btn-xs btn-ghost" href="/projects/atlas/settings">Settings<span class="sr-only"> for Atlas</span></a></td></tr>
          <tr><td><a class="link font-medium" href="/projects/orbit">Orbit</a></td><td>Sam</td><td><span class="badge badge-warning">Paused</span></td><td><a class="btn btn-xs btn-ghost" href="/projects/orbit/settings">Settings<span class="sr-only"> for Orbit</span></a></td></tr>
        </tbody>
      </table>
    </div>

    <nav class="join justify-center" aria-label="Project pages">
      <a class="btn join-item" href="?page=1" aria-label="Previous page">«</a>
      <a class="btn join-item btn-active" href="?page=2" aria-current="page">2</a>
      <a class="btn join-item" href="?page=3">3</a>
      <a class="btn join-item" href="?page=3" aria-label="Next page">»</a>
    </nav>
  </div>
</section>
```

## Accessibility notes

- Update and announce the result count after filters change.
- Preserve filter state in the URL when practical.
- Use table headers, a caption, and record-specific accessible action names.
- Apply `aria-current="page"` to the active pagination link.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/table/, https://daisyui.com/components/filter/, https://daisyui.com/components/pagination/, https://daisyui.com/components/badge/
- Adapted or copied code: No
