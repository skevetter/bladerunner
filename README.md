# Bladerunner

Bladerunner is a GitHub Runner orchestration and management service.

GitHub runners support using a single self-hosted runner for multiple repositories for organizations, but not for personal accounts. You cannot register GitHub runners at the account level for personal accounts the same way you can for organizations. For personal accounts, you need to register a runner for each repository. Also, GitHub runners does not provide a single dashboard to view all runner jobs and track metrics.

There are a few projects that attempt to solve this problem, but the implementations are not complete. The `actions-runner-controller`, requires using Kubernetes and you still need to manually register a runner for each repository.  `garm` seems unnecessarily complex. Ideally, jobs for each repository could be registered to a single runner and processed via a queue.

Bladerunner solves this problem by using a single (or scalable) self-hosted runner capable of processing jobs for multiple repositories for personal accounts. It registers a runner for each repository and processes jobs via a queue. It also provides visibility into the jobs for all your repositories through a unified interface.

## References

- [GitHub Self-Hosted Runners](https://docs.github.com/en/actions/hosting-your-own-runners)
- [actions-runner-controller](https://github.com/actions/actions-runner-controller)
- [garm](https://github.com/cloudbase/garm)
- [runner](https://github.com/actions/runner)
- [multi-gh-action-runner](https://github.com/gershnik/multi-gh-action-runner)