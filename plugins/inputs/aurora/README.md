# Aurora Input Plugin

The Aurora plugin gathers metrics about Apache Aurora, Mesos framework. Examples of measurements and tags below.

### Configuration:

```toml
# Reads metrics from one or more Apache Aurora framework hosts.
[[inputs.aurora]]
  ## aurora servers
  # hosts = ["http://localhost:8081"]
  # hosts = ["https://localhost:8081"]
```

### Measurements & Fields:


- aurora
    - uncaught_exceptions
    - update_transition
    - timed_out_tasks
- aurora:cron
    - events_per_sec
    - events
    - nanos_total
    - events_per_sec
    - nanos_per_event
- aurora:http:responses
    - events (integer, count)
    - nanos_total (integer, count)
    - nanos_per_event
- aurora:job:update
    - recovery_errors
    - delete_errors
    - state_change_errors
- aurora:jvm
    - class_unloaded_count
    - gc_ps_marksweep_collection_count
    - class_total_loaded_count
    - uptime
    - gc_collection_count
    - memory_mb_total
    - memory_heap_mb_max
    - class_loaded_count
    - memory_max_mb
    - gc_ps_scavenge_collection_count
    - memory_non_heap_mb_committed
- aurora:jvm:gc
    - ps_marksweep_collection_time_ms
    - collection_time_ms
    - ps_scavenge_collection_time_ms
- aurora:jvm:memory
    - heap_mb_committed
    - heap_mb_used
    - free_mb
    - non_heap_mb_used
- aurora:jvm:threads
    - active (integer, count)
    - peak (integer, count)
    - started
    - daemon
- aurora:log:storage:write
    - lock_wait_ns_total_per_sec
    - lock_wait_ns_per_event
    - lock_wait_events
    - lock_wait_events_per_sec
    - lock_wait_ns_total
- aurora:process
    - max_fd_count
    - open_fd_count
- aurora:preemptor
    - task_processor_runs
    - missing_attributes
- aurora:preemptor:slot
    - search_attempts_for_non_prod
    - tasks_preempted_non_prod
    - search_successful_for_prod
    - tasks_preempted_prod
    - search_failed_for_non_prod
    - search_attempts_for_prod
    - validation_failed
    - validation_successful
    - search_failed_for_prod
    - search_successful_for_non_prod
- aurora:resources
    - empty_slots
    - cpu_cores
    - disk_mb
    - ram_mb
- aurora:schedule
    - attempts_fired
    - lifecycle
    - attempts_failed
    - attempts_no_match
- aurora:scheduler
    - driver_kill_failures (integer, count)
    - process_cpu_cores_utilized
    - timeout_queue_size
    - framework_registered
    - outstanding_offers
    - backup_failed
    - resource_offers
    - backup_success
- aurora:scheduler:log
    - native_append_nanos_per_event (integer, ns)
    - native_read_failures
    - native_truncate_nanos_total_per_sec
    - native_append_timeouts
    - native_read_events
    - bytes_read
    - un_snapshotted_transactions
    - native_read_nanos_total
    - deflated_entries_read
    - bytes_written
    - native_read_nanos_total_per_sec
    - entries_read
    - snapshots
    - native_truncate_nanos_total
    - bad_frames_read
    - native_read_timeouts
    - native_append_events_per_sec
    - native_append_nanos_total
    - native_truncate_events
    - native_native_entries_skipped
    - native_append_nanos_total_per_sec
    - native_read_nanos_per_event
    - native_truncate_nanos_per_event
    - native_append_events
    - native_truncate_timeouts
    - native_read_events_per_sec
    - entries_written
    - native_append_failures
    - native_truncate_events_per_sec
    - native_truncate_failures
- aurora:scheduler:tasks
    - async
    - throttle_events_per_sec
- aurora:scheduler:storage
    - events_per_sec
    - events
    - nanos_total
    - events_per_sec
    - nanos_per_event
- aurora:scheduler:thrift
    - (<func>\w+)_nanos_per_event
    - (<func>\w+)_events
    - (<func>\w+)_events_per_sec
    - (<func>\w+)_nanos_total_per_sec
    - (<func>\w+)_nanos_total
- aurora:scheduled_task
    - penalty_events_per_sec
    - penalty_ms_total_per_sec
    - penalty_events
    - penalty_ms_total
    - penalty_ms_per_event
- aurora:sla
    - platform_uptime_percent (float, percent)
    - mtta_ms (float, ms)
- aurora:sla:cluster
    - platform_uptime_percent
    - mttr_ms
    - mtta_ms
