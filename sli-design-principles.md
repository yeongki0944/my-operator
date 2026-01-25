### SLI Instrumentation Policy (Design Principles)

This project includes an experimental SLI instrumentation framework intended only for development-time design validation, not for production monitoring.

The following rules are intentional design decisions, not limitations.

#### 1. No operator code modification

This project does not modify operator (controller) source code to insert SLI instrumentation.

- The controller’s Reconcile logic is never touched.  
- No additional metrics are registered inside the controller process.  
- No flags or build-time options are required in the operator code.  

This allows the same SLI instrumentation to be applied to any operator without forking or patching its implementation.

#### 2. **/metrics** is an input source, not an output target

The controller’s /metrics endpoint is used only as a source of raw signals, such as:

- controller_runtime_reconcile_total  
- controller_runtime_reconcile_time_seconds  
- other controller-runtime–provided metrics  

Derived SLI metrics such as:  
- e2e_convergence_time_seconds  
- reconcile_total_delta  

are **not exposed** on the controller’s **/metrics** endpoint.  

This is intentional.  

Because the operator code is not modified, derived SLI metrics cannot and should not be registered in the controller’s Prometheus registry.

#### 3. SLI computation happens in E2E tests

All SLI measurements are computed externally, in E2E test code:

- Start time is recorded when the test creates a primary Custom Resource.  
- End time is recorded when the resource is observed as Ready.  
- Reconcile churn is calculated by reading controller metrics at test start and end.  

This keeps instrumentation:

- independent of operator implementation details  
- resilient to refactoring  
- easy to remove without leaving technical debt

#### 4. Default OFF, explicit opt-in  

SLI instrumentation is disabled by default.  

It is enabled only when explicitly requested (for example via environment variables in E2E runs).

If instrumentation is disabled:  

- no measurements are recorded  
- no metrics are scraped  
- test behavior is unchanged  

#### 5. Measurement failure does not fail tests

SLI instrumentation is strictly observational.

Any failure during measurement, including:

- **/metrics** scrape failure  
- metric not found  
- parsing errors  
- timeouts  
- write failures for summary artifacts  

does **not** cause test failure.

Instead:

- the SLI result is marked as skip  
- the test result remains authoritative  

This ensures that experimental instrumentation never destabilizes CI or development workflows.

#### 6. Experiment artifacts

Each E2E run produces, at most, two artifacts:

- **/metrics** (raw controller metrics, scrapeable during the run)  
- sli-summary.json (a per-run summary of derived SLI values)  

These artifacts exist to support design analysis and comparison between runs.
They are not part of a production SLO pipeline.

### Optional Operator-Attached Instrumentation (Non-default)

While this project is intentionally designed to operate without any operator code modification, it does not forbid all forms of operator-attached instrumentation.

Instead, it defines a strict separation between:  
- Default mode: external, test-driven SLI observation (described above)  
- Optional mode: explicitly opt-in, adapter-based attachment for advanced use cases    

This coexistence is intentional.  

#### Default remains non-invasive  

The rules described above are the default and recommended mode of operation:

- No controller code changes
- No derived SLI metrics exposed from /metrics
- All computation performed in E2E tests
- Zero impact on operator behavior

Any deviation from this model is opt-in and non-default.

#### Adapter-based attachment (opt-in only)

In some scenarios (e.g., long-running development clusters, extended trend analysis, or deeper debugging), it may be desirable to attach instrumentation more closely to an operator.

When this is done, it must follow these rules:  
- Instrumentation is implemented as a separate adapter layer, not inside pkg/slo  
- The core operator logic (Reconcile, state transitions, error handling) remains untouched  
- Instrumentation uses only existing extension points (e.g., controller-runtime manager options, HTTP handlers, or metrics registries)  
- Instrumentation can be enabled or disabled without rebuilding or forking the operator  
- Failure or absence of instrumentation must never affect reconciliation behavior  

This ensures that:  
- pkg/slo remains a pure, reusable library  
- Operator code does not become coupled to experimental SLI logic  
- The instrumentation remains removable without leaving technical debt  

#### Non-goals of attached instrumentation

Even in adapter-based mode, the following are explicitly out of scope:  
- Injecting SLI logic directly into Reconcile  
- Making reconciliation success depend on instrumentation  
- Introducing high-cardinality labels tied to runtime identifiers  
- Turning this framework into a production SLO monitoring system  

The purpose of any attached instrumentation is still design validation, not production observability.

Summary

This SLI framework is designed to answer one question during development:

“How quietly and predictably does this operator converge?”

It does so by default through external, test-driven observation, deliberately avoiding operator intrusion.  

At the same time, it allows carefully constrained, opt-in attachment through adapter layers, enabling deeper experimentation without compromising:  
- operator independence  
- reuse across projects  
- long-term maintainability  

Real-time observability is intentionally traded for:
- zero required operator changes  
- controlled experimental scope  
- confidence that instrumentation can always be removed  