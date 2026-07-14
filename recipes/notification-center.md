### notification-center
A notification dropdown with unread count, grouped events, clear actions, and a full-page fallback.

## When to use

Use this composition for a short preview of recent notifications. Keep the complete history on a dedicated page and avoid placing critical alerts only inside a dropdown.

## Example

```html
<details class="dropdown dropdown-end">
  <summary class="btn btn-circle btn-ghost" aria-label="Notifications, 3 unread">
    <span class="indicator">
      <span class="indicator-item badge badge-primary badge-xs">3</span>
      <span aria-hidden="true">♢</span>
    </span>
  </summary>

  <section class="dropdown-content z-20 mt-3 w-80 rounded-box bg-base-100 shadow-xl" aria-labelledby="notifications-title">
    <header class="flex items-center justify-between border-b border-base-300 p-4">
      <h2 id="notifications-title" class="font-bold">Notifications</h2>
      <button class="btn btn-xs btn-ghost" type="button">Mark all read</button>
    </header>

    <ul class="menu max-h-96 flex-nowrap overflow-y-auto p-2" role="list">
      <li>
        <a class="items-start gap-3 bg-primary/5" href="/deployments/482">
          <span class="status status-success mt-2" aria-label="Successful"></span>
          <span><strong class="block">Deployment complete</strong><span class="text-sm text-base-content/70">Production · 2 minutes ago</span></span>
        </a>
      </li>
      <li>
        <a class="items-start gap-3" href="/team/invitations">
          <span class="status status-info mt-2" aria-label="Information"></span>
          <span><strong class="block">New team invitation</strong><span class="text-sm text-base-content/70">Sam invited Jordan · 1 hour ago</span></span>
        </a>
      </li>
    </ul>

    <footer class="border-t border-base-300 p-2"><a class="btn btn-ghost btn-sm w-full" href="/notifications">View all notifications</a></footer>
  </section>
</details>
```

## Accessibility notes

- Include the unread count in the trigger's accessible name.
- Use text alongside status colors and mark newly arrived important notifications with a polite live region.
- Avoid automatically marking an item read merely because the menu opened.
- Ensure keyboard focus can move through the list and back to the trigger.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/dropdown/, https://daisyui.com/components/indicator/, https://daisyui.com/components/status/, https://daisyui.com/components/menu/
- Adapted or copied code: No
