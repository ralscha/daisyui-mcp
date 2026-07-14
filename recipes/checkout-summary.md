### checkout-summary
A responsive checkout step with contact fields, an order summary, and explicit total and payment action.

## When to use

Use this composition for a simple checkout. Complex payment flows should preserve entered information, clearly identify the current step, and provide a path back without losing work.

## Example

```html
<main class="bg-base-200 px-4 py-10">
  <div class="mx-auto max-w-6xl">
    <h1 class="mb-6 text-3xl font-bold">Checkout</h1>
    <ul class="steps mb-10 w-full" aria-label="Checkout progress">
      <li class="step step-primary">Cart</li><li class="step step-primary" aria-current="step">Details</li><li class="step">Payment</li>
    </ul>

    <div class="grid gap-6 lg:grid-cols-[1fr_24rem]">
      <form id="checkout-form" class="card bg-base-100 shadow">
        <div class="card-body gap-5">
          <h2 class="card-title">Contact and delivery</h2>
          <fieldset class="fieldset">
            <label class="label" for="checkout-email">Email address</label>
            <input id="checkout-email" name="email" type="email" autocomplete="email" class="input w-full" required>
            <label class="label" for="checkout-name">Full name</label>
            <input id="checkout-name" name="name" autocomplete="name" class="input w-full" required>
            <label class="label" for="checkout-country">Country or region</label>
            <select id="checkout-country" name="country" autocomplete="country-name" class="select w-full" required>
              <option value="">Choose a country</option><option>Switzerland</option><option>Germany</option><option>France</option>
            </select>
          </fieldset>
        </div>
      </form>

      <aside class="card h-fit bg-base-100 shadow" aria-labelledby="order-title">
        <div class="card-body">
          <h2 id="order-title" class="card-title">Order summary</h2>
          <dl class="space-y-3">
            <div class="flex justify-between gap-4"><dt>Professional plan</dt><dd>$29.00</dd></div>
            <div class="flex justify-between gap-4"><dt>Tax</dt><dd>$2.32</dd></div>
            <div class="divider my-1"></div>
            <div class="flex justify-between gap-4 text-lg font-bold"><dt>Total</dt><dd>$31.32 USD</dd></div>
          </dl>
          <button class="btn btn-primary mt-4 w-full" type="submit" form="checkout-form">Continue to payment</button>
          <p class="text-center text-xs text-base-content/60">You can review the order before payment.</p>
        </div>
      </aside>
    </div>
  </div>
</main>
```

## Accessibility notes

- Associate the current checkout step with `aria-current="step"` and do not rely on color alone.
- Put currency beside the total and disclose recurring billing before payment.
- Use autocomplete tokens for address and contact fields.
- Preserve form values when users return to an earlier step.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/steps/, https://daisyui.com/components/card/, https://daisyui.com/components/select/, https://daisyui.com/components/divider/
- Adapted or copied code: No
