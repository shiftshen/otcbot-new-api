# OTCBot Fork Maintenance Plan

This fork powers `https://api.otcbot.com` and intentionally keeps customizations
small, explicit, and easy to rebase on top of upstream `QuantumNous/new-api`.

## Current Customizations

### Product and Branding

- Added an internal quick-start docs page at `/docs`.
- Replaced generic homepage/docs entry behavior so the site can default to its
  own documentation instead of `docs.newapi.pro`.
- Adjusted footer and default about-page content to point to OTCBot resources.

### Authentication

- Enabled Google login through the existing OIDC capability.
- Added Google-specific branding in login/register flows so the UI does not
  expose generic `OIDC` wording to end users.

### Payments

- Added XunhuPay compatibility on top of the original EPay-oriented flow.
- Normalized XunhuPay response parsing, including numeric `openid`.
- Added gateway-aware payment availability rules so unsupported methods can stay
  visible but disabled in the UI.
- Current production rule: `wxpay` is shown but disabled for XunhuPay because
  only Alipay is enabled on the merchant account.

### Deployment

- Added GitHub Actions deployment with a self-hosted runner on the OTCBot
  server.
- Production deployment target is the container serving `api.otcbot.com`.

## Compatibility Strategy

To keep this fork compatible with upstream updates, custom logic should follow
these rules:

1. Prefer additive wrappers over broad rewrites.
2. Keep OTCBot-specific rules in narrow helpers or dedicated files.
3. Avoid changing upstream behavior globally unless production requires it.
4. Preserve upstream naming and file layout where possible so rebases stay
   mechanical.
5. Document every production-only customization in this file when it lands.

## Safe Customization Boundaries

These areas are acceptable for fork-specific changes:

- `service/payment_gateway.go`
- payment-related controllers
- frontend branding components
- deployment scripts and GitHub workflows
- docs pages and user-facing guidance

These areas should stay as close to upstream as possible:

- model routing core
- channel scheduling and retry logic
- database schema unrelated to OTCBot features
- shared relay implementations unless a production bug requires a patch

## Upstream Sync Workflow

Use this process for every upstream upgrade:

1. Fetch upstream and create a temporary sync branch from `upstream/main`.
2. Rebase or merge OTCBot `main` onto the latest upstream.
3. Review conflicts by category:
   - payment customizations
   - branding/docs
   - deployment workflow
4. Run smoke verification:
   - Google login entry still renders correctly
   - Alipay payment link still opens successfully
   - `/docs` page is reachable
   - homepage docs button points to OTCBot docs
   - top-up page still disables WeChat when unsupported
5. Deploy to production only after the smoke checks pass.

Recommended git layout:

- `upstream/main`: official project tracking
- `main`: OTCBot production branch
- `codex/sync-YYYYMMDD`: temporary upstream integration branch

## Near-Term Roadmap

### Phase 1: Documentation and Product Clarity

- Expand `/docs` with more SDK examples: Python, Node.js, PHP, Java.
- Add a user-facing "Available Models" section sourced from configured models.
- Add a "How billing works" section for quota, pricing, and recharge flow.

### Phase 2: Payment Hardening

- Add explicit admin-side capability switches for Alipay and WeChat.
- Add structured error messages for gateway failures.
- Add callback verification smoke tests.

### Phase 3: Branding Cleanup

- Replace remaining upstream-facing marketing copy in homepage and settings.
- Audit admin pages for residual official links.
- Add configurable OTCBot-specific homepage blocks through settings.

### Phase 4: Upgrade Safety

- Add a lightweight regression checklist under `docs/`.
- Add targeted tests for payment gateway branching and docs-link resolution.
- Add a release checklist for server deploy, DB backup, and rollback.

## Release Checklist

Before each production release:

1. Confirm the branch contains only intended OTCBot changes.
2. Confirm `git log upstream/main..main` is understood and documented.
3. Verify login, recharge, `/docs`, and token creation manually.
4. Record the deployed commit hash in the release notes or deployment log.
5. Keep the previous image available for rollback.
