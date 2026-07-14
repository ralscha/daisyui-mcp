### dashboard
A responsive dashboard shell with summary statistics, a compact activity table, and mobile-friendly overflow.

## When to use

Use this composition as a starting point for operational dashboards. Prioritize the few metrics that drive decisions; avoid presenting every available number at equal prominence.

## Example

```html
<main class="min-h-screen bg-base-200 p-4 md:p-8">
  <div class="mx-auto max-w-7xl space-y-6">
    <header class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold">Overview</h1>
        <p class="text-base-content/70">Updated a few seconds ago</p>
      </div>
      <button class="btn btn-primary" type="button">Create report</button>
    </header>

    <section class="stats stats-vertical w-full shadow lg:stats-horizontal" aria-label="Account summary">
      <div class="stat">
        <div class="stat-title">Revenue</div>
        <div class="stat-value">$42,800</div>
        <div class="stat-desc">12% above last month</div>
      </div>
      <div class="stat">
        <div class="stat-title">Active users</div>
        <div class="stat-value">8,420</div>
        <div class="stat-desc">320 joined this week</div>
      </div>
      <div class="stat">
        <div class="stat-title">Open incidents</div>
        <div class="stat-value text-warning">3</div>
        <div class="stat-desc">Oldest is 18 minutes</div>
      </div>
    </section>

    <section class="card bg-base-100 shadow" aria-labelledby="activity-title">
      <div class="card-body">
        <h2 id="activity-title" class="card-title">Recent activity</h2>
        <div class="overflow-x-auto">
          <table class="table">
            <caption class="sr-only">Most recent account events</caption>
            <thead><tr><th scope="col">Event</th><th scope="col">Owner</th><th scope="col">Status</th></tr></thead>
            <tbody>
              <tr><td>Production deployment</td><td>Sam</td><td><span class="badge badge-success">Complete</span></td></tr>
              <tr><td>Security review</td><td>Alex</td><td><span class="badge badge-warning">Waiting</span></td></tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>
  </div>
</main>
```

## Accessibility notes

- Do not communicate trends or status through color alone.
- Use real headings and table headers so screen-reader users can navigate the page.
- Add text alternatives for charts and expose the underlying data when possible.
- Announce automatic data refreshes only when the update is important.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/stat/, https://daisyui.com/components/table/, https://daisyui.com/components/card/
- Adapted or copied code: No
