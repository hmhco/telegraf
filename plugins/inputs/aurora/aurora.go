package aurora

import (
	"encoding/json"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	MEASUREMENT = "aurora"
	SEPARATOR   = ":"
)

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{
	Transport: tr,
	Timeout:   time.Duration(4 * time.Second),
}

type Measurement struct {
	fields      []string
	tags        map[string]string
	measurement string
}

type Pendingtasks []struct {
	Name      string   `json:"name"`
	PenaltyMs int64    `json:"penaltyMs"`
	Reason    string   `json:"reason"`
	TaskIDs   []string `json:"taskIds"`


}

type Aurora struct {
	Hosts []string

	acc telegraf.Accumulator

	measurements map[string]Measurement
	sync.WaitGroup
}

type AuroraMetric struct {
	measurement string
	value       string
}

func (c *Aurora) SetDefaults() {
	if len(c.Hosts) == 0 {
		c.Hosts = []string{"http://localhost:8081"}
	}

}

var tagItems = map[string]string{
	"build.version":         "version",
	"jvm_prop_java.version": "jvm",
}
var items = map[string]*AuroraMetric{
	"timed_out_tasks":                                   &AuroraMetric{MEASUREMENT, "timed_out_tasks"},
	"quartz_scheduler_running":                          &AuroraMetric{MEASUREMENT, "quartz_scheduler_running"},
	"uncaught_exceptions":                               &AuroraMetric{MEASUREMENT, "uncaught_exceptions"},
	"gc_executor_tasks_lost":                            &AuroraMetric{MEASUREMENT, "gc_executor_tasks_lost"},
	"job_update_delete_errors":                          &AuroraMetric{MEASUREMENT + SEPARATOR + "job" + SEPARATOR + "update", "delete_errors"},
	"job_update_recovery_errors":                        &AuroraMetric{MEASUREMENT + SEPARATOR + "job" + SEPARATOR + "update", "recovery_errors"},
	"job_update_state_change_errors":                    &AuroraMetric{MEASUREMENT + SEPARATOR + "job" + SEPARATOR + "update", "state_change_errors"},
	"jvm_class_loaded_count":                            &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm", "class_loaded_count"},
	"jvm_class_total_loaded_count":                      &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm", "class_total_loaded_count"},
	"jvm_class_unloaded_count":                          &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm", "class_unloaded_count"},
	"jvm_memory_heap_mb_max":                            &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm", "memory_heap_mb_max"},
	"jvm_memory_max_mb":                                 &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm", "memory_max_mb"},
	"jvm_memory_mb_total":                               &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm", "memory_mb_total"},
	"jvm_memory_non_heap_mb_max":                        &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm", "memory_non_heap_mb_max"},
	"jvm_threads_peak":                                  &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm" + SEPARATOR + "threads", "peak"},
	"jvm_uptime_secs":                                   &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm", "uptime"},
	"jvm_threads_started":                               &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm" + SEPARATOR + "threads", "started"},
	"log_storage_write_lock_wait_events":                &AuroraMetric{MEASUREMENT + SEPARATOR + "log" + SEPARATOR + "storage" + SEPARATOR + "write", "lock_wait_events"},
	"log_storage_write_lock_wait_ns_total":              &AuroraMetric{MEASUREMENT + SEPARATOR + "log" + SEPARATOR + "storage" + SEPARATOR + "write", "lock_wait_ns_total"},
	"offer_accept_races":                                &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler", "offer_accept_races"},
	"offers_rescinded":                                  &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler", "offers_rescinded"},
	"preemptor_missing_attributes":                      &AuroraMetric{MEASUREMENT + SEPARATOR + "preemptor", "missing_attributes"},
	"preemptor_slot_validation_failed":                  &AuroraMetric{MEASUREMENT + SEPARATOR + "preemptor" + SEPARATOR + "slot", "validation_failed"},
	"preemptor_slot_validation_successful":              &AuroraMetric{MEASUREMENT + SEPARATOR + "preemptor" + SEPARATOR + "slot", "validation_successful"},
	"preemptor_task_processor_runs":                     &AuroraMetric{MEASUREMENT + SEPARATOR + "preemptor", "task_processor_runs"},
	"preemptor_tasks_preempted_non_prod":                &AuroraMetric{MEASUREMENT + SEPARATOR + "preemptor", "tasks_preempted_non_prod"},
	"preemptor_tasks_preempted_prod":                    &AuroraMetric{MEASUREMENT + SEPARATOR + "preemptor", "tasks_preempted_prod"},
	"process_max_fd_count":                              &AuroraMetric{MEASUREMENT + SEPARATOR + "process", "max_fd_count"},
	"process_open_fd_count":                             &AuroraMetric{MEASUREMENT + SEPARATOR + "process", "open_fd_count"},
	"schedule_attempts_failed":                          &AuroraMetric{MEASUREMENT + SEPARATOR + "schedule", "attempts_failed"},
	"schedule_attempts_fired":                           &AuroraMetric{MEASUREMENT + SEPARATOR + "schedule", "attempts_fired"},
	"schedule_attempts_no_match":                        &AuroraMetric{MEASUREMENT + SEPARATOR + "schedule", "attempts_no_match"},
	"scheduler_backup_success":                          &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler", "backup_success"},
	"scheduler_backup_failed":                           &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler", "backup_failed"},
	"scheduler_driver_kill_failures":                    &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler", "driver_kill_failures"},
	"framework_registered":                              &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler", "framework_registered"},
	"scheduler_resource_offers":                         &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler", "resource_offers"},
	"scheduler_gc_insufficient_offers":                  &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "gc", "insufficient_offers"},
	"scheduler_gc_offers_consumed":                      &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "gc", "offers_consumed"},
	"scheduler_gc_tasks_created":                        &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "gc", "tasks_created"},
	"scheduler_log_bad_frames_read":                     &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "bad_frames_read"},
	"scheduler_log_bytes_read":                          &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "bytes_read"},
	"scheduler_log_deflated_entries_read":               &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "deflated_entries_read"},
	"scheduler_log_entries_read":                        &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "entries_read"},
	"scheduler_log_entries_written":                     &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "entries_written"},
	"scheduler_log_native_append_events":                &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_append_events"},
	"scheduler_log_native_append_failures":              &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_append_failures"},
	"scheduler_log_native_append_nanos_total":           &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_append_nanos_total"},
	"scheduler_log_native_append_nanos_total_per_sec":   &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_append_nanos_total_per_sec"},
	"scheduler_log_native_append_timeouts":              &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_append_timeouts"},
	"scheduler_log_native_native_entries_skipped":       &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_native_entries_skipped"},
	"scheduler_log_native_read_events":                  &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_read_events"},
	"scheduler_log_native_read_failures":                &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_read_failures"},
	"scheduler_log_native_read_nanos_total":             &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_read_nanos_total"},
	"scheduler_log_native_read_timeouts":                &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_read_timeouts"},
	"scheduler_log_native_truncate_events":              &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_truncate_events"},
	"scheduler_log_native_truncate_failures":            &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_truncate_failures"},
	"scheduler_log_native_truncate_nanos_total":         &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_truncate_nanos_total"},
	"scheduler_log_native_truncate_timeouts":            &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_truncate_timeouts"},
	"scheduler_log_snapshots":                           &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "snapshots"},
	"scheduler_log_un_snapshotted_transactions":         &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "un_snapshotted_transactions"},
	"async_tasks_completed":                             &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "tasks", "async"},
	"scheduled_task_penalty_events":                     &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduled_task", "penalty_events"},
	"scheduled_task_penalty_ms_total":                   &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduled_task", "penalty_ms_total"},
	"task_kill_retries":                                 &AuroraMetric{MEASUREMENT + SEPARATOR + "task", "kill_retries"},
	"task_queries_all":                                  &AuroraMetric{MEASUREMENT + SEPARATOR + "task", "queries_all"},
	"task_queries_by_host":                              &AuroraMetric{MEASUREMENT + SEPARATOR + "task", "queries_by_host"},
	"task_queries_by_id":                                &AuroraMetric{MEASUREMENT + SEPARATOR + "task", "queries_by_id"},
	"task_queries_by_job":                               &AuroraMetric{MEASUREMENT + SEPARATOR + "task", "queries_by_job"},
	"task_throttle_events":                              &AuroraMetric{MEASUREMENT + SEPARATOR + "task", "throttle_events"},
	"task_throttle_ms_total":                            &AuroraMetric{MEASUREMENT + SEPARATOR + "task", "throttle_ms_total"},
	"jvm_gc_PS_MarkSweep_collection_time_ms":            &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm" + SEPARATOR + "gc", "ps_marksweep_collection_time_ms"},
	"jvm_gc_PS_Scavenge_collection_time_ms":             &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm" + SEPARATOR + "gc", "ps_scavenge_collection_time_ms"},
	"jvm_gc_collection_time_ms":                         &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm" + SEPARATOR + "gc", "collection_time_ms"},
	"jvm_gc_PS_MarkSweep_collection_count":              &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm" + SEPARATOR + "gc", "ps_marksweep_collection_count"},
	"jvm_gc_PS_Scavenge_collection_count":               &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm" + SEPARATOR + "gc", "ps_scavenge_collection_count"},
	"jvm_gc_collection_count":                           &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm" + SEPARATOR + "gc", "collection_count"},
	"jvm_memory_free_mb":                                &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm" + SEPARATOR + "memory", "free_mb"},
	"jvm_memory_heap_mb_committed":                      &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm" + SEPARATOR + "memory", "heap_mb_committed"},
	"jvm_memory_heap_mb_used":                           &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm" + SEPARATOR + "memory", "heap_mb_used"},
	"jvm_memory_non_heap_mb_committed":                  &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm", "memory_non_heap_mb_committed"},
	"jvm_memory_non_heap_mb_used":                       &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm" + SEPARATOR + "memory", "non_heap_mb_used"},
	"jvm_threads_active":                                &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm" + SEPARATOR + "threads", "active"},
	"jvm_threads_daemon":                                &AuroraMetric{MEASUREMENT + SEPARATOR + "jvm" + SEPARATOR + "threads", "daemon"},
	"log_storage_write_lock_wait_events_per_sec":        &AuroraMetric{MEASUREMENT + SEPARATOR + "log" + SEPARATOR + "storage" + SEPARATOR + "write", "lock_wait_events_per_sec"},
	"log_storage_write_lock_wait_ns_per_event":          &AuroraMetric{MEASUREMENT + SEPARATOR + "log" + SEPARATOR + "storage" + SEPARATOR + "write", "lock_wait_ns_per_event"},
	"log_storage_write_lock_wait_ns_total_per_sec":      &AuroraMetric{MEASUREMENT + SEPARATOR + "log" + SEPARATOR + "storage" + SEPARATOR + "write", "lock_wait_ns_total_per_sec"},
	"outstanding_offers":                                &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler", "outstanding_offers"},
	"process_cpu_cores_utilized":                        &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler", "process_cpu_cores_utilized"},
	"pubsub_executor_queue_size":                        &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler", "pubsub_executor_queue_size"},
	"schedule_queue_size":                               &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler", "queue_size"},
	"scheduled_task_penalty_events_per_sec":             &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduled_task", "penalty_events_per_sec"},
	"scheduled_task_penalty_ms_per_event":               &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduled_task", "penalty_ms_per_event"},
	"scheduled_task_penalty_ms_total_per_sec":           &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduled_task", "penalty_ms_total_per_sec"},
	"scheduler_log_bytes_written":                       &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "bytes_written"},
	"scheduler_log_native_append_events_per_sec":        &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_append_events_per_sec"},
	"scheduler_log_native_append_nanos_per_event":       &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_append_nanos_per_event"},
	"scheduler_log_native_read_events_per_sec":          &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_read_events_per_sec"},
	"scheduler_log_native_read_nanos_per_event":         &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_read_nanos_per_event"},
	"scheduler_log_native_read_nanos_total_per_sec":     &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_read_nanos_total_per_sec"},
	"scheduler_log_native_truncate_events_per_sec":      &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_truncate_events_per_sec"},
	"scheduler_log_native_truncate_nanos_per_event":     &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_truncate_nanos_per_event"},
	"scheduler_log_native_truncate_nanos_total_per_sec": &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "log", "native_truncate_nanos_total_per_sec"},
	"timeout_queue_size":                                &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler", "timeout_queue_size"},
	"task_throttle_events_per_sec":                      &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "task", "throttle_events_per_sec"},
	"task_throttle_ms_total_per_sec":                    &AuroraMetric{MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "task", "throttle_ms_total_per_sec"},
	"system_free_physical_memory_mb":                    &AuroraMetric{MEASUREMENT + SEPARATOR + "system", "free_physical_memory_mb"},
	"system_load_avg":                                   &AuroraMetric{MEASUREMENT + SEPARATOR + "system", "load_avg"},
	"status_updates_queue_size":                         &AuroraMetric{MEASUREMENT + SEPARATOR + "status" + SEPARATOR + "updates", "queue_size"},
	"slaves_lost":                                       &AuroraMetric{MEASUREMENT, "slaves_lost"},
}

