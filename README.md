# Problem this library tries to solve
When building packages of software automatically, it is useful to also generate incrementing version numbers automatically.
With GitLab-Ci, you can configure multiple build actions and test runs that run with every commit. Each such automatic
build gets an automatic build number. This build number is not ideal to use as part of a version string for various reasons,
the main motivation for me was this one:

* a version string should be the same for all configured steps for each commit -- I don't want a gap in increments just because
a few test ran in between

This library connects to a specified GitLab server via its REST api, looks for the specified project's builds, and essentially
counts the number of commits for which there are builds associated. The proposed version is then that number + 1.