- aurora:system
    - free_physical_memory_mb
    - load_avg
- aurora:task
    - queries_by_id
    - queries_by_host
    - throttle_ms_total_per_sec
    - throttle_events
    - queries_all
    - queries_by_job
    - throttle_ms_total
    - kill_retries
    - exit
- aurora:task:delivery:delay
    - 10_0_percentile
    - 50_0_percentile
    - 90_0_percentile
    - 99_0_percentile
    - 99_99_percentile
    - 99_9_percentile
    - error_rate
    - errors
    - errors_per_sec
    - reconnects
    - requests_events
    - requests_events_per_sec
    - requests_micros_per_event
    - requests_micros_total
    - requests_micros_total_per_sec
    - requests_per_sec
    - timeout_rate
    - timeouts
    - timeouts_per_sec
- aurora:tasks
    - store (integer, count)
    - count (integer, count)
    - rack_lost
    - pending_tasks
    - penaltyMs
- aurora:quota
    - cpu_cores
    - ram_mb
    - disk_mb


### Tags:

- All measurements have the following tags:
    - server    (target providing metrics)
    - host      (host that telegraf is running)
    - acting    (leader|standby)
    - jvm       (version of Java running aurora)
    - version   (version of aurora)
- aurora
    - update_transition
- aurora:cron
    - action
- aurora:resources
    - pool
    - class
- aurora:http:responses
    - code
- aurora:scheduler
    - state
- aurora:scheduler:storage
    - action
- aurora:sla has the following tags:
    - role
    - env
    - job
- aurora:sla:cluster has the following tags:
    - class
    - resource
    - stage
- aurora:task
    - reason
- aurora:task:delivery:delay
    - source
- aurora:tasks
    - state
    - role
    - env
    - job
    - rack
    - reason
- aurora:quota
    - role

### Sample Queries:


Get the max, mean, and min for the task store grouped by state in the last hour:
```
SELECT max(store), mean(store), min(store) FROM aurora:tasks WHERE host=aurora.local AND time > now() - 1h GROUP BY state
```
### Example Output:

```
$ telegraf --input-filter example --test
...
> aurora:jvm,server=127.0.0.1:8081,host=aurora.local memory_max_mb=880 1498665093000000000
> aurora:sla,role=aurora_role,env=prod,job=aurora_task,server=127.0.0.1:8081,host=aurora.local mtta_ms=0 1498665093000000000
> aurora:sla,job=aurora_task,server=127.0.0.1:8081,host=aurora.local,role=aurora_role,env=staging1 mtta_ms=0 1498665093000000000
> aurora:scheduler:task,server=127.0.0.1:8081,host=aurora.local throttle_events_per_sec=0 1498665093000000000
> aurora:sla,server=127.0.0.1:8081,host=aurora.local,role=aurora_role,env=staging1,job=aurora_task platform_uptime_percent=100 1498665093000000000
> aurora:sla,env=prod,job=aurora_task,host=aurora.local,server=127.0.0.1:8081,role=aurora_role mtta_ms=0 1498665093000000000
> aurora:sla,server=127.0.0.1:8081,host=aurora.local,role=aurora_role,env=prod,job=aurora_task platform_uptime_percent=100 1498665093000000000
> aurora:sla,role=aurora_role-services,env=prod,job=aurora_task1assignmentapi-async,server=127.0.0.1:8081,host=aurora.local mtta_ms=0 1498665093000000000
> aurora,host=aurora.local,state=FAILED,server=127.0.0.1:8081 update_transition=0 1498665093000000000
> aurora:sla,host=aurora.local,server=127.0.0.1:8081,role=aurora_role,env=prod,job=aurora_task mtta_ms=0 1498665093000000000
> aurora:sla,job=aurora_task,server=127.0.0.1:8081,host=aurora.local,role=aurora_role,env=prod platform_uptime_percent=100 1498665093000000000
> aurora:tasks,job=aurora_task,server=127.0.0.1:8081,host=aurora.local,state=FAILED,role=aurora_role,env=staging0 count=2 1498665093000000000
> aurora:sla,role=aurora_role,env=staging2,job=aurora_task,host=aurora.local,server=127.0.0.1:8081 platform_uptime_percent=100 1498665093000000000
> aurora:sla,job=aurora_task,server=127.0.0.1:8081,host=aurora.local,role=aurora_role,env=staging1 mtta_ms=0 1498665093000000000
> aurora:sla,server=127.0.0.1:8081,host=aurora.local,role=aurora_role,env=staging1,job=aurora_task mtta_ms=0 1498665093000000000
```