### file-upload-form
An accessible document upload form with file constraints, optional notes, progress, and error feedback.

## When to use

Use this composition for one or a small number of files. For bulk uploads, add a queue with per-file status and cancellation controls.

## Example

```html
<section class="card mx-auto max-w-2xl bg-base-100 shadow" aria-labelledby="upload-title">
  <form class="card-body gap-5" enctype="multipart/form-data">
    <div><h1 id="upload-title" class="card-title text-2xl">Upload supporting document</h1><p class="text-base-content/70">PDF or image, up to 10 MB.</p></div>

    <div id="upload-error" class="alert alert-error hidden" role="alert" tabindex="-1">Select a supported file smaller than 10 MB.</div>

    <fieldset class="fieldset">
      <label class="label" for="document-file">Document</label>
      <input id="document-file" name="document" type="file" class="file-input w-full"
        accept="application/pdf,image/png,image/jpeg" aria-describedby="file-requirements" required>
      <p id="file-requirements" class="label">Accepted: PDF, PNG, or JPEG. Maximum size: 10 MB.</p>

      <label class="label" for="document-note">Description <span class="text-base-content/60">(optional)</span></label>
      <textarea id="document-note" name="description" class="textarea w-full" rows="3"></textarea>
    </fieldset>

    <div class="hidden" aria-live="polite">
      <div class="flex justify-between text-sm"><span>Uploading document.pdf</span><span>60%</span></div>
      <progress class="progress progress-primary w-full" value="60" max="100">60%</progress>
    </div>

    <div class="card-actions justify-end">
      <button class="btn" type="reset">Clear</button>
      <button class="btn btn-primary" type="submit">Upload document</button>
    </div>
  </form>
</section>
```

## Accessibility notes

- State accepted formats and limits in visible text before selection.
- Validate file content on the server; the `accept` attribute is only a picker hint.
- Announce meaningful progress changes without flooding the live region.
- Move focus to the error summary when submission fails.

## Provenance

- Maintainer: daisyui-mcp project
- Compatibility: daisyUI 5
- References: https://daisyui.com/components/file-input/, https://daisyui.com/components/progress/, https://daisyui.com/components/textarea/, https://daisyui.com/components/alert/
- Adapted or copied code: No
