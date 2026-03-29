# Exit Codes

Exit codes indicate why things3-cli terminated and are stable across releases — use them in scripts, not error message text.

## Reference

| Code | Name     | Meaning                          |
|------|----------|----------------------------------|
| 0    | OK       | Success                          |
| 1    | Error    | General error                    |
| 2    | Auth     | Authentication failure (401/403) |
| 3    | NotFound | Resource not found (404)         |

## Constraints

- Do not parse stderr output to detect errors — use exit codes.
- Exit code 2 is returned for both 401 (missing token) and 403 (invalid token). Run `things3-cli status` to distinguish them.
- Exit code 3 is returned when a UUID is valid but the resource does not exist. It is not returned for empty list results.

## Examples

    # Check connectivity and auth before a script
    things3-cli status || exit 1

    # Conditional on not-found
    things3-cli get-todo "$ID"
    case $? in
      0) echo "found" ;;
      3) echo "not found" ;;
      *) echo "error" ;;
    esac
