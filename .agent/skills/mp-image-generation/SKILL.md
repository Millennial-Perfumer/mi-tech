---
name: mp-image-generation
description: Create approval-gated Millennial Perfumer product images from a reference and product image, then write matching social captions and hashtags after approval. Use for product creative, campaign imagery, image revisions, or Instagram-ready content for Millennial Perfumer.
---

# Millennial Perfumer Content Workflow

Create a product image that borrows the reference image's composition, mood, lighting, camera angle, and prop language while faithfully preserving the supplied Millennial Perfumer product. The user remains the final creative approver.

Read [the brand profile](references/brand-profile.md) before generating or editing an image.

## Inputs and setup

Ask for any missing required input before generation:

- a reference/inspiration image;
- the product image with the label clearly visible;
- the product name or SKU, if it is not legible; and
- any campaign-specific brief (optional).

For each job, create a non-destructive directory under `content-jobs/` using a date plus product identifier. Keep the inputs, prompt, generated variants, approved result, caption, and hashtags together. Do not overwrite an existing approved image.

## Image workflow

1. Inspect both images. State the reference traits that will be carried over and the product details that must remain invariant.
2. Write a concise, structured image-generation prompt. Treat the reference as style/composition guidance only; do not reproduce third-party branding, logos, text, or identifiable people from it.
3. Generate one high-quality product-image variant with the image-generation tool. Use the product image as the product-fidelity reference.
4. Perform a visual QA check before presenting it. Check product silhouette, bottle and cap finish, label spelling/layout, legibility, lighting, reflections, perspective, composition, artifacts, and absence of unwanted text or watermarks.
5. Present the image with a compact QA summary and stop for an explicit user decision. Never produce captions or hashtags before approval.

The possible decisions are:

- **Approve**: mark the selected variant approved and proceed to social copy.
- **Fix: _feedback_**: retain all unmentioned approved aspects, make only the requested targeted change, create a new numbered variant, QA it, and return to approval.
- **Regenerate**: make a genuinely new variation from the same brief, QA it, and return to approval.

Do not infer approval from positive feedback such as “nice” or “looks good”; require a clear approval.

## Social copy after approval

Analyze the approved image, not the prompt or a previous variant. Produce:

- three distinct Instagram caption options, each with a strong first line, sensory/product-led body, and a natural call to action; and
- exactly five copy-ready hashtags chosen for relevant organic discovery among fragrance audiences.

Use a balanced mix of broad and niche/contextual fragrance tags. Do not include a branded hashtag (for example, `#MillennialPerfumer`) unless the user explicitly asks for one. Do not promise reach, virality, or a specific audience size. Avoid irrelevant trending, spammy, or repetitive generic tags. Keep claims truthful to the visible product and confirmed product facts. Save the proposed copy as `caption.txt` and the five tags as `hashtags.txt` in the job directory.

## Delivery

At completion, report the approved image path, the job directory, and the captions/hashtags. Preserve prior variants so the user can revisit a creative decision later.
