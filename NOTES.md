## Usage
config priority:
- cli
- ENV
- config

if missing:
prompt

"edit" style menus, use bubbletea for an "editor"

```bash
plextrac clients
plextrac --client DEMO reports
plextrac --client DEMO reports new --title "Internal whatever 2025" --template
plextrac --client DEMO --report '*2024*' sections
plextrac --client DEMO --report '*2024*' sections add --title "Narrative: External"
plextrac --client DEMO --report '*2024*' findings
plextrac --client DEMO --report '*2024*' --finding '*LM*' assets
plextrac --client DEMO --report '*2024*' --finding '*LM*' assets add "example.local"
grep aadc secretsdump.ntds | plextrac --client DEMO --report '*2024*' --finding '*LM*' assets add
```

## Linting
- Check for "n't"
- Check for "he" since we're all males
- Check that each picture has a caption
