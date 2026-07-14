### login-form
A centered, accessible sign-in card with email, password, remember-me, and recovery actions.

## When to use

Use this composition for a dedicated sign-in page or authentication panel. Keep authentication errors close to the affected field and preserve the user's email value after a failed attempt.

## Example

```html
<main class="hero min-h-screen bg-base-200">
  <div class="hero-content w-full max-w-md">
    <section class="card w-full bg-base-100 shadow-xl" aria-labelledby="login-title">
      <div class="card-body">
        <h1 id="login-title" class="card-title text-2xl">Sign in</h1>
        <form class="space-y-4">
          <fieldset class="fieldset">
            <label class="label" for="email">Email address</label>
            <input id="email" name="email" type="email" autocomplete="email"
              class="input w-full" required aria-describedby="email-help">
            <p id="email-help" class="label">Use the address associated with your account.</p>
          </fieldset>

          <fieldset class="fieldset">
            <label class="label" for="password">Password</label>
            <input id="password" name="password" type="password" autocomplete="current-password"
              class="input w-full" required>
          </fieldset>

          <div class="flex items-center justify-between gap-4">
            <label class="label cursor-pointer gap-2">
              <input name="remember" type="checkbox" class="checkbox checkbox-sm">
              Remember me
            </label>
            <a class="link link-primary" href="/forgot-password">Forgot password?</a>
          </div>

          <button class="btn btn-primary w-full" type="submit">Sign in</button>
        </form>
      </div>
    </section>
  </div>
</main>
```

## Accessibility notes

- Use persistent labels; placeholders are not substitutes for labels.
- Put a failed-submission summary above the form and move focus to it.
- Add `aria-invalid="true"` and an `aria-describedby` reference to fields with errors.
- Do not disable password managers or paste.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/fieldset/, https://daisyui.com/components/input/, https://daisyui.com/components/card/
- Adapted or copied code: No