type parser struct {
	measurement string
	field       interface{}
	tags        []string
	regex       *regexp.Regexp
}

func (c *Aurora) parse(key string, value float64, rootTags map[string]string, p *parser) {
	defer c.Done()
	tags := make(map[string]string)
	match := p.regex.FindStringSubmatch(key)
	if len(match)-1 == len(p.tags) {
		field := ""
		for k, v := range rootTags {
			tags[k] = v
		}
		fields := make(map[string]interface{})
		for i, v := range match[1:] {
			tags[p.tags[i]] = v
		}
		if p.field == nil {
			if val, ok := tags["value"]; ok {
				field = val
				delete(tags, "value")
			}
		} else {
			f, ok := p.field.(string)
			if ok {
				field = f
			}
		}
		fields[field] = value
		c.acc.AddFields(p.measurement, fields, tags)
	}
}

var parsed = map[string]*parser{
	"scheduler_thrift_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "thrift",
		field:       nil,
		tags:        []string{"method", "value"},
		regex:       regexp.MustCompile("scheduler_thrift_(?P<method>[0-9a-zA-Z]+)_(?P<value>events|nanos_total|events_per_sec|nanos_per_event|nanos_total_per_sec)"),
	},
	"preemptor_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "preemptor" + SEPARATOR + "slot",
		field:       nil,
		tags:        []string{"stage", "value"},
		regex:       regexp.MustCompile("preemptor_(?P<value>[a-zA-Z0-9]+)_for_(?P<stage>.*)"),
	},
	"tasks_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "tasks",
		field:       "count",
		tags:        []string{"state", "role", "env", "job"},
		regex:       regexp.MustCompile("tasks_(?P<state>.*)_(?P<role>.*)/(?P<env>.*)/(?P<job>.*)"),
	},
	"status_update_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "status" + SEPARATOR + "updates",
		field:       "count",
		tags:        []string{"reason"},
		regex:       regexp.MustCompile("status_update_(?P<reason>.*)$"),
	},
	"tasks_lost_rack_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "tasks",
		field:       "rack_lost",
		tags:        []string{"rack"},
		regex:       regexp.MustCompile("tasks_lost_rack_(?P<rack>.*)"),
	},
	"task_store_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "tasks",
		field:       "store",
		tags:        []string{"state"},
		regex:       regexp.MustCompile("task_store_(?P<state>[A-Z]+)"),
	},
	"update_transition_": &parser{
		measurement: MEASUREMENT,
		field:       "update_transition",
		tags:        []string{"state"},
		regex:       regexp.MustCompile("update_transition_(?P<state>.*)"),
	},
	"scheduler_lifecycle_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "scheduler",
		field:       "status",
		tags:        []string{"state"},
		regex:       regexp.MustCompile("scheduler_lifecycle_(?P<state>[A-Z_]+)"),
	},
	"http_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "http" + SEPARATOR + "responses",
		field:       nil,
		tags:        []string{"code", "value"},
		regex:       regexp.MustCompile("http_(?P<code>[0-9]{3})_responses_(?P<value>events|nanos_total|events_per_sec|nanos_per_event|nanos_total_per_sec)$"),
	},
	"sla_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "sla",
		field:       nil,
		tags:        []string{"role", "env", "job", "value"},
		regex:       regexp.MustCompile("sla_(?P<role>.*)/(?P<env>.*)/(?P<job>.*)_(?P<value>platform_uptime_percent)$"),
	},
	"sla_cluster": &parser{
		measurement: MEASUREMENT + SEPARATOR + "sla" + SEPARATOR + "cluster",
		field:       nil,
		tags:        []string{"value"},
		regex:       regexp.MustCompile("sla_cluster_(?P<value>mtta_ms|_mttr_ms|platform_uptime_percent)$"),
	},
	"sla_cluster_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "sla" + SEPARATOR + "cluster",
		field:       nil,
		tags:        []string{"stage", "value"},
		regex:       regexp.MustCompile("sla_cluster_(?P<value>mtta_ms|mttr_ms|platform_uptime_percent)_(?P<stage>[a-zA-Z0-9]+)"),
	},
	"sla": &parser{
		measurement: MEASUREMENT + SEPARATOR + "sla" + SEPARATOR + "cluster",
		field:       nil,
		tags:        []string{"resource", "class", "value"},
		regex:       regexp.MustCompile("sla_(?P<resource>disk|cpu)_(?P<class>small|medium|large|xlarge)_(?P<value>mtta_ms|mttr_ms|mtts_ms)$"),
	},
	"mem_storage_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "scheduler" + SEPARATOR + "storage",
		field:       nil,
		tags:        []string{"action", "value"},
		regex:       regexp.MustCompile("mem_storage_(?P<action>.*)_(?P<value>events|nanos_total|events_per_sec|nanos_per_event|nanos_total_per_sec)$"),
	},
	"resources_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "resources",
		field:       nil,
		tags:        []string{"state", "value"},
		regex:       regexp.MustCompile("resources_(?P<state>.*)_(?P<value>cpu_cores|disk_mb|ram_mb)$"),
	},
	"empty_slots_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "resources",
		field:       "empty_slots",
		tags:        []string{"pool", "class"},
		regex:       regexp.MustCompile("empty_slots_(?P<pool>.*)_(?P<class>small|medium|large|xlarge)$"),
	},
	"empty_slots": &parser{
		measurement: MEASUREMENT + SEPARATOR + "resources",
		field:       "empty_slots",
		tags:        []string{"class"},
		regex:       regexp.MustCompile("empty_slots_(?P<class>small|medium|large|xlarge)$"),
	},
	"cron_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "cron",
		field:       nil,
		tags:        []string{"value"},
		regex:       regexp.MustCompile("cron_job_(?P<value>.*)$"),
	},
	"task_delivery_delay_SOURCE_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "task" + SEPARATOR + "delivery" + SEPARATOR + "delay",
		field:       nil,
		tags:        []string{"source", "value"},
		regex:       regexp.MustCompile("task_delivery_delay_SOURCE_(?P<source>[A-Z_]+)_(?P<value>.*)$"),
	},
	"task_exit_REASON_": &parser{
		measurement: MEASUREMENT + SEPARATOR + "task" + SEPARATOR + "exit",
		field:       "exit",
		tags:        []string{"reason"},
		regex:       regexp.MustCompile("task_exit_REASON_(?P<reason>[A-Z_]+)$"),
	},
}

