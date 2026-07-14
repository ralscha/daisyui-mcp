### responsive-navigation
A responsive navbar with a keyboard-accessible mobile menu and visible desktop links.

## When to use

Use this composition for a small-to-medium site navigation. For larger information architectures, combine it with a drawer and clearly label the drawer trigger.

## Example

```html
<header class="navbar bg-base-100 shadow-sm" aria-label="Primary navigation">
  <div class="navbar-start">
    <details class="dropdown lg:hidden">
      <summary class="btn btn-ghost btn-square" aria-label="Open navigation menu">
        <span aria-hidden="true">☰</span>
      </summary>
      <ul class="menu dropdown-content z-10 mt-3 w-56 rounded-box bg-base-100 p-2 shadow">
        <li><a href="/products">Products</a></li>
        <li><a href="/pricing">Pricing</a></li>
        <li><a href="/docs">Documentation</a></li>
      </ul>
    </details>
    <a class="btn btn-ghost text-xl" href="/">Acme</a>
  </div>

  <nav class="navbar-center hidden lg:flex" aria-label="Main links">
    <ul class="menu menu-horizontal px-1">
      <li><a href="/products">Products</a></li>
      <li><a href="/pricing">Pricing</a></li>
      <li><a href="/docs">Documentation</a></li>
    </ul>
  </nav>

  <div class="navbar-end gap-2">
    <a class="btn btn-ghost" href="/login">Sign in</a>
    <a class="btn btn-primary" href="/signup">Get started</a>
  </div>
</header>
```

## Accessibility notes

- Keep link names and destinations consistent between mobile and desktop navigation.
- Indicate the current page with `aria-current="page"`.
- Prefer native `details`/`summary` or implement complete disclosure-button keyboard behavior.
- Ensure the mobile menu is not clipped by a parent with `overflow: hidden`.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/navbar/, https://daisyui.com/components/menu/, https://daisyui.com/components/dropdown/
- Adapted or copied code: No
