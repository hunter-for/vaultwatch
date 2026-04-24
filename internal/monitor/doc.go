// Package monitor provides lease-monitoring primitives for vaultwatch.
//
// Core types:
//
//	- Monitor   — fetches active leases from Vault and dispatches alerts.
//	- Scheduler — drives Monitor.CheckAll on a configurable polling interval.
//	- Severity  — classifies remaining TTL as Critical, Warning, or Info.
//	- Formatter — renders a human-readable alert message from a lease event.
//
// Typical usage:
//
//	mon := monitor.New(vaultClient, alertSender, monitor.DefaultThresholds())
//	sched := monitor.NewScheduler(mon, 60*time.Second)
//	if err := sched.Run(ctx); err != nil && err != context.Canceled {
//		log.Fatal(err)
//	}
package monitor
