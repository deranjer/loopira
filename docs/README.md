# Loopira docs

The Loopira documentation site, built with [Docusaurus](https://docusaurus.io/).

## Local development

```bash
pnpm install
pnpm start
```

Starts a local dev server and opens a browser window. Most changes are
reflected live without a restart.

## Build

```bash
pnpm build
```

Generates static content into the `build` directory.

## Deployment

Pushes to `master` that touch `docs/**` are built and deployed to GitHub
Pages automatically by `.github/workflows/docs.yml`. There's no manual
deploy step.
