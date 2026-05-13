# Security Policy

Please do not report security-sensitive issues in public GitHub Issues.

Until a dedicated security mailbox is published, use GitHub's private security
advisory workflow if available for the repository, or contact the maintainer
directly through the profile contact path.

When reporting, include:

- affected command or surface
- impact
- reproduction steps
- suggested mitigation if known

We will treat reports involving credential exposure, unsafe workflow execution,
or approval bypass as high priority.

Trusted-local escape hatches such as raw shell-backed routines should be
considered sensitive execution surfaces. Please report any case where Jini
executes repo-derived or home-derived command text more broadly than intended.
