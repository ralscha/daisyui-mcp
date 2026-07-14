### application-shell
A responsive application shell with a persistent desktop sidebar and toggleable mobile drawer.

## When to use

Use this composition for dashboards and multi-page applications whose primary destinations remain stable across pages.

## Example

```html
<div class="drawer lg:drawer-open">
  <input id="app-navigation" type="checkbox" class="drawer-toggle">

  <div class="drawer-content flex min-h-screen flex-col bg-base-200">
    <header class="navbar bg-base-100 shadow-sm">
      <div class="navbar-start">
        <label for="app-navigation" class="btn btn-square btn-ghost lg:hidden" aria-label="Open navigation">
          <span aria-hidden="true">☰</span>
        </label>
        <span class="px-2 text-xl font-semibold">Operations</span>
      </div>
      <div class="navbar-end">
        <a class="btn btn-ghost" href="/profile">Account</a>
      </div>
    </header>

    <main id="main-content" class="flex-1 p-4 md:p-8">
      <h1 class="text-3xl font-bold">Overview</h1>
      <p class="mt-2 text-base-content/70">Application content belongs here.</p>
    </main>
  </div>

  <aside class="drawer-side" aria-label="Application navigation">
    <label for="app-navigation" class="drawer-overlay" aria-label="Close navigation"></label>
    <div class="min-h-full w-72 bg-base-100 p-4">
      <a class="btn btn-ghost mb-6 justify-start text-xl" href="/">Acme</a>
      <nav>
        <ul class="menu w-full gap-1">
          <li><a class="menu-active" href="/overview" aria-current="page">Overview</a></li>
          <li><a href="/projects">Projects</a></li>
          <li><a href="/team">Team</a></li>
          <li><a href="/settings">Settings</a></li>
        </ul>
      </nav>
    </div>
  </aside>
</div>
```

## Accessibility notes

- Include a skip link before the shell when navigation is repeated across pages.
- Keep the drawer checkbox label accessible and use `aria-current="page"` on the active destination.
- Move focus into the mobile drawer when it opens and return focus to its trigger when it closes.
- Do not hide the desktop sidebar from assistive technology when it is visually present.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/drawer/, https://daisyui.com/components/menu/, https://daisyui.com/components/navbar/
- Adapted or copied code: No
