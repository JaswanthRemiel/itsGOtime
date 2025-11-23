# itsGOtime
![Website](https://img.shields.io/website?url=http%3A%2F%2Fgithub.remiel.work%2FitsGOtime%2F&up_message=online&label=demo)


https://github.com/user-attachments/assets/da5a87e6-fd5e-4096-a572-f3a8806e82a1


This template runs site checks in GitHub Actions and publishes a Next.js dashboard to GitHub Pages.

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
