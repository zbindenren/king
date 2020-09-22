# king

Use as follows:

```go
func main() {
	v := info.NewGoReleaser(version, commit, date)
	cli := cmd.CLI{}
	app := kong.Parse(&cli, king.DefaultOptions(
		king.Config{
			Name:        "appname",
			Description: "A description",
			Version:     v.String(),
		},
	)...)
	defer cancel()
	if err := app.Run(&cli.Globals, l); err != nil {
		log.Fatal(err)
	}
}
```
