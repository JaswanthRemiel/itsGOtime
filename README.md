# itsGOtime

This template runs site checks in GitHub Actions and publishes a tiny dashboard to GitHub Pages.

How to use:
1. Edit `monitors.yaml` to list your targets.
2. Replace module path in `go.mod` if you want.
3. Create a repository on GitHub and push these files.
4. Run the `uptime-check` workflow from Actions (or wait for the cron).
5. After the workflow creates the `gh-pages` branch, go to Settings → Pages and set source to `gh-pages` branch (root).
6. Visit `https://<username>.github.io/<repo>/` to see the dashboard.

To test locally:
- Install Go 1.21+
- Run `go run ./cmd/checker`
- This writes `status.json` and `history.json` to repo root and tries to write `gh-pages/history.json` if `gh-pages` exists.
- Open `web/index.html` in a browser (or serve it with a simple static server).

Notes:
- The workflow commits to `gh-pages` only when changes exist.
- Default schedule is every minute. Adjust cron in `.github/workflows/check.yml` if you want lower frequency.
