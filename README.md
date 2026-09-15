# Go Portfolio

Run the portfolio locally with Go 1.22 or newer:

```sh
go run .
```

Then open <http://127.0.0.1:8000/>. Set `PORT` to use a different port.

Projects with a `Repo` value in `main.go` are checked against that repository's public GitHub Pages site. When Pages is enabled, the portfolio automatically uses its URL and shows the project as **Live**. Results are cached for 10 minutes. Add the GitHub repository name to `Repo` when adding a project to make this work for future GitHub Pages deployments. Deployments on other hosts need a public URL supplied to the portfolio.

Run the automated checks with:

```sh
go test ./...
go vet ./...
```
