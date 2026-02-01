# ⚠️ SECURITY CRITICAL

This project contains REAL Tesla API credentials in `.env` and possibly OAuth tokens.

## Files that MUST NEVER be committed to git:
- ✅ **`.env`** - Protected by .gitignore ✓
- ✅ **`tesla-tokens.json`** - Protected by .gitignore ✓  
- ✅ **`config/`** (all files) - Protected by .gitignore ✓
- ✅ **`*.pem`** (private keys) - Protected by .gitignore ✓

## Before committing to GitHub:

```bash
# Verify nothing is accidentally staged
git status

# Verify .gitignore is working
git check-ignore -v .env tesla-tokens.json config/ *.pem
```

Should show these files are IGNORED (not tracked).

## If you accidentally committed secrets:

1. **Never push** - Delete the branch or revert
2. **Rotate credentials** - Get new CLIENT_ID/CLIENT_SECRET from developer.tesla.com
3. **Regenerate keys** - Run `bash setup-proxy.sh` to get new keys
4. **Don't remove history** - Use `git filter-branch` or GitHub's security tools

## How to safely contribute:

```bash
# Create fresh .env for your local copy only
cp .env.example .env
# Fill with YOUR credentials

# .env is automatically ignored
git status  # Should NOT show .env

# Safe to commit everything else
git add .
git commit -m "Your message"
```

## Verification

Before any git push, verify:
```bash
# These should NOT appear in commits
git log -p | grep -i "client_id\|client_secret\|token\|private"

# Should return nothing (empty)
```

---

✅ Project is set up correctly with `.gitignore` protecting secrets
⚠️ Just make sure you never manually add these files to git
