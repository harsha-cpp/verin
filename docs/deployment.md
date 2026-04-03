# Deployment Notes

Deployment-specific configuration is intentionally not finalized in this MVP scaffold.

The current repo is shaped so the final deployment step can wire:

- API runtime target
- worker runtime target
- object storage vendor
- Redis hosting
- domain/callback URLs
- secret injection strategy
- monitoring vendor

The implementation should remain deployment-agnostic until those inputs are chosen explicitly.