func (c *Aurora) fetchMetrics(address, uri string) ([]byte, error) {
	resp, err := client.Get(address + uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}
	return data, nil

}
func (c *Aurora) fetchLeaderStatus(address string) (string, error) {
	resp, err := client.Get(address + "/leaderhealth")
	if err != nil {
		return "", err
	}
	if resp.StatusCode == 200 {
		return "leader", nil
	}
	return "standby", nil

}


func getIdentity(name string) map[string]string {
	re := regexp.MustCompile("(?P<role>.*)/(?P<env>.*)/(?P<job>.*)")
	tags := []string{"role", "env", "job"}
	identity := make(map[string]string)
	match := re.FindStringSubmatch(name)
	if len(match) == 4 {

		for i, v := range match[1:] {
			identity[tags[i]] = v
		}
	}
	return identity
}

func (c *Aurora) fetchPendingTasks(address string, rootTags map[string]string) {
	var pendingTasks Pendingtasks
	tags := make(map[string]string)
	raw, err := c.fetchMetrics(address, "/pendingtasks")
	if len(raw) == 0 {
	}
	if err != nil {
		//c.acc.AddError(err)
	}
	if err = json.Unmarshal(raw, &pendingTasks); err != nil {
		//c.acc.AddError(err)
	}

	fields := make(map[string]interface{})
	for _, p := range pendingTasks {
		for k, v := range rootTags {
			tags[k] = v
		}

		ident := getIdentity(p.Name)
		for k, v := range rootTags {
			ident[k] = v
		}
		ident["reason"] = p.Reason
		fields["pending_tasks"] = float64(len(p.TaskIDs))
		fields["penaltyMs"] = float64(p.PenaltyMs)
		c.acc.AddFields(MEASUREMENT+SEPARATOR+"tasks", fields, ident)
	}

}

