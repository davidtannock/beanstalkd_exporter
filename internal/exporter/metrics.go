package exporter

var descSystemMetrics = map[string]struct {
	stat string
	help string
}{
	"cmd_bury_total":                 {stat: "cmd-bury", help: "The cumulative number of bury commands."},
	"cmd_delete_total":               {stat: "cmd-delete", help: "The cumulative number of delete commands."},
	"cmd_ignore_total":               {stat: "cmd-ignore", help: "The cumulative number of ignore commands."},
	"cmd_kick_total":                 {stat: "cmd-kick", help: "The cumulative number of kick commands."},
	"cmd_list_tube_used_total":       {stat: "cmd-list-tube-used", help: "The cumulative number of list-tube-used commands."},
	"cmd_list_tubes_total":           {stat: "cmd-list-tubes", help: "The cumulative number of list-tubes commands."},
	"cmd_list_tubes_watched_total":   {stat: "cmd-list-tubes-watched", help: "The cumulative number of list-tubes-watched."},
	"cmd_pause_tube_total":           {stat: "cmd-pause-tube", help: "The cumulative number of pause-tube commands."},
	"cmd_peek_total":                 {stat: "cmd-peek", help: "The cumulative number of peek commands."},
	"cmd_peek_buried_total":          {stat: "cmd-peek-buried", help: "The cumulative number of peek-buried commands."},
	"cmd_peek_delayed_total":         {stat: "cmd-peek-delayed", help: "The cumulative number of peek-delayed commands."},
	"cmd_peek_ready_total":           {stat: "cmd-peek-ready", help: "The cumulative number of peek-ready commands."},
	"cmd_put_total":                  {stat: "cmd-put", help: "The cumulative number of put commands."},
	"cmd_release_total":              {stat: "cmd-release", help: "The cumulative number of release commands."},
	"cmd_reserve_total":              {stat: "cmd-reserve", help: "The cumulative number of reserve commands."},
	"cmd_reserve_with_timeout_total": {stat: "cmd-reserve-with-timeout", help: "The cumulative number of reserve with a timeout commands."},
	"cmd_stats_total":                {stat: "cmd-stats", help: "The cumulative number of stats commands."},
	"cmd_stats_job_total":            {stat: "cmd-stats-job", help: "The cumulative number of stats-job commands."},
	"cmd_stats_tube_total":           {stat: "cmd-stats-tube", help: "The cumulative number of stats-tube commands."},
	"cmd_touch_total":                {stat: "cmd-touch", help: "The cumulative number of touch commands."},
	"cmd_use_total":                  {stat: "cmd-use", help: "The cumulative number of use commands."},
	"cmd_watch_total":                {stat: "cmd-watch", help: "The cumulative number of watch commands."},
	"current_connections_count":      {stat: "current-connections", help: "The number of currently open connections."},
	"current_jobs_buried_count":      {stat: "current-jobs-buried", help: "The number of buried jobs."},
	"current_jobs_delayed_count":     {stat: "current-jobs-delayed", help: "The number of delayed jobs."},
	"current_jobs_ready_count":       {stat: "current-jobs-ready", help: "The number of jobs in the ready queue."},
	"current_jobs_reserved_count":    {stat: "current-jobs-reserved", help: "The number of jobs reserved by all clients."},
	"current_jobs_urgent_count":      {stat: "current-jobs-urgent", help: "The number of ready jobs with priority < 1024."},
	"current_producers_count":        {stat: "current-producers", help: "The number of open connections that have each issued at least one put command."},
	"current_tubes_count":            {stat: "current-tubes", help: "The number of currently-existing tubes."},
	"current_waiting_count":          {stat: "current-waiting", help: "The number of open connections that have issued a reserve command but not yet received a response."},
	"current_workers_count":          {stat: "current-workers", help: "The number of open connections that have each issued at least one reserve command."},
	"job_timeouts_count":             {stat: "job-timeouts", help: "The cumulative count of times a job has timed out."},
	"total_connections_count":        {stat: "total-connections", help: "The cumulative count of connections."},
	"total_jobs_count":               {stat: "total-jobs", help: "The cumulative count of jobs created in the current beanstalkd process."},
}

var descTubeMetrics = map[string]struct {
	stat string
	help string
}{
	"tube_cmd_delete_total":              {stat: "cmd-delete", help: "The cumulative number of delete commands for this tube."},
	"tube_cmd_pause_tube_total":          {stat: "cmd-pause-tube", help: "The cumulative number of pause-tube commands for this tube."},
	"tube_current_jobs_buried_count":     {stat: "current-jobs-buried", help: "The number of buried jobs for this tube."},
	"tube_current_jobs_delayed_count":    {stat: "current-jobs-delayed", help: "The number of delayed jobs for this tube."},
	"tube_current_jobs_ready_count":      {stat: "current-jobs-ready", help: "The number of jobs in the ready queue for this tube."},
	"tube_current_jobs_reserved_count":   {stat: "current-jobs-reserved", help: "The number of jobs reserved by all clients for this tube."},
	"tube_current_jobs_urgent_count":     {stat: "current-jobs-urgent", help: "The number of ready jobs with priority < 1024 for this tube."},
	"tube_current_using_count":           {stat: "current-using", help: "The number of open connections that are currently using this tube."},
	"tube_current_waiting_count":         {stat: "current-waiting", help: "The number of open connections that have issued a reserve command for this tube but not yet received a response."},
	"tube_current_watching_count":        {stat: "current-watching", help: "The number of open connections that are currently watching this tube."},
	"tube_pause_seconds_total":           {stat: "pause", help: "The number of seconds this tube has been paused for."},
	"tube_pause_time_left_seconds_total": {stat: "pause-time-left", help: "The number of seconds until this tube is un-paused"},
	"tube_total_jobs_count":              {stat: "total-jobs", help: "The cumulative count of jobs created for this tube in the current beanstalkd process."},
}
