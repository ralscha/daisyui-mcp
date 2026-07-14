### forgot-password
A compact password-recovery request with privacy-preserving confirmation feedback.

## When to use

Use this composition when a user cannot access an account. Return the same confirmation whether or not the address exists to avoid account enumeration.

## Example

```html
<main class="hero min-h-screen bg-base-200">
  <div class="hero-content w-full max-w-md">
    <section class="card w-full bg-base-100 shadow-xl" aria-labelledby="recovery-title">
      <form class="card-body gap-5">
        <div>
          <h1 id="recovery-title" class="card-title text-2xl">Reset your password</h1>
          <p id="recovery-help" class="text-base-content/70">
            Enter your account email and we'll send recovery instructions if an account matches.
          </p>
        </div>

        <fieldset class="fieldset">
          <label class="label" for="recovery-email">Email address</label>
          <input id="recovery-email" name="email" type="email" autocomplete="email"
            class="input w-full" aria-describedby="recovery-help" required>
        </fieldset>

        <button class="btn btn-primary w-full" type="submit">Send recovery email</button>
        <a class="btn btn-ghost w-full" href="/login">Return to sign in</a>
      </form>
    </section>

    <div class="alert alert-success mt-4 hidden" role="status">
      If an account matches that address, recovery instructions are on their way.
    </div>
  </div>
</main>
```

## Accessibility notes

- Use a `status` region for the confirmation so it is announced without moving focus unnecessarily.
- Keep the response wording identical for registered and unregistered addresses.
- Recovery links should expire, be single-use, and lead to a form that supports password managers.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/card/, https://daisyui.com/components/fieldset/, https://daisyui.com/components/alert/
- Adapted or copied code: No
