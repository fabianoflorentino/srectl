---
title: "Learn SRE, DevOps and Platform Engineering on a real cluster"
description: "A local platform on kind — GitOps, LGTM observability, security, demo apps and chaos — taught through hands-on scenarios."
---

`srectl` stands up a complete platform on **kind**: GitOps with Argo CD, the LGTM
observability stack, security (Kyverno, Trivy, Falco), demo services and chaos drills —
and turns each theme into a hands-on lesson.

```bash
srectl up            # cluster + Argo CD + every app, via GitOps
srectl status        # health matrix
srectl demo load     # generate traffic, light up the dashboards
```

Follow a learning path below, or jump straight to the **Tutorials**.
