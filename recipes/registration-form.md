### registration-form
An accessible account-creation form with password guidance, terms acceptance, and validation hooks.

## When to use

Use this composition for creating a basic user account. Ask only for information required at registration time; defer profile details until they are useful.

## Example

```html
<main class="hero min-h-screen bg-base-200">
  <div class="hero-content w-full max-w-lg">
    <section class="card w-full bg-base-100 shadow-xl" aria-labelledby="register-title">
      <form class="card-body gap-4" novalidate>
        <div>
          <h1 id="register-title" class="card-title text-2xl">Create your account</h1>
          <p class="text-base-content/70">Already registered? <a class="link link-primary" href="/login">Sign in</a></p>
        </div>

        <div id="registration-errors" class="alert alert-error hidden" role="alert" tabindex="-1">
          Correct the highlighted fields and submit again.
        </div>

        <fieldset class="fieldset">
          <label class="label" for="register-name">Full name</label>
          <input id="register-name" name="name" class="input w-full" autocomplete="name" required>

          <label class="label" for="register-email">Email address</label>
          <input id="register-email" name="email" type="email" class="input w-full" autocomplete="email" required>

          <label class="label" for="register-password">Password</label>
          <input id="register-password" name="password" type="password" class="input w-full"
            autocomplete="new-password" minlength="12" aria-describedby="password-requirements" required>
          <p id="password-requirements" class="label">Use at least 12 characters.</p>
        </fieldset>

        <label class="label cursor-pointer justify-start gap-3">
          <input name="terms" type="checkbox" class="checkbox" required>
          <span>I agree to the <a class="link" href="/terms">terms of service</a>.</span>
        </label>

        <button class="btn btn-primary w-full" type="submit">Create account</button>
      </form>
    </section>
  </div>
</main>
```

## Accessibility notes

- Explain password requirements before validation and keep password-manager support enabled.
- Link field errors with `aria-describedby` and set `aria-invalid="true"` after validation fails.
- Focus the error summary after an unsuccessful submission.
- Verify email ownership after registration rather than requiring the email field twice.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/fieldset/, https://daisyui.com/components/input/, https://daisyui.com/components/checkbox/
- Adapted or copied code: No
