# Upstream API Reference Mirror

This folder mirrors selected upstream files from:
https://github.com/jotform/jotform-api-go

Current mirrored file:
- v2/JotForm.go

Purpose:
- Keep a local, browsable API method list used for CLI planning.
- Avoid switching contexts while working in this repository.

Update command:

```bash
curl -fsSL https://raw.githubusercontent.com/jotform/jotform-api-go/master/v2/JotForm.go \
  -o docs/external/jotform-api-go/v2/JotForm.go
```

Notes:
- This is an upstream mirror; do not edit the mirrored file manually.
- Upstream ownership and licensing apply.