func (c *Aurora) fetchQuotas(address string, rootTags map[string]string) {
	quotas := make(map[string]map[string]interface{})
	tags := make(map[string]string)
	raw, err := c.fetchMetrics(address, "/quotas")
	if len(raw) == 0 {
	}
	if err != nil {
		//c.acc.AddError(err)
	}
	if err = json.Unmarshal(raw, &quotas); err != nil {
		//c.acc.AddError(err)
	}

	for k, v := range rootTags {
		tags[k] = v
	}
	for i, q := range quotas {
		tags["role"] = i
		c.acc.AddFields(MEASUREMENT+SEPARATOR+"quota", q, tags)
	}
}

func (c *Aurora) parseAndLabel(name string, value float64, tags map[string]string) {
	for prefix, parser := range parsed {
		if strings.HasPrefix(name, prefix) {
			c.Add(1)
			c.parse(name, value, tags, parser)
		}
	}
}
func (c *Aurora) Description() string {
	return "Reads metrics from one or more Apache Aurora framework hosts."
}

func (c *Aurora) SampleConfig() string {
	return `

  ## aurora servers
  # hosts = ["http://localhost:8081"]
  # hosts = ["https://localhost:8081"]
  `
}
func (c *Aurora) GetTags(vars map[string]interface{}, address string) map[string]string {
	tags := make(map[string]string)
	u, err := url.Parse(address)
	if err != nil {
		//c.acc.AddError(err)
	}
	tags["acting"], err = c.fetchLeaderStatus(u.Scheme + "://" + u.Host)
	if err != nil {
		//c.acc.AddError(err)
	}

	for name, v := range vars {
		v, ok := v.(string)
		if !ok {
			continue
		}
		tags[tagItems[name]] = v
	}
	return tags
}
func (c *Aurora) GatherMetricsByHost(address string) (map[string]Measurement, error) {
	vars := make(map[string]interface{})
	defer c.Done()
	raw, err := c.fetchMetrics(address, "/vars.json")
	if len(raw) == 0 {
	}
	if err != nil {
		c.acc.AddError(err)
	}
	if err = json.Unmarshal(raw, &vars); err != nil {
		c.acc.AddError(err)
	}
	tags := c.GetTags(vars, address)
	c.fetchQuotas(address, tags)
	c.fetchPendingTasks(address, tags)
	for name, v := range vars {
		v, ok := v.(float64)
		if !ok {
			continue
		}
		if measurement, ok := items[name]; ok {
			c.acc.AddFields(measurement.measurement, map[string]interface{}{measurement.value: v}, tags)
		}

		c.parseAndLabel(name, v, tags)
	}

	return c.measurements, nil

}

func (c *Aurora) Gather(acc telegraf.Accumulator) error {
	c.acc = acc
	c.SetDefaults()
	for i := range c.Hosts {
		c.Add(1)
		go c.GatherMetricsByHost(c.Hosts[i])
	}
	c.Wait()
	return nil
}

func init() {
	inputs.Add(MEASUREMENT, func() telegraf.Input { return &Aurora{} })
}
