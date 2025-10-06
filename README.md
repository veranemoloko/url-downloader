# url-downloader
URL Downloader is a web service for asynchronous file downloading from the internet. Users can create tasks containing a list of URLs, and the service will download the files into a local folder. The service supports resilience: tasks in progress are saved and resumed after a restart.

### Features
- Create download tasks with multiple URLs.
- Asynchronous processing with a worker pool.
- Retrieve task status and download progress.
- Recover incomplete tasks after service restart.
- Prometheus metrics for monitoring.

### Architecture
- The service is implemented in Go using common patterns:
- HTTP API via chi for creating and querying tasks.
- Service layer (TaskService, DownloadService) for task and download logic.
- Task storage — JSON file to persist state across restarts.
- Worker pool — limits the number of concurrent downloads.
- Prometheus metrics — number of tasks, successful/failed downloads, download duration and size.
- Graceful shutdown — proper termination on SIGINT/SIGTERM.

### Restart / Recovery Scenario
1. Before shutdown, all tasks are persisted to the state file (state.json).
2. On startup, the service:
    - Loads tasks with Pending and InProgress status.
    - Resumes downloads for tasks that were in progress.
3. New tasks can be submitted at any time, even during recovery.

