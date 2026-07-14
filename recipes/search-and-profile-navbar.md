### search-and-profile-navbar
A responsive application navbar with labeled search, notifications, and a native profile menu.

## When to use

Use this composition for signed-in applications where global search and account actions must remain readily available.

## Example

```html
<header class="navbar gap-2 bg-base-100 shadow-sm" aria-label="Application header">
  <div class="navbar-start min-w-fit">
    <a class="btn btn-ghost text-xl" href="/">Acme</a>
  </div>

  <div class="navbar-center flex-1">
    <form class="w-full max-w-xl" role="search">
      <label class="input flex w-full items-center gap-2">
        <span aria-hidden="true">⌕</span>
        <input name="q" type="search" class="grow" placeholder="Search projects"
          aria-label="Search projects">
      </label>
    </form>
  </div>

  <div class="navbar-end min-w-fit gap-1">
    <a class="btn btn-circle btn-ghost" href="/notifications" aria-label="Notifications, 3 unread">
      <span class="indicator"><span class="indicator-item badge badge-primary badge-xs">3</span><span aria-hidden="true">♢</span></span>
    </a>

    <details class="dropdown dropdown-end">
      <summary class="btn btn-circle btn-ghost" aria-label="Open account menu">
        <span class="avatar avatar-placeholder"><span class="w-9 rounded-full bg-neutral text-neutral-content">AR</span></span>
      </summary>
      <ul class="menu dropdown-content z-10 mt-3 w-56 rounded-box bg-base-100 p-2 shadow" aria-label="Account menu">
        <li><a href="/profile">Profile</a></li>
        <li><a href="/settings">Settings</a></li>
        <li><form action="/logout" method="post"><button type="submit">Sign out</button></form></li>
      </ul>
    </details>
  </div>
</header>
```

## Accessibility notes

- Give search an accessible name and use `type="search"`.
- Include unread counts in the notification link's accessible name.
- Prefer native `details` and `summary` for a simple account disclosure.
- Ensure narrow screens retain a usable brand, search trigger, and account action.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/navbar/, https://daisyui.com/components/dropdown/, https://daisyui.com/components/indicator/, https://daisyui.com/components/avatar/
- Adapted or copied code: No
