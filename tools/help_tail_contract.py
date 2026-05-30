from __future__ import annotations

from typing import Sequence


HELP_TAIL_EXAMPLE_REQUEST = "me edit pear fellow script.txt"


def help_tail_redirect_lines(cli_name: str) -> tuple[str, str]:
    return (
        f"Start with `{cli_name}` and type the request in the shell.",
        f"Or use `{cli_name} open` or `{cli_name} status` if you want the current work surface.",
    )


def help_tail_request_text(request_tokens: Sequence[str]) -> str:
    return " ".join(str(token) for token in request_tokens).strip()


def help_tail_error_line(
    cli_name: str,
    help_invocation: str,
    inventory_name: str,
    request_tokens: Sequence[str],
) -> str:
    request = help_tail_request_text(request_tokens)
    return (
        f'ERROR `{cli_name} {help_invocation}` shows the {inventory_name}; '
        f'it does not take a request like "{request}".'
    )


def help_tail_message_lines(
    cli_name: str,
    help_invocation: str,
    inventory_name: str,
    request_tokens: Sequence[str],
) -> tuple[str, str, str]:
    error_line = help_tail_error_line(
        cli_name,
        help_invocation,
        inventory_name,
        request_tokens,
    )
    return (error_line, *help_tail_redirect_lines(cli_name))
